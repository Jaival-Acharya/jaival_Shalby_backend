package models

import "time"

type User struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	IsActive     bool      `json:"isActive"`
	CreatedAt    time.Time `json:"createdAt"`
	Roles        []string  `json:"roles"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// SignupRequest for patient registration
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

// SignupResponse for successful registration
type SignupResponse struct {
	Message string `json:"message"`
	UserID  string `json:"userId"`
}
