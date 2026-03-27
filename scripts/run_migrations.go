package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	// Load .env file
	err := godotenv.Load("../.env")
	if err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	// Database connection string
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_SSLMODE"),
	)

	// Connect to database
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	// Test connection
	err = db.Ping()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	log.Println("✓ Database connected successfully")

	// Get migration files
	migrationDir := "../migrations"
	files, err := ioutil.ReadDir(migrationDir)
	if err != nil {
		log.Fatal("Failed to read migrations directory:", err)
	}

	// Filter and sort .sql files
	var migrationFiles []string
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".sql") {
			migrationFiles = append(migrationFiles, file.Name())
		}
	}
	sort.Strings(migrationFiles)

	// Execute each migration
	for _, fileName := range migrationFiles {
		filePath := filepath.Join(migrationDir, fileName)
		content, err := ioutil.ReadFile(filePath)
		if err != nil {
			log.Printf("Error reading migration file %s: %v\n", fileName, err)
			continue
		}

		// Execute migration
		_, err = db.Exec(string(content))
		if err != nil {
			log.Printf("❌ Error executing migration %s: %v\n", fileName, err)
			continue
		}
		log.Printf("✓ Migration executed: %s\n", fileName)
	}

	log.Println("✓ All migrations completed successfully")
}
