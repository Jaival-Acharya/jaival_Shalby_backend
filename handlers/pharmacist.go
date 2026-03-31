package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"time"

	"shalby_backend/models"

	"github.com/gin-gonic/gin"
)

// GetPharmacistProfile returns pharmacist's own profile
func GetPharmacistProfile(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		pharmacistID := c.GetString("user_id") // From auth context

		var id, name, email string
		var avatarUrl, phone, address sql.NullString
		var createdAt time.Time

		err := db.QueryRow(`
			SELECT u.id, u.name, u.email, u.avatar_url, u.phone, u.address, u.created_at
			FROM users u
			WHERE u.id = $1
		`, pharmacistID).Scan(&id, &name, &email, &avatarUrl, &phone, &address, &createdAt)

		if err != nil {
			log.Println("Error fetching profile:", err)
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: "User not found",
			})
			return
		}

		profile := gin.H{
			"id":        id,
			"name":      name,
			"email":     email,
			"createdAt": createdAt,
		}

		if avatarUrl.Valid {
			profile["avatarUrl"] = avatarUrl.String
		}
		if phone.Valid {
			profile["phone"] = phone.String
		}
		if address.Valid {
			profile["address"] = address.String
		}

		c.JSON(http.StatusOK, models.SuccessResponse{
			Message: "Profile fetched successfully",
			Data:    profile,
		})
	}
}

// GetPharmacyDashboardStats returns pharmacy dashboard statistics
func GetPharmacyDashboardStats(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		stats := gin.H{}

		// Pending prescriptions
		var pending int
		db.QueryRow(`SELECT COUNT(*) FROM prescriptions WHERE pharmacy_status = 'Pending'`).Scan(&pending)
		stats["pending_prescriptions"] = pending

		// Low stock medicines
		var lowStock int
		db.QueryRow(`
			SELECT COUNT(*) FROM medicine_inventory 
			WHERE stock_quantity <= reorder_level
		`).Scan(&lowStock)
		stats["low_stock_count"] = lowStock

		// Total inventory value
		var totalValue float64
		db.QueryRow(`
			SELECT COALESCE(SUM(mi.stock_quantity * m.price), 0)
			FROM medicine_inventory mi
			JOIN medicines m ON mi.medicine_id = m.id
		`).Scan(&totalValue)
		stats["total_inventory_value"] = totalValue

		// Dispensed today
		var dispensedToday int
		db.QueryRow(`
			SELECT COUNT(*) FROM prescriptions 
			WHERE pharmacy_status = 'Dispensed' AND DATE(dispensed_at) = CURRENT_DATE
		`).Scan(&dispensedToday)
		stats["dispensed_today"] = dispensedToday

		c.JSON(http.StatusOK, models.SuccessResponse{
			Message: "Dashboard stats fetched successfully",
			Data:    stats,
		})
	}
}

// GetPharmacyMedicines returns medicines with inventory info
func GetPharmacyMedicines(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		limit := 20
		offset := 0

		if l := c.Query("limit"); l != "" {
			if parsed, err := strconv.Atoi(l); err == nil {
				limit = parsed
			}
		}

		if o := c.Query("offset"); o != "" {
			if parsed, err := strconv.Atoi(o); err == nil {
				offset = parsed
			}
		}

		rows, err := db.Query(`
			SELECT m.id, m.name, m.generic_name, m.category, m.dosage_form, m.price,
			       COALESCE(mi.stock_quantity, 0), COALESCE(mi.unit, ''), 
			       COALESCE(mi.reorder_level, 0), mi.expiry_date
			FROM medicines m
			LEFT JOIN medicine_inventory mi ON m.id = mi.medicine_id
			WHERE m.is_active = true
			ORDER BY m.name
			LIMIT $1 OFFSET $2
		`, limit, offset)
		if err != nil {
			log.Println("Error querying medicines:", err)
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: "Failed to fetch medicines",
			})
			return
		}
		defer rows.Close()

		var medicines []gin.H
		for rows.Next() {
			var id, name, genericName, category, dosageForm string
			var price float64
			var stock, reorderLevel int
			var unit string
			var expiry sql.NullString

			if err := rows.Scan(&id, &name, &genericName, &category, &dosageForm, &price,
				&stock, &unit, &reorderLevel, &expiry); err != nil {
				log.Println("Error scanning medicine:", err)
				continue
			}

			med := gin.H{
				"id":            id,
				"name":          name,
				"genericName":   genericName,
				"category":      category,
				"dosageForm":    dosageForm,
				"price":         price,
				"stockQuantity": stock,
				"unit":          unit,
				"reorderLevel":  reorderLevel,
			}

			if expiry.Valid {
				med["expiryDate"] = expiry.String
			}

			medicines = append(medicines, med)
		}

		c.JSON(http.StatusOK, models.SuccessResponse{
			Message: "Medicines fetched successfully",
			Data:    medicines,
		})
	}
}

