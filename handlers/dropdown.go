package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// DropdownResponse is a generic response for dropdown options
type DropdownOption struct {
	ID    interface{} `json:"id"`
	Name  string      `json:"name"`
	Extra interface{} `json:"extra,omitempty"`
}

// GetDepartments returns all active departments
// GET /api/dropdowns/departments
func GetDepartments(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		query := `
			SELECT id, name, description, type, is_active 
			FROM departments 
			WHERE is_active = true 
			ORDER BY name ASC
		`

		rows, err := db.Query(query)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch departments"})
			return
		}
		defer rows.Close()

		var departments []map[string]interface{}
		for rows.Next() {
			var id, name, departmentType string
			var description sql.NullString
			var isActive bool
			if err := rows.Scan(&id, &name, &description, &departmentType, &isActive); err != nil {
				continue
			}
			departments = append(departments, map[string]interface{}{
				"id":          id,
				"name":        name,
				"description": description.String,
				"type":        departmentType,
				"is_active":   isActive,
			})
		}

		c.JSON(http.StatusOK, departments)
	}
}

// GetSpecializations returns specializations for a specific department
// GET /api/dropdowns/specializations?department_id={id}
func GetSpecializations(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		departmentID := c.Query("department_id")
		if departmentID == "" {
			// Return all specializations if no department specified
			query := `
				SELECT id, name, department_id, is_active 
				FROM specializations 
				WHERE is_active = true 
				ORDER BY name ASC
			`

			rows, err := db.Query(query)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch specializations"})
				return
			}
			defer rows.Close()

			var specs []map[string]interface{}
			for rows.Next() {
				var id, name, deptID string
				var isActive bool
				if err := rows.Scan(&id, &name, &deptID, &isActive); err != nil {
					continue
				}
				specs = append(specs, map[string]interface{}{
					"id":            id,
					"name":          name,
					"department_id": deptID,
				})
			}

			c.JSON(http.StatusOK, specs)
			return
		}

		// Return specializations for specific department
		query := `
			SELECT id, name, department_id 
			FROM specializations 
			WHERE department_id = $1 AND is_active = true 
			ORDER BY name ASC
		`

		rows, err := db.Query(query, departmentID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch specializations"})
			return
		}
		defer rows.Close()

		var specs []map[string]interface{}
		for rows.Next() {
			var id, name, deptID string
			if err := rows.Scan(&id, &name, &deptID); err != nil {
				continue
			}
			specs = append(specs, map[string]interface{}{
				"id":            id,
				"name":          name,
				"department_id": deptID,
			})
		}

		c.JSON(http.StatusOK, specs)
	}
}

// GetAllergies returns all allergies from allergy_master
// GET /api/dropdowns/allergies
func GetAllergies(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		query := `
			SELECT id, name 
			FROM allergy_master 
			ORDER BY name ASC
		`

		rows, err := db.Query(query)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch allergies"})
			return
		}
		defer rows.Close()

		var allergies []map[string]interface{}
		for rows.Next() {
			var id string
			var name string
			if err := rows.Scan(&id, &name); err != nil {
				continue
			}
			allergies = append(allergies, map[string]interface{}{
				"id":   id,
				"name": name,
			})
		}

		c.JSON(http.StatusOK, allergies)
	}
}

// GetConditions returns all chronic conditions from condition_master
// GET /api/dropdowns/conditions
func GetConditions(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		query := `
			SELECT id, name 
			FROM condition_master 
			ORDER BY name ASC
		`

		rows, err := db.Query(query)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch conditions"})
			return
		}
		defer rows.Close()

		var conditions []map[string]interface{}
		for rows.Next() {
			var id string
			var name string
			if err := rows.Scan(&id, &name); err != nil {
				continue
			}
			conditions = append(conditions, map[string]interface{}{
				"id":   id,
				"name": name,
			})
		}

		c.JSON(http.StatusOK, conditions)
	}
}

// GetMedicineCategories returns all medicine categories
// GET /api/dropdowns/medicine-categories
func GetMedicineCategories(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		query := `
			SELECT id, name, description, color_tag 
			FROM medicine_categories 
			ORDER BY name ASC
		`

		rows, err := db.Query(query)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch medicine categories"})
			return
		}
		defer rows.Close()

		var categories []map[string]interface{}
		for rows.Next() {
			var id, name string
			var description, colorTag sql.NullString
			if err := rows.Scan(&id, &name, &description, &colorTag); err != nil {
				continue
			}
			categories = append(categories, map[string]interface{}{
				"id":          id,
				"name":        name,
				"description": description.String,
				"color_tag":   colorTag.String,
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"data": categories,
		})
	}
}

