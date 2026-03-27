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
	}

	// Admin routes
	adminRoutes := r.Group("/api/admin")
	// adminRoutes.Use(middleware.AuthMiddleware(cfg))
	// adminRoutes.Use(middleware.RoleMiddleware("Admin"))
	{
		// Dashboard
		adminRoutes.GET("/dashboard", handlers.GetAdminDashboard(db))

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
	}
}
