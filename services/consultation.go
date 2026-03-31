package services

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
)

// ConsultationDetails contains consultation with related data
type ConsultationDetails struct {
	ID                  string    `json:"id"`
	AppointmentID       string    `json:"appointmentId"`
	DoctorID            string    `json:"doctorId"`
	DoctorName          string    `json:"doctorName"`
	PatientID           string    `json:"patientId"`
	PatientName         string    `json:"patientName"`
	Diagnosis           string    `json:"diagnosis"`
	DoctorNotes         string    `json:"doctorNotes"`
	PatientInstructions string    `json:"patientInstructions"`
	Investigations      []string  `json:"investigations"`
	FollowUpDate        *string   `json:"followUpDate"`
	ReferralDoctorID    *string   `json:"referralDoctorId"`
	ReferralDoctorName  *string   `json:"referralDoctorName"`
	CreatedAt           time.Time `json:"createdAt"`
	UpdatedAt           time.Time `json:"updatedAt"`
}

// CreateConsultationRequest contains consultation data
type CreateConsultationRequest struct {
	Diagnosis           string   `json:"diagnosis" binding:"required"`
	DoctorNotes         string   `json:"doctorNotes"`
	PatientInstructions string   `json:"patientInstructions"`
	Investigations      []string `json:"investigations"`
	FollowUpDate        *string  `json:"followUpDate"`
	ReferralDoctorID    *string  `json:"referralDoctorId"`
}

// CreateConsultation creates a new consultation
func CreateConsultation(db *sql.DB, appointmentID, doctorID, patientID string, req CreateConsultationRequest) (*ConsultationDetails, error) {
	id := uuid.New().String()
	now := time.Now()

	// Get names
	var doctorName, patientName string
	err := db.QueryRow(`
		SELECT u.name FROM users u
		JOIN doctors d ON u.id = d.user_id
		WHERE d.id = $1
	`, doctorID).Scan(&doctorName)
	if err != nil {
		log.Println("Error getting doctor:", err)
		return nil, fmt.Errorf("failed to get doctor")
	}

	err = db.QueryRow(`
		SELECT u.name FROM users u
		JOIN patients p ON u.id = p.user_id
		WHERE p.id = $1
	`, patientID).Scan(&patientName)
	if err != nil {
		log.Println("Error getting patient:", err)
		return nil, fmt.Errorf("failed to get patient")
	}

	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		log.Println("Error starting transaction:", err)
		return nil, fmt.Errorf("transaction error")
	}
	defer tx.Rollback()

	// Create consultation
	_, err = tx.Exec(`
		INSERT INTO consultations 
		(id, appointment_id, doctor_id, patient_id, diagnosis, doctor_notes, patient_instructions, follow_up_date, referral_doctor_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`, id, appointmentID, doctorID, patientID, req.Diagnosis, req.DoctorNotes, req.PatientInstructions, req.FollowUpDate, req.ReferralDoctorID, now, now)

	if err != nil {
		log.Println("Error creating consultation:", err)
		return nil, fmt.Errorf("failed to create consultation")
	}

	// Add investigations
	for _, inv := range req.Investigations {
		_, err = tx.Exec(`
			INSERT INTO consultation_investigations (consultation_id, investigation)
			VALUES ($1, $2)
		`, id, inv)
		if err != nil {
			log.Println("Error creating investigation:", err)
			tx.Rollback()
			return nil, fmt.Errorf("failed to add investigation")
		}
	}

	// Update appointment status to "In Consultation"
	_, err = tx.Exec(`
		UPDATE appointments 
		SET status = 'In Consultation', updated_at = $2
		WHERE id = $1
	`, appointmentID, now)
	if err != nil {
		log.Println("Error updating appointment:", err)
		tx.Rollback()
		return nil, fmt.Errorf("failed to update appointment")
	}

	if err = tx.Commit(); err != nil {
		log.Println("Error committing transaction:", err)
		return nil, fmt.Errorf("transaction commit failed")
	}

	detail := &ConsultationDetails{
		ID:                  id,
		AppointmentID:       appointmentID,
		DoctorID:            doctorID,
		DoctorName:          doctorName,
		PatientID:           patientID,
		PatientName:         patientName,
		Diagnosis:           req.Diagnosis,
		DoctorNotes:         req.DoctorNotes,
		PatientInstructions: req.PatientInstructions,
		Investigations:      req.Investigations,
		FollowUpDate:        req.FollowUpDate,
		ReferralDoctorID:    req.ReferralDoctorID,
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	return detail, nil
}

// GetConsultationByAppointmentID gets consultation for an appointment
func GetConsultationByAppointmentID(db *sql.DB, appointmentID string) (*ConsultationDetails, error) {
	var id string
	err := db.QueryRow(`
		SELECT id FROM consultations WHERE appointment_id = $1
	`, appointmentID).Scan(&id)

	if err == sql.ErrNoRows {
		return nil, nil // No consultation yet
	}
	if err != nil {
		log.Println("Error querying consultation:", err)
		return nil, fmt.Errorf("failed to get consultation")
	}

	return GetConsultationByID(db, id)
}

