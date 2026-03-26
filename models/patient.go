package models

import "time"

type Patient struct {
	ID          string    `json:"id"`
	UserID      string    `json:"userId"`
	DateOfBirth string    `json:"dateOfBirth"`
	Gender      string    `json:"gender"`
	BloodGroup  string    `json:"bloodGroup"`
	Phone       string    `json:"phone"`
	Allergies   []string  `json:"allergies"`
	Conditions  []string  `json:"conditions"`
	CreatedAt   time.Time `json:"createdAt"`
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

type SignupRequest struct {
	FullName                 string   `json:"fullName" binding:"required"`
	DateOfBirth              string   `json:"dateOfBirth" binding:"required"`
	Gender                   string   `json:"gender" binding:"required"`
	BloodGroup               string   `json:"bloodGroup"`
	Phone                    string   `json:"phone"`
	Email                    string   `json:"email" binding:"required,email"`
	Password                 string   `json:"password" binding:"required,min=6"`
	Allergies                []string `json:"allergies"`
	Conditions               []string `json:"conditions"`
	EmergencyContactName     string   `json:"emergencyContactName"`
	EmergencyContactPhone    string   `json:"emergencyContactPhone"`
	EmergencyContactRelation string   `json:"emergencyContactRelation"`
}

type SignupResponse struct {
	Message string `json:"message"`
	UserId  string `json:"userId"`
}
