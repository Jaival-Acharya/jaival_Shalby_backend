package services

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"shalby_backend/models"

	"golang.org/x/crypto/bcrypt"
)

// ReceptionistService handles business logic for receptionist operations
type ReceptionistService struct {
	db *sql.DB
}

// NewReceptionistService creates a new receptionist service instance
func NewReceptionistService(db *sql.DB) *ReceptionistService {
	return &ReceptionistService{db: db}
}

// EmergencyContactInput holds emergency contact data
type EmergencyContactInput struct {
	Name     string `json:"name"`
	Relation string `json:"relation"`
	Phone    string `json:"phone"`
}

// RegisterPatientInput holds the data needed to register a new patient
type RegisterPatientInput struct {
	FirstName         string
	LastName          string
	Email             string
	Phone             string
	Password          string
	DateOfBirth       string
	Gender            string
	BloodGroup        string
	CityID            int
	Address           string
	Allergies         []string
	Conditions        []string
	EmergencyContacts []EmergencyContactInput
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
	}, "Patient", input.Password)
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

	if len(input.EmergencyContacts) > 0 {
		err = addPatientEmergencyContacts(tx, patientID, input.EmergencyContacts)
		if err != nil {
			return "", "", fmt.Errorf("failed to add emergency contacts: %w", err)
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
	AppointmentType string
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
	err = rs.db.QueryRow("SELECT 1 FROM doctors WHERE id = $1 AND is_active = true", input.DoctorID).Scan(&doctorExists)
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
		AND appointment_date = $2::date
		AND appointment_time = $3::time
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
		AND appointment_date = $2::date
		AND appointment_time = $3::time
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

	createQuery := `
		INSERT INTO appointments (patient_id, doctor_id, appointment_date, appointment_time, chief_complaint, type, status, created_at)
		VALUES ($1, $2, $3::date, $4::time, $5, $6, 'Scheduled', NOW())
		RETURNING id
	`
	err = rs.db.QueryRow(createQuery, input.PatientID, input.DoctorID, input.AppointmentDate, input.AppointmentTime, input.Complaint, input.AppointmentType).Scan(&appointmentID)
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

	// Only validate bed if BedID is provided
	if input.BedID != "" {
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
	}

	// Prepare bed_id value - convert empty string to NULL
	var bedIDValue interface{} = nil
	if input.BedID != "" {
		bedIDValue = input.BedID
	}

	// Prepare checked_in_by value as UUID
	updateApptQuery := `
		UPDATE appointments
		SET 
			status = 'Checked In',
			bed_id = $1::uuid,
			checked_in_at = NOW(),
			checked_in_by = $2::uuid
		WHERE id = $3::uuid
	`
	_, err = tx.Exec(updateApptQuery, bedIDValue, input.ReceptionistID, input.AppointmentID)
	if err != nil {
		return fmt.Errorf("failed to update appointment: %w", err)
	}

	// Only update bed status if BedID is provided
	if input.BedID != "" {
		updateBedQuery := `
			UPDATE beds
			SET status = 'occupied'
			WHERE id = $1
		`
		_, err = tx.Exec(updateBedQuery, input.BedID)
		if err != nil {
			return fmt.Errorf("failed to update bed status: %w", err)
		}
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
			COALESCE(u.name, 'Unknown') as patient_name,
			COALESCE(u.phone, '') as phone,
			a.appointment_date,
			COALESCE(a.chief_complaint, '') as chief_complaint,
			a.status::text,
			d.user_id as doctor_id,
			COALESCE(du.name, 'Unknown') as doctor_name
		FROM appointments a
		JOIN patients p ON a.patient_id = p.id
		LEFT JOIN users u ON p.user_id = u.id
		JOIN doctors d ON a.doctor_id = d.id
		LEFT JOIN users du ON d.user_id = du.id
		WHERE a.status::text IN ('Scheduled', 'Checked In', 'Ready for Doctor')
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
		var id, patientID, userID, patientName, phone string
		var appointmentDate time.Time
		var chiefComplaint, status, doctorID, doctorName string

		err := rows.Scan(&id, &patientID, &userID, &patientName, &phone,
			&appointmentDate, &chiefComplaint, &status, &doctorID, &doctorName)
		if err != nil {
			continue
		}

		appointments = append(appointments, map[string]interface{}{
			"id":               id,
			"patient_id":       patientID,
			"patient_name":     patientName,
			"patient_phone":    phone,
			"appointment_date": appointmentDate,
			"chief_complaint":  chiefComplaint,
			"status":           status,
			"doctor_name":      doctorName,
			"doctor_id":        doctorID,
		})
	}

	return appointments, nil
}

// GetAppointmentsByFilter returns appointments filtered by date range
func (rs *ReceptionistService) GetAppointmentsByFilter(filter string, limit int, offset int) ([]map[string]interface{}, error) {
	var query string
	var rows *sql.Rows
	var err error

	switch filter {
	case "all":
		query = `
			SELECT 
				a.id,
				a.patient_id,
				COALESCE(u.name, 'Unknown') as patient_name,
				a.doctor_id,
				COALESCE(du.name, 'Unknown') as doctor_name,
				a.appointment_date,
				COALESCE(a.appointment_time::text, '00:00:00') as appointment_time,
				a.status::text,
				COALESCE(a.chief_complaint, '') as chief_complaint,
				COALESCE(a.bed_id::text, '') as bed_id
			FROM appointments a
			LEFT JOIN patients p ON a.patient_id = p.id
			LEFT JOIN users u ON p.user_id = u.id
			LEFT JOIN doctors d ON a.doctor_id = d.id
			LEFT JOIN users du ON d.user_id = du.id
			ORDER BY a.appointment_date DESC, a.appointment_time DESC
			LIMIT $1 OFFSET $2
		`
		rows, err = rs.db.Query(query, limit, offset)
	case "today":
		query = `
			SELECT 
				a.id,
				a.patient_id,
				COALESCE(u.name, 'Unknown') as patient_name,
				a.doctor_id,
				COALESCE(du.name, 'Unknown') as doctor_name,
				a.appointment_date,
				COALESCE(a.appointment_time::text, '00:00:00') as appointment_time,
				a.status::text,
				COALESCE(a.chief_complaint, '') as chief_complaint,
				COALESCE(a.bed_id::text, '') as bed_id
			FROM appointments a
			LEFT JOIN patients p ON a.patient_id = p.id
			LEFT JOIN users u ON p.user_id = u.id
			LEFT JOIN doctors d ON a.doctor_id = d.id
			LEFT JOIN users du ON d.user_id = du.id
			WHERE a.appointment_date = CURRENT_DATE
			ORDER BY a.appointment_time ASC
			LIMIT $1 OFFSET $2
		`
		rows, err = rs.db.Query(query, limit, offset)

	case "upcoming":
		query = `
			SELECT 
				a.id,
				a.patient_id,
				COALESCE(u.name, 'Unknown') as patient_name,
				a.doctor_id,
				COALESCE(du.name, 'Unknown') as doctor_name,
				a.appointment_date,
				COALESCE(a.appointment_time::text, '00:00:00') as appointment_time,
				a.status::text,
				COALESCE(a.chief_complaint, '') as chief_complaint,
				COALESCE(a.bed_id::text, '') as bed_id
			FROM appointments a
			LEFT JOIN patients p ON a.patient_id = p.id
			LEFT JOIN users u ON p.user_id = u.id
			LEFT JOIN doctors d ON a.doctor_id = d.id
			LEFT JOIN users du ON d.user_id = du.id
			WHERE a.appointment_date > CURRENT_DATE
			AND a.status::text != 'Cancelled'
			ORDER BY a.appointment_date ASC
			LIMIT $1 OFFSET $2
		`
		rows, err = rs.db.Query(query, limit, offset)

	case "past":
		query = `
			SELECT 
				a.id,
				a.patient_id,
				COALESCE(u.name, 'Unknown') as patient_name,
				a.doctor_id,
				COALESCE(du.name, 'Unknown') as doctor_name,
				a.appointment_date,
				COALESCE(a.appointment_time::text, '00:00:00') as appointment_time,
				a.status::text,
				COALESCE(a.chief_complaint, '') as chief_complaint,
				COALESCE(a.bed_id::text, '') as bed_id
			FROM appointments a
			LEFT JOIN patients p ON a.patient_id = p.id
			LEFT JOIN users u ON p.user_id = u.id
			LEFT JOIN doctors d ON a.doctor_id = d.id
			LEFT JOIN users du ON d.user_id = du.id
			WHERE a.appointment_date < CURRENT_DATE
			ORDER BY a.appointment_date DESC
			LIMIT $1 OFFSET $2
		`
		rows, err = rs.db.Query(query, limit, offset)

	default:
		return nil, errors.New("invalid filter: must be all, today, upcoming, or past")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to fetch appointments: %w", err)
	}
	defer rows.Close()

	var appointments []map[string]interface{}
	for rows.Next() {
		var id, patientID, patientName, doctorID, doctorName, status, chiefComplaint, bedID string
		var appointmentDate time.Time
		var appointmentTime string

		err := rows.Scan(&id, &patientID, &patientName, &doctorID, &doctorName,
			&appointmentDate, &appointmentTime, &status, &chiefComplaint, &bedID)
		if err != nil {
			continue
		}

		appointments = append(appointments, map[string]interface{}{
			"id":               id,
			"patient_id":       patientID,
			"patient_name":     patientName,
			"doctor_id":        doctorID,
			"doctor_name":      doctorName,
			"appointment_date": appointmentDate.Format("2006-01-02"),
			"appointment_time": appointmentTime,
			"status":           status,
			"chief_complaint":  chiefComplaint,
			"bed_id":           bedID,
		})
	}

	return appointments, nil
}

// GetAllPatients returns all registered patients with optional search
func (rs *ReceptionistService) GetAllPatients(search string, limit int, offset int) ([]map[string]interface{}, error) {
	query := `
		SELECT 
			p.id,
			u.name as patient_name,
			u.email,
			COALESCE(p.phone, '') as phone,
			p.date_of_birth,
			p.gender,
			p.blood_group,
			(SELECT COUNT(*) FROM appointments WHERE patient_id = p.id) as appointment_count
		FROM patients p
		JOIN users u ON p.user_id = u.id
	`

	var args []interface{}
	if search != "" {
		query += " WHERE u.name ILIKE $1 OR u.email ILIKE $1"
		args = append(args, "%"+search+"%")
		args = append(args, limit)
		args = append(args, offset)
		query += " ORDER BY u.created_at DESC LIMIT $2 OFFSET $3"
	} else {
		args = append(args, limit)
		args = append(args, offset)
		query += " ORDER BY u.created_at DESC LIMIT $1 OFFSET $2"
	}

	rows, err := rs.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch patients: %w", err)
	}
	defer rows.Close()

	var patients []map[string]interface{}
	for rows.Next() {
		var patientID, patientName, email, phone, gender, bloodGroup string
		var dateOfBirth time.Time
		var appointmentCount int

		err := rows.Scan(&patientID, &patientName, &email, &phone, &dateOfBirth, &gender, &bloodGroup, &appointmentCount)
		if err != nil {
			continue
		}

		patients = append(patients, map[string]interface{}{
			"id":                patientID,
			"name":              patientName,
			"email":             email,
			"phone":             phone,
			"date_of_birth":     dateOfBirth.Format("2006-01-02"),
			"gender":            gender,
			"blood_group":       bloodGroup,
			"appointment_count": appointmentCount,
		})
	}

	return patients, nil
}