// GetLowStockMedicines returns medicines below reorder level
func GetLowStockMedicines(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := db.Query(`
			SELECT m.id, m.name, COALESCE(mi.stock_quantity, 0), mi.reorder_level
			FROM medicines m
			LEFT JOIN medicine_inventory mi ON m.id = mi.medicine_id
			WHERE m.is_active = true AND COALESCE(mi.stock_quantity, 0) <= COALESCE(mi.reorder_level, 50)
			ORDER BY mi.stock_quantity ASC
		`)
		if err != nil {
			log.Println("Error querying medicines:", err)
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: "Failed to fetch low stock medicines",
			})
			return
		}
		defer rows.Close()

		var medicines []gin.H
		for rows.Next() {
			var id, name string
			var stock int
			var reorderLevel sql.NullInt64

			if err := rows.Scan(&id, &name, &stock, &reorderLevel); err != nil {
				log.Println("Error scanning medicine:", err)
				continue
			}

			med := gin.H{
				"id":            id,
				"name":          name,
				"stockQuantity": stock,
			}

			if reorderLevel.Valid {
				med["reorderLevel"] = reorderLevel.Int64
			}

			medicines = append(medicines, med)
		}

		c.JSON(http.StatusOK, models.SuccessResponse{
			Message: "Low stock medicines fetched successfully",
			Data:    medicines,
		})
	}
}

// GetExpiringMedicines returns medicines expiring soon
func GetExpiringMedicines(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		daysUntilExpiry := 30
		if d := c.Query("days"); d != "" {
			if parsed, err := strconv.Atoi(d); err == nil {
				daysUntilExpiry = parsed
			}
		}

		rows, err := db.Query(`
			SELECT m.id, m.name, mi.expiry_date, COALESCE(mi.stock_quantity, 0)
			FROM medicines m
			JOIN medicine_inventory mi ON m.id = mi.medicine_id
			WHERE m.is_active = true 
			AND mi.expiry_date <= NOW() + INTERVAL '1 day' * $1
			AND mi.expiry_date > NOW()
			ORDER BY mi.expiry_date ASC
		`, daysUntilExpiry)
		if err != nil {
			log.Println("Error querying medicines:", err)
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: "Failed to fetch expiring medicines",
			})
			return
		}
		defer rows.Close()

		var medicines []gin.H
		for rows.Next() {
			var id, name string
			var stock int
			var expiry time.Time

			if err := rows.Scan(&id, &name, &expiry, &stock); err != nil {
				log.Println("Error scanning medicine:", err)
				continue
			}

			med := gin.H{
				"id":            id,
				"name":          name,
				"expiryDate":    expiry.Format("2006-01-02"),
				"stockQuantity": stock,
			}

			medicines = append(medicines, med)
		}

		c.JSON(http.StatusOK, models.SuccessResponse{
			Message: "Expiring medicines fetched successfully",
			Data:    medicines,
		})
	}
}