// CreateMedicineCategoryRequest holds the request data for creating a medicine category
type CreateMedicineCategoryRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	ColorTag    string `json:"color_tag"`
}

// CreateMedicineCategory creates a new medicine category
// POST /api/dropdowns/medicine-categories
func CreateMedicineCategory(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateMedicineCategoryRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format: " + err.Error()})
			return
		}

		// Generate UUID
		categoryID := uuid.New().String()

		// Insert category
		query := `
			INSERT INTO medicine_categories (id, name, description, color_tag)
			VALUES ($1, $2, $3, $4)
		`

		_, err := db.Exec(query, categoryID, req.Name, req.Description, req.ColorTag)
		if err != nil {
			// Return error with proper status code
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to create medicine category: " + err.Error(),
			})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"data": gin.H{
				"id":          categoryID,
				"name":        req.Name,
				"description": req.Description,
				"color_tag":   req.ColorTag,
			},
			"message": "Category created successfully",
		})
	}
}

// GetMedicineGenericNames returns distinct generic names filtered by category
// GET /api/dropdowns/medicine-generic-names?category_id={id}
func GetMedicineGenericNames(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		categoryID := c.Query("category_id")

		var query string
		var rows *sql.Rows
		var err error

		if categoryID == "" {
			// Return all distinct generic names
			query = `
				SELECT DISTINCT 
					generic_name 
				FROM medicines 
				WHERE is_active = true 
				ORDER BY generic_name ASC
			`
			rows, err = db.Query(query)
		} else {
			// Return generic names for specific category
			query = `
				SELECT DISTINCT 
					id, generic_name 
				FROM medicines 
				WHERE category = $1 AND is_active = true 
				ORDER BY generic_name ASC
			`
			rows, err = db.Query(query, categoryID)
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch medicine generic names"})
			return
		}
		defer rows.Close()

		var generics []map[string]interface{}
		if categoryID == "" {
			for rows.Next() {
				var name string
				if err := rows.Scan(&name); err != nil {
					continue
				}
				generics = append(generics, map[string]interface{}{
					"name": name,
				})
			}
		} else {
			for rows.Next() {
				var id, name string
				if err := rows.Scan(&id, &name); err != nil {
					continue
				}
				generics = append(generics, map[string]interface{}{
					"id":   id,
					"name": name,
				})
			}
		}

		c.JSON(http.StatusOK, generics)
	}
}

// GetCities returns all cities with state and country
// GET /api/dropdowns/cities
func GetCities(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		query := `
			SELECT id, city, state, country 
			FROM cities 
			ORDER BY state ASC, city ASC
		`

		rows, err := db.Query(query)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch cities"})
			return
		}
		defer rows.Close()

		var cities []map[string]interface{}
		for rows.Next() {
			var id int
			var city, state, country string
			if err := rows.Scan(&id, &city, &state, &country); err != nil {
				continue
			}
			cities = append(cities, map[string]interface{}{
				"id":      id,
				"city":    city,
				"state":   state,
				"country": country,
			})
		}

		c.JSON(http.StatusOK, cities)
	}
}

// GetRoles returns all active roles
// GET /api/dropdowns/roles
func GetRoles(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		query := `
			SELECT id, name 
			FROM roles 
			ORDER BY name ASC
		`

		rows, err := db.Query(query)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch roles"})
			return
		}
		defer rows.Close()

		var roles []map[string]interface{}
		for rows.Next() {
			var id int
			var name string
			if err := rows.Scan(&id, &name); err != nil {
				continue
			}
			roles = append(roles, map[string]interface{}{
				"id":   id,
				"name": name,
			})
		}

		c.JSON(http.StatusOK, roles)
	}
}