// GetAllDoctorsWithSchedules returns all active doctors with their schedules
func (rs *ReceptionistService) GetAllDoctorsWithSchedules() ([]map[string]interface{}, error) {
	query := `
		SELECT 
			d.id,
			u.name as doctor_name,
			d.specialization,
			d.department,
			d.consultation_fee,
			d.license_number
		FROM doctors d
		JOIN users u ON d.user_id = u.id
		WHERE d.is_active = true
		ORDER BY u.name ASC
	`

	rows, err := rs.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch doctors: %w", err)
	}
	defer rows.Close()

	var doctors []map[string]interface{}
	for rows.Next() {
		var doctorID, doctorName, specialization, department, licenseNumber string
		var consultationFee float64

		err := rows.Scan(&doctorID, &doctorName, &specialization, &department, &consultationFee, &licenseNumber)
		if err != nil {
			continue
		}

		// Fetch doctor schedules
		scheduleQuery := `
			SELECT id, day_of_week, start_time, end_time, slot_duration_minutes
			FROM doctor_schedules
			WHERE doctor_id = $1 AND is_active = true
		`
		schedRows, err := rs.db.Query(scheduleQuery, doctorID)
		if err != nil {
			continue
		}

		var schedules []map[string]interface{}
		for schedRows.Next() {
			var schedID, startTime, endTime string
			var dayOfWeek, slotDuration int

			err := schedRows.Scan(&schedID, &dayOfWeek, &startTime, &endTime, &slotDuration)
			if err != nil {
				continue
			}

			schedules = append(schedules, map[string]interface{}{
				"id":                    schedID,
				"day_of_week":           dayOfWeek,
				"start_time":            startTime,
				"end_time":              endTime,
				"slot_duration_minutes": slotDuration,
			})
		}
		schedRows.Close()

		doctors = append(doctors, map[string]interface{}{
			"id":               doctorID,
			"name":             doctorName,
			"specialization":   specialization,
			"department":       department,
			"consultation_fee": consultationFee,
			"license_number":   licenseNumber,
			"schedules":        schedules,
		})
	}

	return doctors, nil
}

