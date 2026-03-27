package services

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"shalby_backend/models"

	"github.com/google/uuid"
)

// GetAllMedicines retrieves all medicines with their inventory information
func GetAllMedicines(db *sql.DB) (*models.MedicineListResponse, error) {
	query := `
	SELECT 
		m.id, m.name, m.generic_name, m.category, m.dosage_form, 
		m.manufacturer, m.price, m.is_active, m.created_at,
		COALESCE(mi.stock_quantity, 0), COALESCE(mi.unit, ''), COALESCE(mi.reorder_level, 0), 
		mi.expiry_date, mi.last_restocked_at, mi.updated_at
	FROM medicines m
	LEFT JOIN medicine_inventory mi ON m.id = mi.medicine_id
	ORDER BY m.name ASC
	`

	rows, err := db.Query(query)
	if err != nil {
		log.Println("Error querying medicines:", err)
		return nil, err
	}
	defer rows.Close()

	var medicines []models.MedicineWithStock
	for rows.Next() {
		var med models.MedicineWithStock
		var expDate, lastRestocked sql.NullString
		var invUpdatedAt sql.NullTime

		err := rows.Scan(
			&med.ID, &med.Name, &med.GenericName, &med.Category, &med.DosageForm,
			&med.Manufacturer, &med.Price, &med.IsActive, &med.CreatedAt,
			&med.StockQuantity, &med.Unit, &med.ReorderLevel,
			&expDate, &lastRestocked, &invUpdatedAt,
		)
		if err != nil {
			log.Println("Error scanning medicine:", err)
			continue
		}

		if expDate.Valid {
			med.ExpiryDate = expDate.String
		}
		if lastRestocked.Valid {
			med.LastRestockedAt = lastRestocked.String
		}
		if invUpdatedAt.Valid {
			med.UpdatedAt = invUpdatedAt.Time
		}

		medicines = append(medicines, med)
	}

	return &models.MedicineListResponse{
		Medicines: medicines,
		Total:     len(medicines),
	}, nil
}

// GetMedicineStats returns medicine inventory statistics
func GetMedicineStats(db *sql.DB) (*models.MedicineStatsResponse, error) {
	stats := &models.MedicineStatsResponse{}

	// Total medicines
	err := db.QueryRow("SELECT COUNT(*) FROM medicines WHERE is_active = true").Scan(&stats.TotalMedicines)
	if err != nil && err != sql.ErrNoRows {
		log.Println("Error counting medicines:", err)
		return nil, err
	}

	// Low stock count
	err = db.QueryRow(`
		SELECT COUNT(*) FROM medicine_inventory 
		WHERE stock_quantity <= reorder_level
	`).Scan(&stats.LowStockCount)
	if err != nil && err != sql.ErrNoRows {
		log.Println("Error counting low stock:", err)
		return nil, err
	}

	// Total inventory value
	err = db.QueryRow(`
		SELECT COALESCE(SUM(mi.stock_quantity * m.price), 0) 
		FROM medicine_inventory mi
		JOIN medicines m ON mi.medicine_id = m.id
	`).Scan(&stats.TotalValue)
	if err != nil && err != sql.ErrNoRows {
		log.Println("Error calculating total value:", err)
		return nil, err
	}

	// Expiring count (next 30 days)
	err = db.QueryRow(`
		SELECT COUNT(*) FROM medicine_inventory 
		WHERE expiry_date <= NOW() + INTERVAL '30 days' AND expiry_date > NOW()
	`).Scan(&stats.ExpiringCount)
	if err != nil && err != sql.ErrNoRows {
		log.Println("Error counting expiring:", err)
		return nil, err
	}

	// Discontinued count
	err = db.QueryRow("SELECT COUNT(*) FROM medicines WHERE is_active = false").Scan(&stats.DiscontinuedCount)
	if err != nil && err != sql.ErrNoRows {
		log.Println("Error counting discontinued:", err)
		return nil, err
	}

	return stats, nil
}