// GetConsultationByID gets a consultation by ID
func GetConsultationByID(db *sql.DB, consultationID string) (*ConsultationDetails, error) {
	detail := &ConsultationDetails{}

	err := db.QueryRow(`
		SELECT c.id, c.appointment_id, c.doctor_id, u1.name, c.patient_id, u2.name,
		       c.diagnosis, c.doctor_notes, c.patient_instructions, c.follow_up_date, c.referral_doctor_id,
		       c.created_at, c.updated_at
		FROM consultations c
		JOIN doctors d ON c.doctor_id = d.id
		JOIN users u1 ON d.user_id = u1.id
		JOIN patients p ON c.patient_id = p.id
		JOIN users u2 ON p.user_id = u2.id
		WHERE c.id = $1
	`, consultationID).Scan(
		&detail.ID, &detail.AppointmentID, &detail.DoctorID, &detail.DoctorName,
		&detail.PatientID, &detail.PatientName, &detail.Diagnosis, &detail.DoctorNotes,
		&detail.PatientInstructions, &detail.FollowUpDate, &detail.ReferralDoctorID,
		&detail.CreatedAt, &detail.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("consultation not found")
	}
	if err != nil {
		log.Println("Error querying consultation:", err)
		return nil, fmt.Errorf("failed to get consultation")
	}

	// Get investigations
	rows, err := db.Query(`
		SELECT investigation FROM consultation_investigations WHERE consultation_id = $1
	`, consultationID)
	if err != nil {
		log.Println("Error querying investigations:", err)
		return nil, fmt.Errorf("failed to get investigations")
	}
	defer rows.Close()

	var investigations []string
	for rows.Next() {
		var inv string
		if err := rows.Scan(&inv); err != nil {
			log.Println("Error scanning investigation:", err)
			continue
		}
		investigations = append(investigations, inv)
	}

	detail.Investigations = investigations

	// Get referral doctor name if exists
	if detail.ReferralDoctorID != nil {
		var refName string
		err := db.QueryRow(`
			SELECT u.name FROM users u
			JOIN doctors d ON u.id = d.user_id
			WHERE d.id = $1
		`, *detail.ReferralDoctorID).Scan(&refName)
		if err == nil {
			detail.ReferralDoctorName = &refName
		}
	}

	return detail, nil
}

// GetConsultationHistory gets all consultations for a patient
func GetConsultationHistory(db *sql.DB, patientID string, limit, offset int) ([]ConsultationDetails, int64, error) {
	var count int64
	err := db.QueryRow(`
		SELECT COUNT(*) FROM consultations WHERE patient_id = $1
	`, patientID).Scan(&count)
	if err != nil {
		log.Println("Error counting consultations:", err)
		return nil, 0, err
	}

	rows, err := db.Query(`
		SELECT c.id, c.appointment_id, c.doctor_id, u1.name, c.patient_id, u2.name,
		       c.diagnosis, c.doctor_notes, c.patient_instructions, c.follow_up_date, c.referral_doctor_id,
		       c.created_at, c.updated_at
		FROM consultations c
		JOIN doctors d ON c.doctor_id = d.id
		JOIN users u1 ON d.user_id = u1.id
		JOIN patients p ON c.patient_id = p.id
		JOIN users u2 ON p.user_id = u2.id
		WHERE c.patient_id = $1
		ORDER BY c.created_at DESC
		LIMIT $2 OFFSET $3
	`, patientID, limit, offset)
	if err != nil {
		log.Println("Error querying consultations:", err)
		return nil, 0, err
	}
	defer rows.Close()

	var consultations []ConsultationDetails
	for rows.Next() {
		detail := ConsultationDetails{}
		err := rows.Scan(
			&detail.ID, &detail.AppointmentID, &detail.DoctorID, &detail.DoctorName,
			&detail.PatientID, &detail.PatientName, &detail.Diagnosis, &detail.DoctorNotes,
			&detail.PatientInstructions, &detail.FollowUpDate, &detail.ReferralDoctorID,
			&detail.CreatedAt, &detail.UpdatedAt,
		)
		if err != nil {
			log.Println("Error scanning consultation:", err)
			continue
		}

		// Get investigations
		invRows, err := db.Query(`
			SELECT investigation FROM consultation_investigations WHERE consultation_id = $1
		`, detail.ID)
		if err == nil {
			var investigations []string
			for invRows.Next() {
				var inv string
				if err := invRows.Scan(&inv); err == nil {
					investigations = append(investigations, inv)
				}
			}
			invRows.Close()
			detail.Investigations = investigations
		}

		consultations = append(consultations, detail)
	}

	return consultations, count, nil
}