// GetAllBeds returns all beds organized by room and department
func (rs *ReceptionistService) GetAllBeds() ([]map[string]interface{}, error) {
	query := `
		SELECT 
			b.id,
			b.room_number,
			b.bed_number,
			b.department_id,
			COALESCE(d.name, 'Unknown') as department_name,
			COALESCE(d.type, 'room') as department_type,
			b.status::text as status,
			b.capacity,
			COALESCE(a.patient_id::text, '') as patient_id,
			COALESCE(u.name, '') as patient_name
		FROM beds b
		LEFT JOIN departments d ON b.department_id = d.id
		LEFT JOIN appointments a ON b.id = a.bed_id 
		LEFT JOIN patients p ON a.patient_id = p.id
		LEFT JOIN users u ON p.user_id = u.id
		ORDER BY b.room_number ASC, b.bed_number ASC
	`

	rows, err := rs.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch beds: %w", err)
	}
	defer rows.Close()

	var beds []map[string]interface{}
	for rows.Next() {
		var bedID, roomNumber, bedNumber, departmentID, departmentName, departmentType, status string
		var capacity int
		var patientID, patientName string

		err := rows.Scan(&bedID, &roomNumber, &bedNumber, &departmentID, &departmentName, &departmentType, &status, &capacity, &patientID, &patientName)
		if err != nil {
			continue
		}

		beds = append(beds, map[string]interface{}{
			"id":              bedID,
			"room_number":     roomNumber,
			"bed_number":      bedNumber,
			"department_id":   departmentID,
			"department_name": departmentName,
			"department_type": departmentType,
			"status":          status,
			"capacity":        capacity,
			"patient_id":      patientID,
			"patient_name":    patientName,
		})
	}

	return beds, nil
}

