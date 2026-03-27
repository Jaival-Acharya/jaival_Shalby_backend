package models

import (
	"database/sql"
	"time"
)

// Patient represents a patient in the hospital
type Patient struct {
	ID                    string                    `json:"id"`
	UserID                string                    `json:"userId"`
	Name                  string                    `json:"name"`
	Email                 string                    `json:"email"`
	DateOfBirth           string                    `json:"dateOfBirth"`
	Gender                string                    `json:"gender"`
	BloodGroup            string                    `json:"bloodGroup"`
	Phone                 string                    `json:"phone"`
	InsuranceProvider     sql.NullString            `json:"insuranceProvider"`
	InsurancePolicyNumber sql.NullString            `json:"insurancePolicyNumber"`
	Allergies             []PatientAllergy          `json:"allergies"`
	ChronicConditions     []PatientChronicCondition `json:"chronicConditions"`
	EmergencyContacts     []PatientEmergencyContact `json:"emergencyContacts"`
	CreatedAt             time.Time                 `json:"createdAt"`
}

type PatientAllergy struct {
	ID        int    `json:"id"`
	PatientID string `json:"patientId"`
	Allergy   string `json:"allergy"`
}

type PatientChronicCondition struct {
	ID        int    `json:"id"`
	PatientID string `json:"patientId"`
	Condition string `json:"condition"`
}

type PatientEmergencyContact struct {
	ID        string `json:"id"`
	PatientID string `json:"patientId"`
	Name      string `json:"name"`
	Phone     string `json:"phone"`
	Relation  string `json:"relation"`
}

type PatientVital struct {
	ID                     string    `json:"id"`
	PatientID              string    `json:"patientId"`
	RecordedAt             time.Time `json:"recordedAt"`
	RecordedByID           string    `json:"recordedByUserId"`
	RecordedByName         string    `json:"recordedByName"`
	HeightCm               float64   `json:"heightCm"`
	WeightKg               float64   `json:"weightKg"`
	BMI                    float64   `json:"bmi"`
	BloodPressureSystolic  int       `json:"bloodPressureSystolic"`
	BloodPressureDiastolic int       `json:"bloodPressureDiastolic"`
	BloodSugarMgDL         float64   `json:"bloodSugarMgDl"`
	TemperatureCelsius     float64   `json:"temperatureCelsius"`
	PulseBPM               int       `json:"pulseBpm"`
}

type PatientVitalRequest struct {
	HeightCm               float64 `json:"heightCm"`
	WeightKg               float64 `json:"weightKg"`
	BloodPressureSystolic  int     `json:"bloodPressureSystolic"`
	BloodPressureDiastolic int     `json:"bloodPressureDiastolic"`
	BloodSugarMgDL         float64 `json:"bloodSugarMgDl"`
	TemperatureCelsius     float64 `json:"temperatureCelsius"`
	PulseBPM               int     `json:"pulseBpm"`
}

type PatientResponse struct {
	ID                    string    `json:"id"`
	Name                  string    `json:"name"`
	Email                 string    `json:"email"`
	DateOfBirth           string    `json:"dateOfBirth"`
	Gender                string    `json:"gender"`
	BloodGroup            string    `json:"bloodGroup"`
	Phone                 string    `json:"phone"`
	InsuranceProvider     string    `json:"insuranceProvider"`
	InsurancePolicyNumber string    `json:"insurancePolicyNumber"`
	CreatedAt             time.Time `json:"createdAt"`
}

type PatientDetailResponse struct {
	ID                    string                    `json:"id"`
	Name                  string                    `json:"name"`
	Email                 string                    `json:"email"`
	DateOfBirth           string                    `json:"dateOfBirth"`
	Gender                string                    `json:"gender"`
	BloodGroup            string                    `json:"bloodGroup"`
	Phone                 string                    `json:"phone"`
	InsuranceProvider     string                    `json:"insuranceProvider"`
	InsurancePolicyNumber string                    `json:"insurancePolicyNumber"`
	Allergies             []PatientAllergy          `json:"allergies"`
	ChronicConditions     []PatientChronicCondition `json:"chronicConditions"`
	EmergencyContacts     []PatientEmergencyContact `json:"emergencyContacts"`
	LatestVital           *PatientVital             `json:"latestVital"`
	CreatedAt             time.Time                 `json:"createdAt"`
}

type PatientListResponse struct {
	Patients []PatientResponse `json:"patients"`
	Total    int               `json:"total"`
}

// PatientCreateRequest for creating new patient (admin form)
type PatientCreateRequest struct {
	Name                     string   `json:"name" binding:"required"`
	Email                    string   `json:"email" binding:"required,email"`
	Password                 string   `json:"password"` // Optional - auto-generated if not provided
	DateOfBirth              string   `json:"dateOfBirth" binding:"required"`
	Gender                   string   `json:"gender" binding:"required"`
	BloodGroup               string   `json:"bloodGroup" binding:"required"`
	Phone                    string   `json:"phone"`
	InsuranceProvider        string   `json:"insuranceProvider"`
	InsurancePolicyNumber    string   `json:"insurancePolicyNumber"`
	Allergies                []string `json:"allergies"`  // array of allergy strings
	Conditions               []string `json:"conditions"` // array of condition strings
	EmergencyContactName     string   `json:"emergencyContactName"`
	EmergencyContactPhone    string   `json:"emergencyContactPhone"`
	EmergencyContactRelation string   `json:"emergencyContactRelation"`
}

type PatientUpdateRequest struct {
	DateOfBirth           string `json:"dateOfBirth"`
	Gender                string `json:"gender"`
	BloodGroup            string `json:"bloodGroup"`
	Phone                 string `json:"phone"`
	InsuranceProvider     string `json:"insuranceProvider"`
	InsurancePolicyNumber string `json:"insurancePolicyNumber"`
}
