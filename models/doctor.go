package models

import (
	"database/sql"
	"time"
)

// Doctor represents a doctor in the system - links to users table
type Doctor struct {
	ID              string         `json:"id"`
	UserID          string         `json:"userId"`
	Name            string         `json:"name"`
	Email           string         `json:"email"`
	AvatarURL       string         `json:"avatarUrl"`
	Specialization  string         `json:"specialization"`
	Department      string         `json:"department"`
	LicenseNumber   string         `json:"licenseNumber"`
	ConsultationFee float64        `json:"consultationFee"`
	JoiningDate     sql.NullString `json:"joiningDate"` // nullable DATE
	IsActive        bool           `json:"isActive"`
	CreatedAt       time.Time      `json:"createdAt"`
}

type DoctorRequest struct {
	Name            string  `json:"name" binding:"required"`
	Email           string  `json:"email" binding:"required,email"`
	Password        string  `json:"password"` // Optional - auto-generated if not provided
	Phone           string  `json:"phone"`
	Address         string  `json:"address"`
	Specialization  string  `json:"specialization" binding:"required"`
	Department      string  `json:"department" binding:"required"`
	LicenseNumber   string  `json:"licenseNumber" binding:"required"`
	ConsultationFee float64 `json:"consultationFee"`
	JoiningDate     string  `json:"joiningDate"`
	AvatarURL       string  `json:"avatarUrl"`
}

type DoctorUpdateRequest struct {
	Name            string  `json:"name" binding:"required"`
	Email           string  `json:"email" binding:"required,email"`
	Phone           string  `json:"phone"`
	Address         string  `json:"address"`
	Specialization  string  `json:"specialization" binding:"required"`
	Department      string  `json:"department" binding:"required"`
	LicenseNumber   string  `json:"licenseNumber" binding:"required"`
	ConsultationFee float64 `json:"consultationFee"`
	JoiningDate     string  `json:"joiningDate"`
}

type DoctorResponse struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Email           string    `json:"email"`
	Specialization  string    `json:"specialization"`
	Department      string    `json:"department"`
	LicenseNumber   string    `json:"licenseNumber"`
	ConsultationFee float64   `json:"consultationFee"`
	JoiningDate     string    `json:"joiningDate"`
	IsActive        bool      `json:"isActive"`
	AvatarURL       string    `json:"avatarUrl"`
	CreatedAt       time.Time `json:"createdAt"`
}

type DoctorListResponse struct {
	Doctors []DoctorResponse `json:"doctors"`
	Total   int              `json:"total"`
}

type DoctorDetailResponse struct {
	ID              string                   `json:"id"`
	Name            string                   `json:"name"`
	Email           string                   `json:"email"`
	Phone           string                   `json:"phone"`
	Address         string                   `json:"address"`
	Specialization  string                   `json:"specialization"`
	Department      string                   `json:"department"`
	LicenseNumber   string                   `json:"licenseNumber"`
	ConsultationFee float64                  `json:"consultationFee"`
	JoiningDate     string                   `json:"joiningDate"`
	IsActive        bool                     `json:"isActive"`
	CreatedAt       time.Time                `json:"createdAt"`
	Education       []map[string]interface{} `json:"education"`
	WorkExperience  []map[string]interface{} `json:"workExperience"`
}

// DoctorSchedule represents working hours for a doctor
type DoctorSchedule struct {
	ID                  string `json:"id"`
	DoctorID            string `json:"doctorId"`
	DayOfWeek           int    `json:"dayOfWeek"` // 0=Sunday, 6=Saturday
	StartTime           string `json:"startTime"`
	EndTime             string `json:"endTime"`
	SlotDurationMinutes int    `json:"slotDurationMinutes"`
	MaxPatientsPerSlot  int    `json:"maxPatientsPerSlot"`
	IsActive            bool   `json:"isActive"`
}

type DoctorScheduleRequest struct {
	DayOfWeek           int    `json:"dayOfWeek" binding:"required,min=0,max=6"`
	StartTime           string `json:"startTime" binding:"required"`
	EndTime             string `json:"endTime" binding:"required"`
	SlotDurationMinutes int    `json:"slotDurationMinutes"`
	MaxPatientsPerSlot  int    `json:"maxPatientsPerSlot"`
	IsActive            bool   `json:"isActive"`
}

// DoctorLeave represents leave period for a doctor
type DoctorLeave struct {
	ID        string    `json:"id"`
	DoctorID  string    `json:"doctorId"`
	StartDate string    `json:"startDate"`
	EndDate   string    `json:"endDate"`
	Reason    string    `json:"reason"`
	CreatedAt time.Time `json:"createdAt"`
}

type DoctorLeaveRequest struct {
	StartDate string `json:"startDate" binding:"required"`
	EndDate   string `json:"endDate" binding:"required"`
	Reason    string `json:"reason"`
}
