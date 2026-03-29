package services

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"shalby_backend/models"
)

// ReceptionistService handles business logic for receptionist operations
type ReceptionistService struct {
	db *sql.DB
}

// NewReceptionistService creates a new receptionist service instance
func NewReceptionistService(db *sql.DB) *ReceptionistService {
	return &ReceptionistService{db: db}
}

// RegisterPatientInput holds the data needed to register a new patient
type RegisterPatientInput struct {
	FirstName   string
	LastName    string
	Email       string
	Phone       string
	DateOfBirth string
	Gender      string
	CityID      int
	Address     string
	Allergies   []string
	Conditions  []string
}

// RegisterPatient creates a new user and patient record
func (rs *ReceptionistService) RegisterPatient(input RegisterPatientInput) (patientID string, userID string, err error) {
	tx, err := rs.db.Begin()
	if err != nil {
		return "", "", fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	var existingUserID string
	emailCheckQuery := "SELECT id FROM users WHERE email = $1 LIMIT 1"
	err = tx.QueryRow(emailCheckQuery, input.Email).Scan(&existingUserID)
	if err == nil {
		return "", "", errors.New("email already registered")
	}
	if err != sql.ErrNoRows {
		return "", "", fmt.Errorf("email check failed: %w", err)
	}

	userID, err = createUserWithRole(tx, models.User{
		Email: input.Email,
		Name:  input.FirstName + " " + input.LastName,
	}, "Patient")
	if err != nil {
		return "", "", fmt.Errorf("failed to create user: %w", err)
	}

	patientID, err = createPatientRecord(tx, userID, input)
	if err != nil {
		return "", "", fmt.Errorf("failed to create patient record: %w", err)
	}

	if len(input.Allergies) > 0 {
		err = addPatientAllergies(tx, patientID, input.Allergies)
		if err != nil {
			return "", "", fmt.Errorf("failed to add allergies: %w", err)
		}
	}

	if len(input.Conditions) > 0 {
		err = addPatientConditions(tx, patientID, input.Conditions)
		if err != nil {
			return "", "", fmt.Errorf("failed to add conditions: %w", err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return "", "", fmt.Errorf("failed to commit transaction: %w", err)
	}

	return patientID, userID, nil
}

// BookAppointmentInput holds appointment booking data
type BookAppointmentInput struct {
	PatientID       string
	DoctorID        string
	AppointmentDate string
	AppointmentTime string
	Complaint       string
}

// BookAppointment creates a new appointment record
func (rs *ReceptionistService) BookAppointment(input BookAppointmentInput) (appointmentID string, err error) {
	var patientExists int
	err = rs.db.QueryRow("SELECT 1 FROM patients WHERE id = $1", input.PatientID).Scan(&patientExists)
	if err == sql.ErrNoRows {
		return "", errors.New("patient not found")
	}
	if err != nil {
		return "", fmt.Errorf("failed to validate patient: %w", err)
	}

	var doctorExists int
	err = rs.db.QueryRow("SELECT 1 FROM doctors WHERE user_id = $1 AND is_active = true", input.DoctorID).Scan(&doctorExists)
	if err == sql.ErrNoRows {
		return "", errors.New("doctor not found or inactive")
	}
	if err != nil {
		return "", fmt.Errorf("failed to validate doctor: %w", err)
	}

	conflictQuery := `
		SELECT COUNT(*) as count
		FROM appointments
		WHERE doctor_id = $1 
		AND DATE(appointment_date) = $2::date
		AND TIME(appointment_date) = $3::time
		AND status != 'Cancelled'
	`
	var conflictCount int
	err = rs.db.QueryRow(conflictQuery, input.DoctorID, input.AppointmentDate, input.AppointmentTime).Scan(&conflictCount)
	if err != nil {
		return "", fmt.Errorf("failed to check conflicts: %w", err)
	}
	if conflictCount > 0 {
		return "", errors.New("doctor already has an appointment at this time")
	}

	patientConflictQuery := `
		SELECT COUNT(*) as count
		FROM appointments
		WHERE patient_id = $1 
		AND DATE(appointment_date) = $2::date
		AND TIME(appointment_date) = $3::time
		AND status != 'Cancelled'
	`
	var patientConflict int
	err = rs.db.QueryRow(patientConflictQuery, input.PatientID, input.AppointmentDate, input.AppointmentTime).Scan(&patientConflict)
	if err != nil {
		return "", fmt.Errorf("failed to check patient conflicts: %w", err)
	}
	if patientConflict > 0 {
		return "", errors.New("patient already has an appointment at this time")
	}

	appointmentDateTime := fmt.Sprintf("%s %s:00", input.AppointmentDate, input.AppointmentTime)
	createQuery := `
		INSERT INTO appointments (patient_id, doctor_id, appointment_date, complaint, status, created_at)
		VALUES ($1, $2, $3::timestamp, $4, 'Scheduled', NOW())
		RETURNING id
	`
	err = rs.db.QueryRow(createQuery, input.PatientID, input.DoctorID, appointmentDateTime, input.Complaint).Scan(&appointmentID)
	if err != nil {
		return "", fmt.Errorf("failed to create appointment: %w", err)
	}

	return appointmentID, nil
}

// CheckInPatientInput holds check-in data
type CheckInPatientInput struct {
	AppointmentID  string
	BedID          string
	ReceptionistID string
}

// CheckInPatient marks a patient as checked in and assigns a bed
func (rs *ReceptionistService) CheckInPatient(input CheckInPatientInput) error {
	tx, err := rs.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	var appointmentID, patientID string
	var currentStatus string
	apptQuery := "SELECT id, patient_id, status FROM appointments WHERE id = $1"
	err = tx.QueryRow(apptQuery, input.AppointmentID).Scan(&appointmentID, &patientID, &currentStatus)
	if err == sql.ErrNoRows {
		return errors.New("appointment not found")
	}
	if err != nil {
		return fmt.Errorf("failed to fetch appointment: %w", err)
	}

	if currentStatus != "Scheduled" {
		return fmt.Errorf("appointment is already %s, cannot check in", currentStatus)
	}

	var bedStatus string
	bedQuery := "SELECT status FROM beds WHERE id = $1"
	err = tx.QueryRow(bedQuery, input.BedID).Scan(&bedStatus)
	if err == sql.ErrNoRows {
		return errors.New("bed not found")
	}
	if err != nil {
		return fmt.Errorf("failed to validate bed: %w", err)
	}

	if bedStatus != "available" && bedStatus != "ready" {
		return fmt.Errorf("bed is not available (status: %s)", bedStatus)
	}

	updateApptQuery := `
		UPDATE appointments
		SET 
			status = 'Checked In',
			bed_id = $1,
			checked_in_at = NOW(),
			checked_in_by = $2
		WHERE id = $3
	`
	_, err = tx.Exec(updateApptQuery, input.BedID, input.ReceptionistID, input.AppointmentID)
	if err != nil {
		return fmt.Errorf("failed to update appointment: %w", err)
	}

	updateBedQuery := `
		UPDATE beds
		SET status = 'occupied'
		WHERE id = $1
	`
	_, err = tx.Exec(updateBedQuery, input.BedID)
	if err != nil {
		return fmt.Errorf("failed to update bed status: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetPendingAppointments returns appointments with pending status
func (rs *ReceptionistService) GetPendingAppointments(limit int, offset int) ([]map[string]interface{}, error) {
	query := `
		SELECT 
			a.id,
			a.patient_id,
			p.user_id,
			u.first_name,
			u.last_name,
			u.phone,
			a.appointment_date,
			a.complaint,
			a.status,
			d.user_id as doctor_id,
			du.first_name as doctor_first_name,
			du.last_name as doctor_last_name
		FROM appointments a
		JOIN patients p ON a.patient_id = p.id
		JOIN users u ON p.user_id = u.id
		JOIN doctors d ON a.doctor_id = d.id
		JOIN users du ON d.user_id = du.id
		WHERE a.status IN ('Scheduled', 'Checked In', 'Ready for Doctor')
		ORDER BY a.appointment_date ASC
		LIMIT $1 OFFSET $2
	`

	rows, err := rs.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch appointments: %w", err)
	}
	defer rows.Close()

	var appointments []map[string]interface{}
	for rows.Next() {
		var id, patientID, userID, firstName, lastName, phone string
		var appointmentDate time.Time
		var complaint, status, doctorID, doctorFirstName, doctorLastName string

		err := rows.Scan(&id, &patientID, &userID, &firstName, &lastName, &phone,
			&appointmentDate, &complaint, &status, &doctorID, &doctorFirstName, &doctorLastName)
		if err != nil {
			continue
		}

		appointments = append(appointments, map[string]interface{}{
			"id":               id,
			"patient_id":       patientID,
			"patient_name":     firstName + " " + lastName,
			"patient_phone":    phone,
			"appointment_date": appointmentDate,
			"complaint":        complaint,
			"status":           status,
			"doctor_name":      doctorFirstName + " " + doctorLastName,
			"doctor_id":        doctorID,
		})
	}

	return appointments, nil
}

// Helper functions

func createUserWithRole(tx *sql.Tx, user models.User, roleName string) (userID string, err error) {
	createUserQuery := `
		INSERT INTO users (email, name, password_hash, is_active, created_at)
		VALUES ($1, $2, $3, true, NOW())
		RETURNING id
	`
	// Use empty password hash for now - should be set during signup
	err = tx.QueryRow(createUserQuery, user.Email, user.Name, "").Scan(&userID)
	if err != nil {
		return "", fmt.Errorf("failed to create user: %w", err)
	}

	var roleID int
	getRoleQuery := "SELECT id FROM roles WHERE name = $1"
	err = tx.QueryRow(getRoleQuery, roleName).Scan(&roleID)
	if err != nil {
		return userID, fmt.Errorf("failed to get role ID: %w", err)
	}

	assignRoleQuery := `
		INSERT INTO user_roles (user_id, role_id)
		VALUES ($1, $2)
	`
	_, err = tx.Exec(assignRoleQuery, userID, roleID)
	if err != nil {
		return userID, fmt.Errorf("failed to assign role: %w", err)
	}

	return userID, nil
}

func createPatientRecord(tx *sql.Tx, userID string, input RegisterPatientInput) (patientID string, err error) {
	createQuery := `
		INSERT INTO patients (user_id, date_of_birth, gender, city_id, address, created_at)
		VALUES ($1, $2::date, $3, $4, $5, NOW())
		RETURNING id
	`
	err = tx.QueryRow(createQuery, userID, input.DateOfBirth, input.Gender, input.CityID, input.Address).Scan(&patientID)
	if err != nil {
		return "", fmt.Errorf("failed to create patient: %w", err)
	}
	return patientID, nil
}

func addPatientAllergies(tx *sql.Tx, patientID string, allergyIDs []string) error {
	for _, allergyID := range allergyIDs {
		query := `
			INSERT INTO patient_allergies (patient_id, allergy_id)
			VALUES ($1, $2)
			ON CONFLICT DO NOTHING
		`
		_, err := tx.Exec(query, patientID, allergyID)
		if err != nil {
			return fmt.Errorf("failed to add allergy: %w", err)
		}
	}
	return nil
}

func addPatientConditions(tx *sql.Tx, patientID string, conditionIDs []string) error {
	for _, conditionID := range conditionIDs {
		query := `
			INSERT INTO patient_conditions (patient_id, condition_id)
			VALUES ($1, $2)
			ON CONFLICT DO NOTHING
		`
		_, err := tx.Exec(query, patientID, conditionID)
		if err != nil {
			return fmt.Errorf("failed to add condition: %w", err)
		}
	}
	return nil
}