// ReassignBedToAppointment reassigns a patient from one bed to another
func (rs *ReceptionistService) ReassignBedToAppointment(appointmentID string, oldBedID string, newBedID string) error {
	tx, err := rs.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Verify appointment exists and is checked in
	var currentBedID string
	apptQuery := "SELECT bed_id FROM appointments WHERE id = $1 AND status IN ('Checked In', 'Ready for Doctor', 'In Consultation')"
	err = tx.QueryRow(apptQuery, appointmentID).Scan(&currentBedID)
	if err == sql.ErrNoRows {
		return errors.New("appointment not found or not in valid state for reassignment")
	}
	if err != nil {
		return fmt.Errorf("failed to verify appointment: %w", err)
	}

	// Verify new bed is available
	var newBedStatus string
	bedQuery := "SELECT status FROM beds WHERE id = $1"
	err = tx.QueryRow(bedQuery, newBedID).Scan(&newBedStatus)
	if err == sql.ErrNoRows {
		return errors.New("new bed not found")
	}
	if err != nil {
		return fmt.Errorf("failed to validate new bed: %w", err)
	}

	if newBedStatus != "available" {
		return fmt.Errorf("new bed is not available (status: %s)", newBedStatus)
	}

	// Update appointment with new bed
	updateApptQuery := "UPDATE appointments SET bed_id = $1 WHERE id = $2"
	_, err = tx.Exec(updateApptQuery, newBedID, appointmentID)
	if err != nil {
		return fmt.Errorf("failed to update appointment: %w", err)
	}

	// Release old bed
	releaseOldQuery := "UPDATE beds SET status = 'available' WHERE id = $1"
	_, err = tx.Exec(releaseOldQuery, oldBedID)
	if err != nil {
		return fmt.Errorf("failed to release old bed: %w", err)
	}

	// Occupy new bed
	occupyNewQuery := "UPDATE beds SET status = 'occupied' WHERE id = $1"
	_, err = tx.Exec(occupyNewQuery, newBedID)
	if err != nil {
		return fmt.Errorf("failed to occupy new bed: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// CreateNurseTask creates a new nurse task record
func (rs *ReceptionistService) CreateNurseTask(appointmentID string, patientID string, taskType string, description string, nurseID string, receptionistID string, priority string, dueBy string) (string, error) {
	// Validate nurse
	if nurseID != "" {
		var nurseExists int
		err := rs.db.QueryRow("SELECT 1 FROM users WHERE id = $1 AND role = 'Nurse'", nurseID).Scan(&nurseExists)
		if err == sql.ErrNoRows {
			return "", errors.New("nurse not found or invalid role")
		}
		if err != nil {
			return "", fmt.Errorf("failed to validate nurse: %w", err)
		}
	}

	// Validate appointment if provided
	if appointmentID != "" {
		var apptExists int
		err := rs.db.QueryRow("SELECT 1 FROM appointments WHERE id = $1", appointmentID).Scan(&apptExists)
		if err == sql.ErrNoRows {
			return "", errors.New("appointment not found")
		}
		if err != nil {
			return "", fmt.Errorf("failed to validate appointment: %w", err)
		}
	}

	// Validate patient if provided
	if patientID != "" {
		var patientExists int
		err := rs.db.QueryRow("SELECT 1 FROM patients WHERE id = $1", patientID).Scan(&patientExists)
		if err == sql.ErrNoRows {
			return "", errors.New("patient not found")
		}
		if err != nil {
			return "", fmt.Errorf("failed to validate patient: %w", err)
		}
	}

	// Generate task ID
	taskID := fmt.Sprintf("task_%d", time.Now().UnixNano())

	// Store in audit_logs with task details
	insertQuery := `
		INSERT INTO audit_logs (user_id, user_name, action, resource_type, resource_id, details, created_at)
		VALUES ($1, '', 'CREATE_NURSE_TASK', 'nurse_task', $2, $3, NOW())
	`

	// Build JSON details with all fields including assigned nurse
	detailsJSON := fmt.Sprintf(`{"appointment_id": "%s", "patient_id": "%s", "nurse_id": "%s", "task_type": "%s", "description": "%s", "priority": "%s", "due_by": "%s"}`,
		appointmentID, patientID, nurseID, taskType, description, priority, dueBy)

	_, err := rs.db.Exec(insertQuery, receptionistID, taskID, detailsJSON)
	if err != nil {
		return "", fmt.Errorf("failed to create nurse task: %w", err)
	}

	return taskID, nil
}

// GetNurseTasks returns pending nurse tasks
func (rs *ReceptionistService) GetNurseTasks() ([]map[string]interface{}, error) {
	query := `
		SELECT 
			a.id as task_id,
			a.resource_id,
			a.details,
			a.created_at
		FROM audit_logs a
		WHERE a.action = 'CREATE_NURSE_TASK'
		AND a.created_at > NOW() - INTERVAL '7 days'
		ORDER BY a.created_at DESC
		LIMIT 100
	`

	rows, err := rs.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch nurse tasks: %w", err)
	}
	defer rows.Close()

	var tasks []map[string]interface{}
	for rows.Next() {
		var taskID, resourceID string
		var createdAt time.Time
		var details string

		err := rows.Scan(&taskID, &resourceID, &details, &createdAt)
		if err != nil {
			continue
		}

		tasks = append(tasks, map[string]interface{}{
			"id":         resourceID,
			"task_id":    taskID,
			"details":    details,
			"created_at": createdAt,
		})
	}

	return tasks, nil
}

// UpdateNurseTask updates nurse task details in the audit log
func (rs *ReceptionistService) UpdateNurseTask(taskID, taskType, patientID, description, priority, dueBy string) error {
	// First retrieve the existing task to get the current details
	var details string
	err := rs.db.QueryRow(`
		SELECT details FROM audit_logs 
		WHERE resource_id = $1 AND action = 'CREATE_NURSE_TASK'
		LIMIT 1
	`, taskID).Scan(&details)

	if err == sql.ErrNoRows {
		return fmt.Errorf("task not found")
	}
	if err != nil {
		return fmt.Errorf("failed to fetch task: %w", err)
	}

	// Parse existing details
	var taskDetails map[string]interface{}
	if err := json.Unmarshal([]byte(details), &taskDetails); err != nil {
		taskDetails = make(map[string]interface{})
	}

	// Update the fields
	if taskType != "" {
		taskDetails["task_type"] = taskType
	}
	if patientID != "" {
		taskDetails["patient_id"] = patientID
	}
	if description != "" {
		taskDetails["description"] = description
	}
	if priority != "" {
		taskDetails["priority"] = priority
	}
	if dueBy != "" {
		taskDetails["due_by"] = dueBy
	}

	// Convert back to JSON
	updatedDetails, err := json.Marshal(taskDetails)
	if err != nil {
		return fmt.Errorf("failed to marshal task details: %w", err)
	}

	// Update the audit log
	_, err = rs.db.Exec(`
		UPDATE audit_logs 
		SET details = $1
		WHERE resource_id = $2 AND action = 'CREATE_NURSE_TASK'
	`, string(updatedDetails), taskID)

	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	return nil
}

// UpdateTaskStatus updates the status of a nurse task
func (rs *ReceptionistService) UpdateTaskStatus(taskID, status string) error {
	// First retrieve the existing task to get the current details
	var details string
	err := rs.db.QueryRow(`
		SELECT details FROM audit_logs 
		WHERE resource_id = $1 AND action = 'CREATE_NURSE_TASK'
		LIMIT 1
	`, taskID).Scan(&details)

	if err == sql.ErrNoRows {
		return fmt.Errorf("task not found")
	}
	if err != nil {
		return fmt.Errorf("failed to fetch task: %w", err)
	}

	// Parse existing details
	var taskDetails map[string]interface{}
	if err := json.Unmarshal([]byte(details), &taskDetails); err != nil {
		taskDetails = make(map[string]interface{})
	}

	// Update the status field
	taskDetails["status"] = status
	taskDetails["updated_at"] = time.Now().Format(time.RFC3339)

	// Convert back to JSON
	updatedDetails, err := json.Marshal(taskDetails)
	if err != nil {
		return fmt.Errorf("failed to marshal task details: %w", err)
	}

	// Update the audit log
	_, err = rs.db.Exec(`
		UPDATE audit_logs 
		SET details = $1
		WHERE resource_id = $2 AND action = 'CREATE_NURSE_TASK'
	`, string(updatedDetails), taskID)

	if err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}

	return nil
}

// DischargeBedPatient discharges a patient from a bed
func (rs *ReceptionistService) DischargeBedPatient(bedID string) error {
	// Update bed status to available
	_, err := rs.db.Exec(`
		UPDATE beds 
		SET status = 'available'
		WHERE id = $1
	`, bedID)

	if err != nil {
		return fmt.Errorf("failed to discharge patient from bed: %w", err)
	}

	// Update associated appointment status to completed
	_, err = rs.db.Exec(`
		UPDATE appointments 
		SET status = 'Completed'
		WHERE bed_id = $1 AND status != 'Completed'
	`, bedID)

	if err != nil {
		return fmt.Errorf("failed to update appointment status: %w", err)
	}

	return nil
}

// GetBedDischargeInfo returns information about a patient in a bed for discharge
func (rs *ReceptionistService) GetBedDischargeInfo(bedID string) (map[string]interface{}, error) {
	query := `
		SELECT 
			b.id as bed_id,
			b.room_id,
			b.bed_number,
			b.status,
			a.id as appointment_id,
			p.id as patient_id,
			u.name as patient_name,
			COALESCE(a.status, 'N/A') as appointment_status,
			COALESCE(a.check_in_time::text, '') as check_in_time
		FROM beds b
		LEFT JOIN appointments a ON b.id = a.bed_id
		LEFT JOIN patients p ON a.patient_id = p.id
		LEFT JOIN users u ON p.user_id = u.id
		WHERE b.id = $1
	`

	var bedID_, roomID, bedNumber, status, appointmentID, patientID, patientName, appointmentStatus, checkInTime string

	row := rs.db.QueryRow(query, bedID)
	err := row.Scan(&bedID_, &roomID, &bedNumber, &status, &appointmentID, &patientID, &patientName, &appointmentStatus, &checkInTime)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("bed not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch bed information: %w", err)
	}

	return map[string]interface{}{
		"bed_id":             bedID_,
		"room_id":            roomID,
		"bed_number":         bedNumber,
		"status":             status,
		"appointment_id":     appointmentID,
		"patient_id":         patientID,
		"patient_name":       patientName,
		"appointment_status": appointmentStatus,
		"check_in_time":      checkInTime,
	}, nil
}

