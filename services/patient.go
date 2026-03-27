package services

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"shalby_backend/models"

	"golang.org/x/crypto/bcrypt"
)

// generateDefaultPassword creates a password from the first name
func generatePatientDefaultPassword(fullName string) string {
	// Extract first name from full name
	parts := strings.Fields(strings.TrimSpace(fullName))
	if len(parts) == 0 {
		return "Patient@123"
	}
	firstName := strings.ToLower(parts[0])
	return firstName + "@123"
}

type PatientDetail struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Age       int    `json:"age"`
	Gender    string `json:"gender"`
	Blood     string `json:"blood"`
	Phone     string `json:"phone"`
	Condition string `json:"condition"`
	LastVisit string `json:"lastVisit"`
	Meds      int    `json:"meds"`
}

// GetAllPatients returns all patients
func GetAllPatients(db *sql.DB) ([]PatientDetail, error) {
	// First check if patients table exists
	checkTableQuery := `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables 
			WHERE table_name = 'patients'
		)
	`
	var tableExists bool
	err := db.QueryRow(checkTableQuery).Scan(&tableExists)
	if err != nil {
		return nil, fmt.Errorf("error checking table existence: %v", err)
	}

	if !tableExists {
		return nil, fmt.Errorf("patients table does not exist in database")
	}

	// Now try the main query
	rows, err := db.Query(`
		SELECT 
			p.id, u.name, u.email, 
			p.date_of_birth,
			p.gender, p.blood_group, p.phone
		FROM patients p
		JOIN users u ON p.user_id = u.id
		ORDER BY u.name
	`)
	if err != nil {
		return nil, fmt.Errorf("error querying patients: %v", err)
	}
	defer rows.Close()

	var patients []PatientDetail

	for rows.Next() {
		var p PatientDetail
		var phone sql.NullString
		var dateOfBirth sql.NullTime

		err := rows.Scan(&p.ID, &p.Name, &p.Email, &dateOfBirth, &p.Gender, &p.Blood, &phone)
		if err != nil {
			return nil, fmt.Errorf("error scanning patient row: %v", err)
		}

		// Calculate age from date of birth
		if dateOfBirth.Valid {
			today := time.Now()
			p.Age = today.Year() - dateOfBirth.Time.Year()
			// Adjust if birthday hasn't occurred this year
			if today.Month() < dateOfBirth.Time.Month() ||
				(today.Month() == dateOfBirth.Time.Month() && today.Day() < dateOfBirth.Time.Day()) {
				p.Age--
			}
		}

		if phone.Valid {
			p.Phone = phone.String
		}

		// Get medicine count - simplified query without status filter
		err = db.QueryRow(
			"SELECT COUNT(*) FROM prescriptions WHERE patient_id = $1",
			p.ID,
		).Scan(&p.Meds)
		if err != nil && err != sql.ErrNoRows {
			// TODO: In real scenarios, log this error but continue
			p.Meds = 0
		}

		// Get last visit
		err = db.QueryRow(
			"SELECT MAX(appointment_date) FROM appointments WHERE patient_id = $1",
			p.ID,
		).Scan(&p.LastVisit)
		if err != nil && err != sql.ErrNoRows {
			// TODO: In real scenarios, log this error but continue
			p.LastVisit = ""
		}

		patients = append(patients, p)
	}

	return patients, rows.Err()
}

