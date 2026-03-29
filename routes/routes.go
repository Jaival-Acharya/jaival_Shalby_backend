package routes

import (
	"database/sql"

	"shalby_backend/config"
	"shalby_backend/handlers"
	"shalby_backend/middleware"

	"github.com/gin-gonic/gin"
)

// SetupRoutes configures all application routes
func SetupRoutes(r *gin.Engine, db *sql.DB, cfg *config.Config) {
	// Apply CORS middleware globally
	r.Use(middleware.CORSMiddleware())

	// Health check endpoint
	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Public routes (no authentication required)
	authRoutes := r.Group("/api/auth")
	{
		authRoutes.POST("/login", handlers.LoginHandler(db, cfg))
		authRoutes.POST("/signup", handlers.SignupHandler(db))
		authRoutes.POST("/logout", middleware.AuthMiddleware(cfg), handlers.LogoutHandler())
	}

	// Admin routes
	adminRoutes := r.Group("/api/admin")
	// adminRoutes.Use(middleware.AuthMiddleware(cfg))
	// adminRoutes.Use(middleware.RoleMiddleware("Admin"))
	{
		// Dashboard & Profile
		adminRoutes.GET("/dashboard", handlers.GetAdminDashboard(db))
		adminRoutes.GET("/profile", handlers.GetAdminProfile(db))
		adminRoutes.PATCH("/profile", handlers.UpdateAdminProfile(db))

		// Doctor Management
		adminRoutes.GET("/doctors", handlers.GetAllDoctors(db))
		adminRoutes.GET("/doctors/:id", handlers.GetDoctorByID(db))
		adminRoutes.POST("/doctors", handlers.CreateDoctor(db))
		adminRoutes.PUT("/doctors/:id", handlers.UpdateDoctor(db))
		adminRoutes.DELETE("/doctors/:id", handlers.DeleteDoctor(db))

		// Patient Management
		adminRoutes.GET("/patients", handlers.GetAllPatients(db))
		adminRoutes.POST("/patients", handlers.CreatePatient(db))
		adminRoutes.GET("/patients/:id", handlers.GetPatientDetails(db))
		adminRoutes.PUT("/patients/:id", handlers.UpdatePatient(db))
		adminRoutes.DELETE("/patients/:id", handlers.DeletePatient(db))

		// Medicine Management
		adminRoutes.GET("/medicines", handlers.GetAllMedicines(db))
		adminRoutes.GET("/medicines/stats", handlers.GetMedicineStats(db))
		adminRoutes.POST("/medicines", handlers.CreateMedicine(db))
		adminRoutes.PUT("/medicines/:id", handlers.UpdateMedicine(db))
		adminRoutes.DELETE("/medicines/:id", handlers.DeleteMedicine(db))

		// Settings
		adminRoutes.GET("/settings", handlers.GetSystemSettings(db))
		adminRoutes.PUT("/settings", handlers.UpdateSystemSettings(db))

		// Reports
		adminRoutes.GET("/reports", handlers.GetReports(db))

		// Staff Management
		adminRoutes.POST("/staff", handlers.CreateStaff(db))
		adminRoutes.GET("/staff", handlers.GetAllStaff(db))
		adminRoutes.GET("/staff/role/:role", handlers.GetStaffByRole(db))
		adminRoutes.PUT("/staff/:id", handlers.UpdateStaff(db))
		adminRoutes.PUT("/staff/:id/deactivate", handlers.DeactivateStaff(db))
		adminRoutes.PUT("/staff/:id/activate", handlers.ActivateStaff(db))
		adminRoutes.GET("/doctors/department/:department", handlers.GetDoctorsInDepartment(db))
	}

	// Nurse routes
	nurseRoutes := r.Group("/api/nurse")
	// nurseRoutes.Use(middleware.AuthMiddleware(cfg))
	// nurseRoutes.Use(middleware.RoleMiddleware("Nurse"))
	{
		nurseRoutes.POST("/vitals", handlers.RecordVitals(db))
		nurseRoutes.GET("/patients/checkedin", handlers.GetCheckedInPatients(db))
		nurseRoutes.GET("/patients/:id/vitals", handlers.GetPatientVitals(db))
	}

	// Receptionist routes
	receptionistRoutes := r.Group("/api/receptionist")
	// receptionistRoutes.Use(middleware.AuthMiddleware(cfg))
	// receptionistRoutes.Use(middleware.RoleMiddleware("Receptionist"))
	{
		receptionistRoutes.POST("/patients/register", handlers.RegisterPatient(db))
		receptionistRoutes.POST("/appointments/book", handlers.BookAppointment(db))
		receptionistRoutes.POST("/patients/checkin", handlers.CheckInPatient(db))
		receptionistRoutes.GET("/appointments/pending", handlers.GetPendingAppointments(db))
	}

	// Dropdown/Lookup routes
	dropdownRoutes := r.Group("/api/dropdowns")
	{
		dropdownRoutes.GET("/departments", handlers.GetDepartments(db))
		dropdownRoutes.POST("/departments", handlers.CreateDepartment(db))
		dropdownRoutes.GET("/specializations", handlers.GetSpecializations(db))
		dropdownRoutes.GET("/allergies", handlers.GetAllergies(db))
		dropdownRoutes.GET("/conditions", handlers.GetConditions(db))
		dropdownRoutes.GET("/medicine-categories", handlers.GetMedicineCategories(db))
		dropdownRoutes.POST("/medicine-categories", handlers.CreateMedicineCategory(db))
		dropdownRoutes.GET("/medicine-generic-names", handlers.GetMedicineGenericNames(db))
		dropdownRoutes.GET("/cities", handlers.GetCities(db))
		dropdownRoutes.GET("/roles", handlers.GetRoles(db))
		dropdownRoutes.POST("/roles", handlers.CreateRole(db))
		dropdownRoutes.GET("/beds", handlers.GetBeds(db))
		dropdownRoutes.GET("/beds/occupancy-stats", handlers.GetBedsOccupancyStats(db))
	}
}
