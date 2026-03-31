package scripts

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

/**
 * Comprehensive Seed Database Script
 * Populates HMS database with test data for all 6 roles:
 * Admin, Doctor, Nurse, Receptionist, Pharmacist, Patient
 */

// SeedData structure for organized test data
type SeedData struct {
	db *sql.DB
}

func NewSeedData(db *sql.DB) *SeedData {
	return &SeedData{db: db}
}

// ==================== SEED EXECUTION ====================

func (s *SeedData) SeedAll() error {
	fmt.Println("Starting comprehensive database seeding...")

	// Seed in order: users first, then dependent tables
	if err := s.seedAdmins(); err != nil {
		return fmt.Errorf("seeding admins failed: %w", err)
	}
	fmt.Println("✓ Seeded admins")

	if err := s.seedDoctors(); err != nil {
		return fmt.Errorf("seeding doctors failed: %w", err)
	}
	fmt.Println("✓ Seeded doctors")

	if err := s.seedNurses(); err != nil {
		return fmt.Errorf("seeding nurses failed: %w", err)
	}
	fmt.Println("✓ Seeded nurses")

	if err := s.seedReceptionists(); err != nil {
		return fmt.Errorf("seeding receptionists failed: %w", err)
	}
	fmt.Println("✓ Seeded receptionists")

	if err := s.seedPharmacists(); err != nil {
		return fmt.Errorf("seeding pharmacists failed: %w", err)
	}
	fmt.Println("✓ Seeded pharmacists")

	if err := s.seedPatients(); err != nil {
		return fmt.Errorf("seeding patients failed: %w", err)
	}
	fmt.Println("✓ Seeded patients")

	if err := s.seedMedicines(); err != nil {
		return fmt.Errorf("seeding medicines failed: %w", err)
	}
	fmt.Println("✓ Seeded medicines")

	if err := s.seedAppointments(); err != nil {
		return fmt.Errorf("seeding appointments failed: %w", err)
	}
	fmt.Println("✓ Seeded appointments")

	if err := s.seedPrescriptions(); err != nil {
		return fmt.Errorf("seeding prescriptions failed: %w", err)
	}
	fmt.Println("✓ Seeded prescriptions")

	if err := s.seedVitals(); err != nil {
		return fmt.Errorf("seeding vitals failed: %w", err)
	}
	fmt.Println("✓ Seeded vitals")

	fmt.Println("\n✅ Database seeding completed successfully!")
	return nil
}

// ==================== SEED IMPLEMENTATIONS ====================