// GetBeds returns beds filtered by status
// GET /api/dropdowns/beds?status={available/all}
func GetBeds(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		status := c.DefaultQuery("status", "available")

		var query string
		var rows *sql.Rows
		var err error

		if status == "all" {
			query = `
				SELECT 
					b.id, b.room_number, b.bed_number, b.department_id, b.status, d.name as department_name
				FROM beds b
				LEFT JOIN departments d ON b.department_id = d.id
				ORDER BY b.room_number ASC, b.bed_number ASC
			`
			rows, err = db.Query(query)
		} else {
			// Default to available beds
			query = `
				SELECT 
					b.id, b.room_number, b.bed_number, b.department_id, b.status, d.name as department_name
				FROM beds b
				LEFT JOIN departments d ON b.department_id = d.id
				WHERE b.status = $1
				ORDER BY b.room_number ASC, b.bed_number ASC
			`
			rows, err = db.Query(query, status)
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch beds"})
			return
		}
		defer rows.Close()

		var beds []map[string]interface{}
		for rows.Next() {
			var id, roomNumber, bedNumber, departmentID, bedStatus string
			var departmentName sql.NullString

			if err := rows.Scan(&id, &roomNumber, &bedNumber, &departmentID, &bedStatus, &departmentName); err != nil {
				continue
			}
			beds = append(beds, map[string]interface{}{
				"id":              id,
				"room_number":     roomNumber,
				"bed_number":      bedNumber,
				"department_id":   departmentID,
				"status":          bedStatus,
				"department_name": departmentName.String,
			})
		}

		c.JSON(http.StatusOK, beds)
	}
}

// GetBedsOccupancyStats returns occupancy statistics for beds
// GET /api/admin/beds/occupancy-stats
func GetBedsOccupancyStats(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get total and occupied beds
		statsQuery := `
			SELECT 
				COUNT(*) as total_beds,
				COUNT(CASE WHEN status = 'occupied' THEN 1 END) as occupied_beds
			FROM beds
		`

		var totalBeds, occupiedBeds int
		err := db.QueryRow(statsQuery).Scan(&totalBeds, &occupiedBeds)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch occupancy stats"})
			return
		}

		occupancyPercentage := 0.0
		if totalBeds > 0 {
			occupancyPercentage = (float64(occupiedBeds) / float64(totalBeds)) * 100
		}

		// Get per-department breakdown
		deptQuery := `
			SELECT 
				d.id,
				d.name,
				COUNT(b.id) as total,
				COUNT(CASE WHEN b.status = 'occupied' THEN 1 END) as occupied
			FROM departments d
			LEFT JOIN beds b ON d.id = b.department_id
			WHERE d.is_active = true
			GROUP BY d.id, d.name
			ORDER BY d.name ASC
		`

		deptRows, err := db.Query(deptQuery)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch department breakdown"})
			return
		}
		defer deptRows.Close()

		var byDepartment []map[string]interface{}
		for deptRows.Next() {
			var deptID, deptName string
			var total, occupied int
			if err := deptRows.Scan(&deptID, &deptName, &total, &occupied); err != nil {
				continue
			}

			deptPercentage := 0.0
			if total > 0 {
				deptPercentage = (float64(occupied) / float64(total)) * 100
			}

			byDepartment = append(byDepartment, map[string]interface{}{
				"department_id":        deptID,
				"department_name":      deptName,
				"total":                total,
				"occupied":             occupied,
				"available":            total - occupied,
				"occupancy_percentage": deptPercentage,
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"total_beds":           totalBeds,
			"occupied_beds":        occupiedBeds,
			"available_beds":       totalBeds - occupiedBeds,
			"occupancy_percentage": occupancyPercentage,
			"by_department":        byDepartment,
		})
	}
}

// CreateDepartmentRequest is the request body for creating a department
type CreateDepartmentRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

// CreateDepartment creates a new department
// POST /api/dropdowns/departments
func CreateDepartment(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateDepartmentRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
			return
		}

		query := `
			INSERT INTO departments (name, description, is_active, created_at)
			VALUES ($1, $2, true, NOW())
			RETURNING id, name, description, is_active, created_at
		`

		var deptID, name, description string
		var isActive bool
		var createdAt interface{}

		err := db.QueryRow(query, req.Name, req.Description).Scan(&deptID, &name, &description, &isActive, &createdAt)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to create department: " + err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"id":          deptID,
			"name":        name,
			"description": description,
			"is_active":   isActive,
			"message":     "Department created successfully",
		})
	}
}

// CreateRoleRequest is the request body for creating a role
type CreateRoleRequest struct {
	Name string `json:"name" binding:"required"`
}

// CreateRole creates a new role
// POST /api/dropdowns/roles
func CreateRole(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateRoleRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
			return
		}

		query := `
			INSERT INTO roles (name)
			VALUES ($1)
			RETURNING id, name
		`

		var roleID int
		var name string

		err := db.QueryRow(query, req.Name).Scan(&roleID, &name)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to create role: " + err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"id":      roleID,
			"name":    name,
			"message": "Role created successfully",
		})
	}
}
