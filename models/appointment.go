package models

import (
	"database/sql"
	"time"
)

// Appointment represents a doctor appointment
type Appointment struct {
	ID                 string         `json:"id"`
	PatientID          string         `json:"patientId"`
	PatientName        string         `json:"patientName"`
	DoctorID           string         `json:"doctorId"`
	DoctorName         string         `json:"doctorName"`
	DoctorSpecialty    string         `json:"doctorSpecialty"`
	AppointmentDate    string         `json:"appointmentDate"`
	AppointmentTime    string         `json:"appointmentTime"`
	DurationMinutes    int            `json:"durationMinutes"`
	Type               string         `json:"type"`     // Regular, Emergency, Follow-up
	Status             string         `json:"status"`   // Scheduled, In Progress, Completed, Cancelled
	Priority           string         `json:"priority"` // Routine, Urgent
	ChiefComplaint     sql.NullString `json:"chiefComplaint"`
	CancellationReason sql.NullString `json:"cancellationReason"`
	CreatedAt          time.Time      `json:"createdAt"`
	UpdatedAt          time.Time      `json:"updatedAt"`
}

type AppointmentRequest struct {
	PatientID       string `json:"patientId" binding:"required"`
	DoctorID        string `json:"doctorId" binding:"required"`
	AppointmentDate string `json:"appointmentDate" binding:"required"`
	AppointmentTime string `json:"appointmentTime" binding:"required"`
	DurationMinutes int    `json:"durationMinutes"`
	Type            string `json:"type"`
	Priority        string `json:"priority"`
	ChiefComplaint  string `json:"chiefComplaint"`
}

type AppointmentUpdateRequest struct {
	Status             string `json:"status"`
	CancellationReason string `json:"cancellationReason"`
}

type AppointmentResponse struct {
	ID              string    `json:"id"`
	PatientName     string    `json:"patientName"`
	DoctorName      string    `json:"doctorName"`
	AppointmentDate string    `json:"appointmentDate"`
	AppointmentTime string    `json:"appointmentTime"`
	Type            string    `json:"type"`
	Status          string    `json:"status"`
	Priority        string    `json:"priority"`
	ChiefComplaint  string    `json:"chiefComplaint"`
	CreatedAt       time.Time `json:"createdAt"`
}

type AppointmentListResponse struct {
	Appointments []AppointmentResponse `json:"appointments"`
	Total        int                   `json:"total"`
}

// Consultation represents doctor's diagnosis and notes from an appointment
type Consultation struct {
	ID                  string         `json:"id"`
	AppointmentID       string         `json:"appointmentId"`
	DoctorID            string         `json:"doctorId"`
	PatientID           string         `json:"patientId"`
	Diagnosis           string         `json:"diagnosis"`
	DoctorNotes         sql.NullString `json:"doctorNotes"`
	PatientInstructions sql.NullString `json:"patientInstructions"`
	FollowUpDate        sql.NullString `json:"followUpDate"`
	ReferralDoctorID    sql.NullString `json:"referralDoctorId"`
	Investigations      []string       `json:"investigations"`
	CreatedAt           time.Time      `json:"createdAt"`
	UpdatedAt           time.Time      `json:"updatedAt"`
}

type ConsultationRequest struct {
	Diagnosis           string   `json:"diagnosis" binding:"required"`
	DoctorNotes         string   `json:"doctorNotes"`
	PatientInstructions string   `json:"patientInstructions"`
	FollowUpDate        string   `json:"followUpDate"`
	ReferralDoctorID    string   `json:"referralDoctorId"`
	Investigations      []string `json:"investigations"`
}

type ConsultationResponse struct {
	ID                  string    `json:"id"`
	AppointmentID       string    `json:"appointmentId"`
	Diagnosis           string    `json:"diagnosis"`
	DoctorNotes         string    `json:"doctorNotes"`
	PatientInstructions string    `json:"patientInstructions"`
	FollowUpDate        string    `json:"followUpDate"`
	Investigations      []string  `json:"investigations"`
	CreatedAt           time.Time `json:"createdAt"`
}

// Prescription represents a medicine prescription for a patient
type Prescription struct {
	ID             string             `json:"id"`
	ConsultationID string             `json:"consultationId"`
	PatientID      string             `json:"patientId"`
	PatientName    string             `json:"patientName"`
	DoctorID       string             `json:"doctorId"`
	DoctorName     string             `json:"doctorName"`
	PharmacyStatus string             `json:"pharmacyStatus"` // Pending, Dispensed, Completed
	Items          []PrescriptionItem `json:"items"`
	DispensedAt    sql.NullTime       `json:"dispensedAt"`
	DispensedByID  sql.NullString     `json:"dispensedByUserId"`
	CreatedAt      time.Time          `json:"createdAt"`
}

type PrescriptionItem struct {
	ID                  string `json:"id"`
	PrescriptionID      string `json:"prescriptionId"`
	MedicineID          string `json:"medicineId"`
	MedicineName        string `json:"medicineName"`
	Strength            string `json:"strength"`
	DosageForm          string `json:"dosageForm"`
	Frequency           string `json:"frequency"` // Once daily, Twice daily, etc
	Timing              string `json:"timing"`    // Morning, Evening, Night
	Duration            string `json:"duration"`  // 7 days, 2 weeks, etc
	SpecialInstructions string `json:"specialInstructions"`
	QuantityDispensed   int    `json:"quantityDispensed"`
}

type PrescriptionRequest struct {
	ConsultationID string                    `json:"consultationId" binding:"required"`
	Items          []PrescriptionItemRequest `json:"items" binding:"required"`
}

type PrescriptionItemRequest struct {
	MedicineID          string `json:"medicineId" binding:"required"`
	MedicineName        string `json:"medicineName" binding:"required"`
	Strength            string `json:"strength" binding:"required"`
	DosageForm          string `json:"dosageForm" binding:"required"`
	Frequency           string `json:"frequency" binding:"required"`
	Timing              string `json:"timing" binding:"required"`
	Duration            string `json:"duration" binding:"required"`
	SpecialInstructions string `json:"specialInstructions"`
}

type PrescriptionResponse struct {
	ID             string             `json:"id"`
	PatientName    string             `json:"patientName"`
	DoctorName     string             `json:"doctorName"`
	Items          []PrescriptionItem `json:"items"`
	PharmacyStatus string             `json:"pharmacyStatus"`
	CreatedAt      time.Time          `json:"createdAt"`
}

type PrescriptionListResponse struct {
	Prescriptions []PrescriptionResponse `json:"prescriptions"`
	Total         int                    `json:"total"`
}
