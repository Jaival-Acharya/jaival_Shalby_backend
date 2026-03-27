package main

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// GenerateBcryptHashesForSeed - Helper to generate bcrypt password hashes for seed data
// To use: Comment out the main function above and uncomment this to generate hashes
func GenerateBcryptHashesForSeed() {
	// List of test passwords for different user types
	passwords := map[string]string{
		"Admin@123":      "admin@shalby.com",
		"Doctor@123":     "doctor@shalby.com",
		"Pharmacist@123": "pharmacist@shalby.com",
		"Patient@123":    "john@example.com & jane@example.com",
	}

	fmt.Println("========================================")
	fmt.Println("Shalby HMS - Bcrypt Hash Generator")
	fmt.Println("========================================\n")
	fmt.Println("Use these hashes in seed.sql for test user passwords\n")

	for password, email := range passwords {
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			fmt.Printf("Error generating hash for %s: %v\n", password, err)
			continue
		}

		fmt.Printf("Email(s): %s\n", email)
		fmt.Printf("Password: %s\n", password)
		fmt.Printf("Bcrypt Hash:\n%s\n", string(hash))
		fmt.Println("----------------------------------------\n")
	}

	fmt.Println("✓ Copy the above hashes into seed.sql")
	fmt.Println("✓ Run: psql -U shalby_hospital -d shalby_hospital -f seed.sql")
	fmt.Println("✓ Then start the backend: go run main.go")
}