// CreateMedicine creates a new medicine and its inventory entry
func CreateMedicine(db *sql.DB, med models.MedicineRequest, inv models.MedicineInventoryRequest) (*models.MedicineResponse, error) {
	id := uuid.New().String()

	tx, err := db.Begin()
	if err != nil {
		log.Println("Error starting transaction:", err)
		return nil, fmt.Errorf("transaction error")
	}
	defer tx.Rollback()

	// Create medicine
	query := `
	INSERT INTO medicines (id, name, generic_name, category, dosage_form, manufacturer, price, is_active, created_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, true, NOW())
	`

	_, err = tx.Exec(query,
		id, med.Name, med.GenericName, med.Category, med.DosageForm,
		med.Manufacturer, med.Price,
	)

	if err != nil {
		log.Println("Error creating medicine:", err)
		return nil, fmt.Errorf("failed to create medicine")
	}

	// Create inventory entry
	invQuery := `
	INSERT INTO medicine_inventory (medicine_id, stock_quantity, unit, reorder_level, expiry_date, updated_at)
	VALUES ($1, $2, $3, $4, NULLIF($5, '')::date, NOW())
	`

	_, err = tx.Exec(invQuery,
		id, inv.StockQuantity, inv.Unit, inv.ReorderLevel, inv.ExpiryDate,
	)

	if err != nil {
		log.Println("Error creating inventory:", err)
		return nil, fmt.Errorf("failed to create inventory")
	}

	if err = tx.Commit(); err != nil {
		log.Println("Error committing transaction:", err)
		return nil, err
	}

	return &models.MedicineResponse{
		ID:           id,
		Name:         med.Name,
		GenericName:  med.GenericName,
		Category:     med.Category,
		DosageForm:   med.DosageForm,
		Manufacturer: med.Manufacturer,
		Price:        med.Price,
		IsActive:     true,
		CreatedAt:    time.Now(),
	}, nil
}

// UpdateMedicine updates medicine information
func UpdateMedicine(db *sql.DB, medicineID string, req models.MedicineRequest) (*models.MedicineResponse, error) {
	query := `
	UPDATE medicines 
	SET name = $1, generic_name = $2, category = $3, dosage_form = $4, 
	    manufacturer = $5, price = $6
	WHERE id = $7
	`

	_, err := db.Exec(query,
		req.Name, req.GenericName, req.Category, req.DosageForm,
		req.Manufacturer, req.Price, medicineID,
	)

	if err != nil {
		log.Println("Error updating medicine:", err)
		return nil, fmt.Errorf("failed to update medicine")
	}

	return &models.MedicineResponse{
		ID:           medicineID,
		Name:         req.Name,
		GenericName:  req.GenericName,
		Category:     req.Category,
		DosageForm:   req.DosageForm,
		Manufacturer: req.Manufacturer,
		Price:        req.Price,
		IsActive:     req.IsActive,
		CreatedAt:    time.Now(),
	}, nil
}

// UpdateMedicineInventory updates stock and expiry information
func UpdateMedicineInventory(db *sql.DB, medicineID string, req models.MedicineInventoryRequest) error {
	query := `
	UPDATE medicine_inventory 
	SET stock_quantity = $1, unit = $2, reorder_level = $3, expiry_date = $4::date, updated_at = NOW()
	WHERE medicine_id = $5
	`

	result, err := db.Exec(query,
		req.StockQuantity, req.Unit, req.ReorderLevel, req.ExpiryDate, medicineID,
	)

	if err != nil {
		log.Println("Error updating inventory:", err)
		return fmt.Errorf("failed to update inventory")
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected")
	}

	if rows == 0 {
		// If no inventory record exists, create one
		createQuery := `
		INSERT INTO medicine_inventory (medicine_id, stock_quantity, unit, reorder_level, expiry_date, updated_at)
		VALUES ($1, $2, $3, $4, $5::date, NOW())
		`
		_, err := db.Exec(createQuery,
			medicineID, req.StockQuantity, req.Unit, req.ReorderLevel, req.ExpiryDate,
		)
		if err != nil {
			log.Println("Error creating inventory record:", err)
			return fmt.Errorf("failed to create inventory record")
		}
	}

	return nil
}

// DeleteMedicine soft deletes a medicine by marking as inactive
func DeleteMedicine(db *sql.DB, medicineID string) error {
	result, err := db.Exec("UPDATE medicines SET is_active = false WHERE id = $1", medicineID)
	if err != nil {
		log.Println("Error deleting medicine:", err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("medicine not found")
	}

	return nil
}
