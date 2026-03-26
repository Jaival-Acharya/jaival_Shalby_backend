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

	// Protected routes (require authentication - add middleware here when needed)
	// protectedRoutes := r.Group("/api")
	// protectedRoutes.Use(middleware.AuthMiddleware(cfg))
	// {
	//     // Add protected routes here
	// }
}