// CreatePatient creates a new patient with user account
func CreatePatient(db *sql.DB, req models.PatientCreateRequest) (interface{}, error) {
	// Auto-generate password if not provided
	password := req.Password
	if password == "" {
		password = generatePatientDefaultPassword(req.Name)
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Println("Error hashing password:", err)
		return nil, fmt.Errorf("failed to hash password")
	}

	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		log.Println("Error starting transaction:", err)
		return nil, err
	}
	defer tx.Rollback()

	// Create user
	var userID string
	userQuery := `
	INSERT INTO users (name, email, password_hash, is_active)
	VALUES ($1, $2, $3, true)
	RETURNING id
	`

	err = tx.QueryRow(userQuery, req.Name, req.Email, string(hashedPassword)).Scan(&userID)
	if err != nil {
		log.Println("Error creating user:", err)
		return nil, fmt.Errorf("failed to create user account")
	}

	// Assign Patient role
	roleQuery := `
	INSERT INTO user_roles (user_id, role_id)
	VALUES ($1, (SELECT id FROM roles WHERE name = 'Patient'))
	`

	_, err = tx.Exec(roleQuery, userID)
	if err != nil {
		log.Println("Error assigning patient role:", err)
		return nil, fmt.Errorf("failed to assign patient role")
	}

	// Parse date of birth
	var dateOfBirth time.Time
	dateOfBirth, err = time.Parse("2006-01-02", req.DateOfBirth)
	if err != nil {
		return nil, fmt.Errorf("invalid date of birth format, use YYYY-MM-DD")
	}

	// Create patient record
	var patientID string
	patientQuery := `
	INSERT INTO patients (
		user_id, date_of_birth, gender, blood_group, phone,
		insurance_provider, insurance_policy_number
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	RETURNING id
	`

	err = tx.QueryRow(
		patientQuery,
		userID,
		dateOfBirth,
		req.Gender,
		req.BloodGroup,
		req.Phone,
		req.InsuranceProvider,
		req.InsurancePolicyNumber,
	).Scan(&patientID)
	if err != nil {
		log.Println("Error creating patient:", err)
		return nil, fmt.Errorf("failed to create patient record")
	}

	// Insert emergency contact if provided
	if req.EmergencyContactName != "" && req.EmergencyContactPhone != "" {
		_, err := tx.Exec(
			"INSERT INTO patient_emergency_contacts (patient_id, name, phone, relation) VALUES ($1, $2, $3, $4)",
			patientID,
			req.EmergencyContactName,
			req.EmergencyContactPhone,
			req.EmergencyContactRelation,
		)
		if err != nil {
			log.Println("Error inserting emergency contact:", err)
		}
	}

	// Insert allergies if provided
	if len(req.Allergies) > 0 {
		for _, allergy := range req.Allergies {
			if allergy != "" {
				_, err := tx.Exec(
					"INSERT INTO patient_allergies (patient_id, allergy) VALUES ($1, $2)",
					patientID, allergy,
				)
				if err != nil {
					log.Println("Error inserting allergy:", err)
				}
			}
		}
	}

	// Insert conditions if provided
	if len(req.Conditions) > 0 {
		for _, condition := range req.Conditions {
			if condition != "" {
				_, err := tx.Exec(
					"INSERT INTO patient_chronic_conditions (patient_id, condition) VALUES ($1, $2)",
					patientID, condition,
				)
				if err != nil {
					log.Println("Error inserting condition:", err)
				}
			}
		}
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		log.Println("Error committing transaction:", err)
		return nil, err
	}

	return map[string]interface{}{
		"id":    patientID,
		"name":  req.Name,
		"email": req.Email,
	}, nil
}

// GetPatientDetails returns detailed information about a patient
func GetPatientDetails(db *sql.DB, patientID string) (interface{}, error) {
	var patient PatientDetail

	err := db.QueryRow(`
		SELECT 
			p.id, u.name, u.email, 
			EXTRACT(YEAR FROM AGE(p.date_of_birth))::int,
			p.gender, p.blood_group, p.phone
		FROM patients p
		JOIN users u ON p.user_id = u.id
		WHERE p.id = $1
	`, patientID).Scan(&patient.ID, &patient.Name, &patient.Email, &patient.Age,
		&patient.Gender, &patient.Blood, &patient.Phone)

	if err == sql.ErrNoRows {
		return nil, errors.New("Patient not found")
	}
	if err != nil {
		return nil, err
	}

	return patient, nil
}

// UpdatePatient updates patient information
func UpdatePatient(db *sql.DB, patientID string, data map[string]interface{}) (interface{}, error) {
	// Get user_id first
	var userID string
	err := db.QueryRow("SELECT user_id FROM patients WHERE id = $1", patientID).Scan(&userID)
	if err == sql.ErrNoRows {
		return nil, errors.New("Patient not found")
	}
	if err != nil {
		return nil, err
	}

	// Update user info if provided
	if name, ok := data["name"]; ok {
		_, err = db.Exec("UPDATE users SET name = $1 WHERE id = $2", name, userID)
		if err != nil {
			return nil, err
		}
	}

	// Update patient info if provided
	if gender, ok := data["gender"]; ok {
		_, err = db.Exec("UPDATE patients SET gender = $1 WHERE id = $2", gender, patientID)
		if err != nil {
			return nil, err
		}
	}

	return map[string]string{"message": "Patient updated successfully"}, nil
}

// DeletePatient deletes a patient
func DeletePatient(db *sql.DB, patientID string) error {
	// Get user_id
	var userID string
	err := db.QueryRow("SELECT user_id FROM patients WHERE id = $1", patientID).Scan(&userID)
	if err == sql.ErrNoRows {
		return errors.New("Patient not found")
	}
	if err != nil {
		return err
	}

	// Delete patient record
	_, err = db.Exec("DELETE FROM patients WHERE id = $1", patientID)
	if err != nil {
		return err
	}

	// Delete user roles
	_, err = db.Exec("DELETE FROM user_roles WHERE user_id = $1", userID)
	if err != nil {
		return err
	}

	// Delete user
	_, err = db.Exec("DELETE FROM users WHERE id = $1", userID)
	return err
}