// DeleteTask deletes a nurse task
func (rs *ReceptionistService) DeleteTask(taskID string) error {
	query := `
		DELETE FROM audit_logs 
		WHERE resource_id = $1 AND action = 'CREATE_NURSE_TASK'
		RETURNING id
	`

	var id string
	err := rs.db.QueryRow(query, taskID).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("task not found")
		}
		return fmt.Errorf("failed to delete task: %w", err)
	}

	return nil
}

// Helper functions

func createUserWithRole(tx *sql.Tx, user models.User, roleName string, password string) (userID string, err error) {
	// Hash password using bcrypt
	var passwordHash string
	if password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return "", fmt.Errorf("failed to hash password: %w", err)
		}
		passwordHash = string(hashedPassword)
	}

	createUserQuery := `
		INSERT INTO users (email, name, password_hash, is_active, created_at)
		VALUES ($1, $2, $3, true, NOW())
		RETURNING id
	`
	err = tx.QueryRow(createUserQuery, user.Email, user.Name, passwordHash).Scan(&userID)
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
		INSERT INTO patients (user_id, date_of_birth, gender, blood_group, phone, created_at)
		VALUES ($1, $2::date, $3, $4, $5, NOW())
		RETURNING id
	`
	err = tx.QueryRow(createQuery, userID, input.DateOfBirth, input.Gender, input.BloodGroup, input.Phone).Scan(&patientID)
	if err != nil {
		return "", fmt.Errorf("failed to create patient: %w", err)
	}
	return patientID, nil
}

func addPatientAllergies(tx *sql.Tx, patientID string, allergies []string) error {
	for _, allergy := range allergies {
		query := `
			INSERT INTO patient_allergies (patient_id, allergy)
			VALUES ($1, $2)
			ON CONFLICT DO NOTHING
		`
		_, err := tx.Exec(query, patientID, allergy)
		if err != nil {
			return fmt.Errorf("failed to add allergy: %w", err)
		}
	}
	return nil
}

func addPatientConditions(tx *sql.Tx, patientID string, conditions []string) error {
	for _, condition := range conditions {
		query := `
			INSERT INTO patient_chronic_conditions (patient_id, condition)
			VALUES ($1, $2)
			ON CONFLICT DO NOTHING
		`
		_, err := tx.Exec(query, patientID, condition)
		if err != nil {
			return fmt.Errorf("failed to add condition: %w", err)
		}
	}
	return nil
}

func addPatientEmergencyContacts(tx *sql.Tx, patientID string, contacts []EmergencyContactInput) error {
	for _, contact := range contacts {
		query := `
			INSERT INTO patient_emergency_contacts (patient_id, name, relation, phone)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT DO NOTHING
		`
		_, err := tx.Exec(query, patientID, contact.Name, contact.Relation, contact.Phone)
		if err != nil {
			return fmt.Errorf("failed to add emergency contact: %w", err)
		}
	}
	return nil
}

// GetProfile retrieves the receptionist profile
func (rs *ReceptionistService) GetProfile(userID string) (map[string]interface{}, error) {
	var profile map[string]interface{} = make(map[string]interface{})
	var id, email, name string
	var createdAt time.Time

	// Query with just the guaranteed columns
	profileQuery := `
		SELECT id, email, name, created_at 
		FROM users 
		WHERE id = $1
	`

	err := rs.db.QueryRow(profileQuery, userID).Scan(&id, &email, &name, &createdAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}

	// Try to get optional fields if they exist
	var phone, address, bio string
	optionalQuery := `
		SELECT COALESCE(phone, ''), COALESCE(address, ''), COALESCE(bio, '')
		FROM users 
		WHERE id = $1
	`

	_ = rs.db.QueryRow(optionalQuery, userID).Scan(&phone, &address, &bio)
	// If this fails, we just use the defaults (empty strings)

	profile["id"] = id
	profile["email"] = email
	profile["name"] = name
	profile["phone"] = phone
	profile["address"] = address
	profile["bio"] = bio
	profile["created_at"] = createdAt
	profile["role"] = "Receptionist"
	profile["department"] = "Reception"
	profile["status"] = "Active"

	return profile, nil
}

// UpdateProfile updates the receptionist profile
func (rs *ReceptionistService) UpdateProfile(userID string, updateData map[string]interface{}) (map[string]interface{}, error) {
	// Only update fields that we know exist in the base users table
	// name, email are guaranteed to exist. phone and address may exist from migrations

	var updateFields []string
	var args []interface{}
	argCounter := 1

	if name, ok := updateData["name"].(string); ok && name != "" {
		updateFields = append(updateFields, fmt.Sprintf("name = $%d", argCounter))
		args = append(args, name)
		argCounter++
	}

	if email, ok := updateData["email"].(string); ok && email != "" {
		updateFields = append(updateFields, fmt.Sprintf("email = $%d", argCounter))
		args = append(args, email)
		argCounter++
	}

	// Try to update optional columns if they exist, but don't fail if they don't
	if phone, ok := updateData["phone"].(string); ok && phone != "" {
		updateFields = append(updateFields, fmt.Sprintf("phone = $%d", argCounter))
		args = append(args, phone)
		argCounter++
	}

	if address, ok := updateData["address"].(string); ok && address != "" {
		updateFields = append(updateFields, fmt.Sprintf("address = $%d", argCounter))
		args = append(args, address)
		argCounter++
	}

	// bio is nice-to-have, try to include but don't break if it doesn't exist
	if bio, ok := updateData["bio"].(string); ok && bio != "" {
		updateFields = append(updateFields, fmt.Sprintf("bio = $%d", argCounter))
		args = append(args, bio)
		argCounter++
	}

	// If no fields to update, just return current profile
	if len(updateFields) == 0 {
		return rs.GetProfile(userID)
	}

	// Build and execute update query
	query := fmt.Sprintf(`
		UPDATE users
		SET %s
		WHERE id = $%d
	`, strings.Join(updateFields, ", "), argCounter)

	args = append(args, userID)

	_, err := rs.db.Exec(query, args...)
	if err != nil {
		// Log the error but don't fail - try to return the current profile instead
		fmt.Printf("WARNING: Failed to update all profile fields: %v. Attempting fallback...\n", err)

		// Try updating just name and email which should always exist
		nameEmailQuery := `
			UPDATE users
			SET name = COALESCE($1, name), email = COALESCE($2, email)
			WHERE id = $3
		`
		fallbackName := updateData["name"]
		fallbackEmail := updateData["email"]
		if fallbackName != nil || fallbackEmail != nil {
			rs.db.Exec(nameEmailQuery, fallbackName, fallbackEmail, userID)
		}
	}

	// Return updated profile
	return rs.GetProfile(userID)
}
