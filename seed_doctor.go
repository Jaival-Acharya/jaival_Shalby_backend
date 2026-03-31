package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

func main() {
	// Database connection string
	dsn := "user=postgres password=postgres dbname=shalby_hospital sslmode=disable"

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Test connection
	err = db.Ping()
	if err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	fmt.Println("✓ Connected to database")

	// Insert doctor profile for test doctor
	query := `
		INSERT INTO doctors (user_id, specialization, department, license_number, consultation_fee, joining_date, is_active, created_at)
		SELECT u.id, 'Cardiology', 'Cardiology', 'DOC12345', 500.00, $1, true, NOW()
		FROM users u WHERE u.email = 'doctor@shalby.com'
		ON CONFLICT (user_id) DO NOTHING;
	`

	result, err := db.Exec(query, time.Date(2023, 1, 15, 0, 0, 0, 0, time.UTC))
	if err != nil {
		log.Fatal("Failed to insert doctor profile:", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		log.Fatal("Failed to get rows affected:", err)
	}

	if rows > 0 {
		fmt.Printf("✓ Inserted %d doctor record(s)\n", rows)
	} else {
		fmt.Println("✓ Doctor record already exists")
	}

	// Insert doctor schedules
	for day := 1; day <= 5; day++ {
		scheduleQuery := `
			INSERT INTO doctor_schedules (doctor_id, day_of_week, start_time, end_time, slot_duration_minutes, max_patients_per_slot)
			SELECT d.id, $1, $2::TIME, $3::TIME, 30, 1
			FROM doctors d JOIN users u ON d.user_id = u.id WHERE u.email = 'doctor@shalby.com'
			ON CONFLICT DO NOTHING;
		`

		_, err := db.Exec(scheduleQuery, day, "09:00:00", "17:00:00")
		if err != nil {
			log.Fatal("Failed to insert doctor schedule:", err)
		}
	}

	fmt.Println("✓ Inserted doctor schedules")
	fmt.Println("\n✅ Seed data initialization complete!")
}
