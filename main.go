package main

import (
	"log"

	"shalby_backend/config"
	"shalby_backend/database"
	"shalby_backend/routes"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func init() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}
}

func main() {
	// Load configuration
	cfg := config.Load()
	log.Println("✓ Configuration loaded")

	// Connect to database
	db := database.Connect(cfg)
	defer db.Close()

	// Create Gin router
	r := gin.Default()

	// Setup routes
	routes.SetupRoutes(r, db, cfg)
	log.Println("✓ Routes setup complete")

	// Start server
	port := ":" + cfg.Port
	log.Printf("🚀 Shalby HMS Backend running on http://localhost%s\n", port)
	log.Printf("📝 API docs available at http://localhost%s/api/health\n", port)
	log.Printf("🔐 Authentication endpoints:\n")
	log.Printf("   - POST http://localhost%s/api/auth/login\n", port)
	log.Printf("   - POST http://localhost%s/api/auth/signup\n", port)

	if err := r.Run(port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
