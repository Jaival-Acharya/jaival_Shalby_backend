package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"shalby_backend/services"

	"github.com/gin-gonic/gin"
)

// CreateStaffRequest is the request body for creating a staff member
type CreateStaffRequest struct {
	Name           string `json:"name" binding:"required"`        // Full name
	Email          string `json:"email" binding:"required,email"` // Email address
	Phone          string `json:"phone" binding:"required"`       // Phone number
	Password       string `json:"password"`                       // Optional - auto-generated if empty
	Role           string `json:"role" binding:"required"`        // Doctor, Nurse, Receptionist, Pharmacist, Admin
	Department     string `json:"department"`                     // For Doctor, Nurse roles
	Address        string `json:"address"`                        // Address (optional)
	Specialization string `json:"specialization"`                 // For Doctor role
	LicenseNumber  string `json:"licenseNumber"`                  // For Doctor role
}

// UpdateStaffRequest is the request body for updating staff
type UpdateStaffRequest struct {
	Phone            string   `json:"phone"`
	ConsultationFee  *float64 `json:"consultation_fee,omitempty"`
	SpecializationID *string  `json:"specialization_id,omitempty"`
	DepartmentID     *string  `json:"department_id,omitempty"`
}

// CreateStaff handles creating a new staff member
// POST /api/admin/staff
func CreateStaff(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateStaffRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
			return
		}

		adminService := services.NewAdminService(db)
		staffID, userID, generatedPassword, err := adminService.CreateStaff(services.CreateStaffInput{
			Name:           req.Name,
			Email:          req.Email,
			Phone:          req.Phone,
			Password:       req.Password,
			Role:           req.Role,
			Department:     req.Department,
			Address:        req.Address,
			Specialization: req.Specialization,
			LicenseNumber:  req.LicenseNumber,
		})

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message":            "Staff member created successfully",
			"staff_id":           staffID,
			"user_id":            userID,
			"email":              req.Email,
			"name":               req.Name,
			"role":               req.Role,
			"generated_password": generatedPassword,
		})
	}
}

// GetAllStaff returns all staff members
// GET /api/admin/staff?limit=20&offset=0
func GetAllStaff(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		limit := 20
		offset := 0

		if limitStr := c.Query("limit"); limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
				limit = l
			}
		}

		if offsetStr := c.Query("offset"); offsetStr != "" {
			if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
				offset = o
			}
		}

		adminService := services.NewAdminService(db)
		staff, err := adminService.GetAllStaff(limit, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch staff"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"staff": staff,
			"count": len(staff),
		})
	}
}

// GetStaffByRole returns staff members by specific role
// GET /api/admin/staff/role/:role?limit=20&offset=0
func GetStaffByRole(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.Param("role")
		if role == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Role is required"})
			return
		}

		limit := 20
		offset := 0

		if limitStr := c.Query("limit"); limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
				limit = l
			}
		}

		if offsetStr := c.Query("offset"); offsetStr != "" {
			if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
				offset = o
			}
		}

		adminService := services.NewAdminService(db)
		staff, err := adminService.GetStaffByRole(role, limit, offset)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"role":  role,
			"staff": staff,
			"count": len(staff),
		})
	}
}

// UpdateStaff updates staff member information
// PUT /api/admin/staff/:staff_id/:role
func UpdateStaff(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		staffID := c.Param("staff_id")
		role := c.Param("role")

		if staffID == "" || role == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Staff ID and role are required"})
			return
		}

		var req UpdateStaffRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
			return
		}

		adminService := services.NewAdminService(db)
		err := adminService.UpdateStaff(staffID, services.UpdateStaffInput{
			FirstName:      "", // Not being updated in UI
			LastName:       "", // Not being updated in UI
			Phone:          req.Phone,
			Department:     "", // Not being updated in UI
			Specialization: req.SpecializationID,
		})

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Staff member updated successfully",
		})
	}
}

// DeactivateStaff marks a staff member as inactive
// POST /api/admin/staff/:staff_id/:role/deactivate
func DeactivateStaff(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		staffID := c.Param("staff_id")
		role := c.Param("role")

		if staffID == "" || role == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Staff ID and role are required"})
			return
		}

		adminService := services.NewAdminService(db)
		err := adminService.DeactivateStaff(staffID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Staff member deactivated successfully",
		})
	}
}

// ActivateStaff marks a staff member as active
// POST /api/admin/staff/:staff_id/:role/activate
func ActivateStaff(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		staffID := c.Param("staff_id")
		role := c.Param("role")

		if staffID == "" || role == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Staff ID and role are required"})
			return
		}

		adminService := services.NewAdminService(db)
		err := adminService.ActivateStaff(staffID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Staff member activated successfully",
		})
	}
}

// GetDoctorsInDepartment returns all doctors in a specific department
// GET /api/admin/doctors/department/:department
func GetDoctorsInDepartment(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		departmentID := c.Param("department")
		if departmentID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Department ID is required"})
			return
		}

		adminService := services.NewAdminService(db)
		doctors, err := adminService.GetDoctorsInDepartment(departmentID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch doctors"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"department_id": departmentID,
			"doctors":       doctors,
			"count":         len(doctors),
		})
	}
}