func (s *SeedData) seedAdmins() error {
	admins := []map[string]interface{}{
		{
			"email":      "admin@hospital.com",
			"password":   "Admin@123",
			"first_name": "Admin",
			"last_name":  "User",
			"phone":      "9876543210",
			"address":    "Hospital Headquarters",
		},
		{
			"email":      "superadmin@hospital.com",
			"password":   "SuperAdmin@123",
			"first_name": "Super",
			"last_name":  "Admin",
			"phone":      "9876543211",
			"address":    "Hospital Headquarters",
		},
	}

	for _, admin := range admins {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(admin["password"].(string)), bcrypt.DefaultCost)
		if err != nil {
			return err
		}

		query := `
			INSERT INTO users (email, password, first_name, last_name, phone, address, role, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
			ON CONFLICT DO NOTHING
		`

		_, err = s.db.Exec(query,
			admin["email"],
			hashedPassword,
			admin["first_name"],
			admin["last_name"],
			admin["phone"],
			admin["address"],
			"Admin",
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *SeedData) seedDoctors() error {
	doctors := []map[string]interface{}{
		{
			"email":            "dr.sharma@hospital.com",
			"password":         "Doctor@123",
			"first_name":       "Raj",
			"last_name":        "Sharma",
			"phone":            "9876543212",
			"address":          "123 Medical Lane, City",
			"specialization":   "Cardiology",
			"license_number":   "MED001",
			"qualification":    "MD, DM Cardiology",
			"years_experience": 15,
			"consultation_fee": 500,
		},
		{
			"email":            "dr.patel@hospital.com",
			"password":         "Doctor@123",
			"first_name":       "Priya",
			"last_name":        "Patel",
			"phone":            "9876543213",
			"address":          "456 Health St, City",
			"specialization":   "Orthopedics",
			"license_number":   "MED002",
			"qualification":    "MD, DNB Orthopedics",
			"years_experience": 12,
			"consultation_fee": 400,
		},
		{
			"email":            "dr.singh@hospital.com",
			"password":         "Doctor@123",
			"first_name":       "Arjun",
			"last_name":        "Singh",
			"phone":            "9876543214",
			"address":          "789 Wellness Ave, City",
			"specialization":   "General Medicine",
			"license_number":   "MED003",
			"qualification":    "MD, MRCP",
			"years_experience": 10,
			"consultation_fee": 300,
		},
	}

	for _, doctor := range doctors {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(doctor["password"].(string)), bcrypt.DefaultCost)
		if err != nil {
			return err
		}

		query := `
			INSERT INTO users (email, password, first_name, last_name, phone, address, role, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
			ON CONFLICT DO NOTHING
			RETURNING id
		`

		var userID int
		err = s.db.QueryRow(query,
			doctor["email"],
			hashedPassword,
			doctor["first_name"],
			doctor["last_name"],
			doctor["phone"],
			doctor["address"],
			"Doctor",
		).Scan(&userID)

		if err != nil && err != sql.ErrNoRows {
			return err
		}

		// Insert doctor-specific details
		if userID > 0 {
			doctorQuery := `
				INSERT INTO doctors (user_id, specialization, license_number, qualification, years_experience, consultation_fee, created_at)
				VALUES ($1, $2, $3, $4, $5, $6, NOW())
				ON CONFLICT DO NOTHING
			`
			_, err = s.db.Exec(doctorQuery,
				userID,
				doctor["specialization"],
				doctor["license_number"],
				doctor["qualification"],
				doctor["years_experience"],
				doctor["consultation_fee"],
			)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *SeedData) seedNurses() error {
	nurses := []map[string]interface{}{
		{
			"email":            "nurse.sharma@hospital.com",
			"password":         "Nurse@123",
			"first_name":       "Sneha",
			"last_name":        "Sharma",
			"phone":            "9876543215",
			"address":          "101 Care Lane, City",
			"department":       "General Ward",
			"license_number":   "NURSE001",
			"qualification":    "BSc Nursing, RN",
			"years_experience": 8,
		},
		{
			"email":            "nurse.pandey@hospital.com",
			"password":         "Nurse@123",
			"first_name":       "Aisha",
			"last_name":        "Pandey",
			"phone":            "9876543216",
			"address":          "202 Care Ave, City",
			"department":       "ICU",
			"license_number":   "NURSE002",
			"qualification":    "BSc Nursing, ICU Cert",
			"years_experience": 6,
		},
	}

	for _, nurse := range nurses {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(nurse["password"].(string)), bcrypt.DefaultCost)
		if err != nil {
			return err
		}

		query := `
			INSERT INTO users (email, password, first_name, last_name, phone, address, role, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
			ON CONFLICT DO NOTHING
			RETURNING id
		`

		var userID int
		err = s.db.QueryRow(query,
			nurse["email"],
			hashedPassword,
			nurse["first_name"],
			nurse["last_name"],
			nurse["phone"],
			nurse["address"],
			"Nurse",
		).Scan(&userID)

		if err != nil && err != sql.ErrNoRows {
			return err
		}
	}

	return nil
}

func (s *SeedData) seedReceptionists() error {
	receptionists := []map[string]interface{}{
		{
			"email":      "reception.joshi@hospital.com",
			"password":   "Reception@123",
			"first_name": "Kavya",
			"last_name":  "Joshi",
			"phone":      "9876543217",
			"address":    "303 Front Desk, City",
		},
		{
			"email":      "reception.gupta@hospital.com",
			"password":   "Reception@123",
			"first_name": "Anjali",
			"last_name":  "Gupta",
			"phone":      "9876543218",
			"address":    "304 Front Desk, City",
		},
	}

	for _, receptionist := range receptionists {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(receptionist["password"].(string)), bcrypt.DefaultCost)
		if err != nil {
			return err
		}

		query := `
			INSERT INTO users (email, password, first_name, last_name, phone, address, role, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
			ON CONFLICT DO NOTHING
		`

		_, err = s.db.Exec(query,
			receptionist["email"],
			hashedPassword,
			receptionist["first_name"],
			receptionist["last_name"],
			receptionist["phone"],
			receptionist["address"],
			"Receptionist",
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *SeedData) seedPharmacists() error {
	pharmacists := []map[string]interface{}{
		{
			"email":            "pharmacist.kumari@hospital.com",
			"password":         "Pharmacist@123",
			"first_name":       "Divya",
			"last_name":        "Kumari",
			"phone":            "9876543219",
			"address":          "405 Pharmacy, City",
			"license_number":   "PHARM001",
			"qualification":    "B.Pharmacy, RPh",
			"years_experience": 7,
		},
		{
			"email":            "pharmacist.sahni@hospital.com",
			"password":         "Pharmacist@123",
			"first_name":       "Neha",
			"last_name":        "Sahni",
			"phone":            "9876543220",
			"address":          "406 Pharmacy, City",
			"license_number":   "PHARM002",
			"qualification":    "M.Pharmacy, RPh",
			"years_experience": 5,
		},
	}

	for _, pharmacist := range pharmacists {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(pharmacist["password"].(string)), bcrypt.DefaultCost)
		if err != nil {
			return err
		}

		query := `
			INSERT INTO users (email, password, first_name, last_name, phone, address, role, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
			ON CONFLICT DO NOTHING
		`

		_, err = s.db.Exec(query,
			pharmacist["email"],
			hashedPassword,
			pharmacist["first_name"],
			pharmacist["last_name"],
			pharmacist["phone"],
			pharmacist["address"],
			"Pharmacist",
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *SeedData) seedPatients() error {
	patients := []map[string]interface{}{
		{
			"email":         "patient1@email.com",
			"password":      "Patient@123",
			"first_name":    "Ramesh",
			"last_name":     "Kumar",
			"phone":         "9876543221",
			"address":       "101 Patient Lane, City",
			"date_of_birth": "1980-05-15",
			"gender":        "Male",
			"blood_group":   "O+",
			"allergies":     "Penicillin",
		},
		{
			"email":         "patient2@email.com",
			"password":      "Patient@123",
			"first_name":    "Amrita",
			"last_name":     "Singh",
			"phone":         "9876543222",
			"address":       "102 Patient Lane, City",
			"date_of_birth": "1975-08-22",
			"gender":        "Female",
			"blood_group":   "A+",
			"allergies":     "Aspirin",
		},
		{
			"email":         "patient3@email.com",
			"password":      "Patient@123",
			"first_name":    "Vikram",
			"last_name":     "Patel",
			"phone":         "9876543223",
			"address":       "103 Patient Lane, City",
			"date_of_birth": "1990-03-10",
			"gender":        "Male",
			"blood_group":   "B+",
			"allergies":     "None",
		},
	}

	for _, patient := range patients {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(patient["password"].(string)), bcrypt.DefaultCost)
		if err != nil {
			return err
		}

		fullName := patient["first_name"].(string) + " " + patient["last_name"].(string)

		query := `
			INSERT INTO users (name, email, password_hash, avatar_url, is_active, created_at, updated_at)
			VALUES ($1, $2, $3, '', true, NOW(), NOW())
			ON CONFLICT DO NOTHING
			RETURNING id
		`

		var userID string
		err = s.db.QueryRow(query,
			fullName,
			patient["email"],
			hashedPassword,
		).Scan(&userID)

		if err != nil && err != sql.ErrNoRows {
			return err
		}

		// Insert patient-specific details
		if userID != "" {
			patientQuery := `
				INSERT INTO patients (user_id, date_of_birth, gender, blood_group, phone, created_at)
				VALUES ($1, $2::date, $3, $4, $5, NOW())
				ON CONFLICT DO NOTHING
			`
			_, err = s.db.Exec(patientQuery,
				userID,
				patient["date_of_birth"],
				patient["gender"],
				patient["blood_group"],
				patient["phone"],
			)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *SeedData) seedMedicines() error {
	medicines := []map[string]interface{}{
		{"name": "Aspirin", "dosage": "500mg", "unit": "tablet", "price": 10, "stock": 1000},
		{"name": "Paracetamol", "dosage": "500mg", "unit": "tablet", "price": 8, "stock": 1500},
		{"name": "Ibuprofen", "dosage": "400mg", "unit": "tablet", "price": 12, "stock": 800},
		{"name": "Amoxicillin", "dosage": "500mg", "unit": "capsule", "price": 25, "stock": 500},
		{"name": "Ciprofloxacin", "dosage": "500mg", "unit": "tablet", "price": 30, "stock": 600},
		{"name": "Metformin", "dosage": "500mg", "unit": "tablet", "price": 15, "stock": 2000},
		{"name": "Lisinopril", "dosage": "10mg", "unit": "tablet", "price": 20, "stock": 1200},
		{"name": "Atorvastatin", "dosage": "20mg", "unit": "tablet", "price": 35, "stock": 900},
		{"name": "Omeprazole", "dosage": "20mg", "unit": "capsule", "price": 18, "stock": 700},
		{"name": "Ranitidine", "dosage": "150mg", "unit": "tablet", "price": 12, "stock": 600},
	}

	for _, med := range medicines {
		query := `
			INSERT INTO medicines (name, dosage, unit, price, stock, reorder_level, created_at)
			VALUES ($1, $2, $3, $4, $5, 100, NOW())
			ON CONFLICT DO NOTHING
		`

		_, err := s.db.Exec(query,
			med["name"],
			med["dosage"],
			med["unit"],
			med["price"],
			med["stock"],
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *SeedData) seedAppointments() error {
	// Get doctor and patient IDs
	var doctorID, patientID int

	s.db.QueryRow("SELECT id FROM users WHERE email = 'dr.sharma@hospital.com' AND role = 'Doctor'").Scan(&doctorID)
	s.db.QueryRow("SELECT id FROM users WHERE email = 'patient1@email.com' AND role = 'Patient'").Scan(&patientID)

	if doctorID == 0 || patientID == 0 {
		return nil // Skip if test data not found
	}

	appointments := []map[string]interface{}{
		{
			"doctor_id":        doctorID,
			"patient_id":       patientID,
			"appointment_date": time.Now().AddDate(0, 0, 1),
			"appointment_time": "10:00:00",
			"chief_complaint":  "Chest pain and breathing difficulty",
			"appointment_type": "Follow-up",
			"status":           "Scheduled",
		},
	}

	for _, apt := range appointments {
		query := `
			INSERT INTO appointments (doctor_id, patient_id, appointment_date, appointment_time, chief_complaint, appointment_type, status, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
			ON CONFLICT DO NOTHING
		`

		_, err := s.db.Exec(query,
			apt["doctor_id"],
			apt["patient_id"],
			apt["appointment_date"],
			apt["appointment_time"],
			apt["chief_complaint"],
			apt["appointment_type"],
			apt["status"],
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *SeedData) seedPrescriptions() error {
	// Get doctor and patient IDs
	var doctorID, patientID int

	s.db.QueryRow("SELECT id FROM users WHERE email = 'dr.sharma@hospital.com' AND role = 'Doctor'").Scan(&doctorID)
	s.db.QueryRow("SELECT id FROM users WHERE email = 'patient1@email.com' AND role = 'Patient'").Scan(&patientID)

	if doctorID == 0 || patientID == 0 {
		return nil
	}

	query := `
		INSERT INTO prescriptions (doctor_id, patient_id, consultation_notes, created_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT DO NOTHING
	`

	_, err := s.db.Exec(query,
		doctorID,
		patientID,
		"Patient presenting with chest pain. Prescribed medication for cardiac monitoring.",
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *SeedData) seedVitals() error {
	// Get patient ID
	var patientID int
	s.db.QueryRow("SELECT id FROM users WHERE email = 'patient1@email.com' AND role = 'Patient'").Scan(&patientID)

	if patientID == 0 {
		return nil
	}

	vitals := []map[string]interface{}{
		{
			"patient_id":        patientID,
			"temperature":       98.6,
			"bp_systolic":       120,
			"bp_diastolic":      80,
			"heart_rate":        75,
			"respiratory_rate":  16,
			"oxygen_saturation": 98,
			"weight":            70,
			"height":            175,
		},
	}

	for _, vital := range vitals {
		query := `
			INSERT INTO vitals (patient_id, temperature, bp_systolic, bp_diastolic, heart_rate, respiratory_rate, oxygen_saturation, weight, height, recorded_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
			ON CONFLICT DO NOTHING
		`

		_, err := s.db.Exec(query,
			vital["patient_id"],
			vital["temperature"],
			vital["bp_systolic"],
			vital["bp_diastolic"],
			vital["heart_rate"],
			vital["respiratory_rate"],
			vital["oxygen_saturation"],
			vital["weight"],
			vital["height"],
		)
		if err != nil {
			return err
		}
	}

	return nil
}

// ==================== MAIN ====================

func ComprehensiveSeed() {
	// Database connection
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://user:password@localhost:5432/hms_db?sslmode=disable"
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	fmt.Println("Connected to database")

	seed := NewSeedData(db)
	if err := seed.SeedAll(); err != nil {
		log.Fatalf("Seeding failed: %v", err)
	}
}
