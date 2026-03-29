package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"shalby_backend/services"

	"github.com/gin-gonic/gin"
)

// RegisterPatientRequest is the request body for patient registration
type RegisterPatientRequest struct {
	FirstName   string   `json:"first_name" binding:"required"`
	LastName    string   `json:"last_name" binding:"required"`
	Email       string   `json:"email" binding:"required,email"`
	Phone       string   `json:"phone" binding:"required"`
	DateOfBirth string   `json:"date_of_birth" binding:"required"` // YYYY-MM-DD
	Gender      string   `json:"gender" binding:"required"`        // M/F/O
	CityID      int      `json:"city_id" binding:"required"`
	Address     string   `json:"address" binding:"required"`
	Allergies   []string `json:"allergies"`
	Conditions  []string `json:"conditions"`
}

// BookAppointmentRequest is the request body for booking an appointment
type BookAppointmentRequest struct {
	PatientID       string `json:"patient_id" binding:"required"`
	DoctorID        string `json:"doctor_id" binding:"required"`
	AppointmentDate string `json:"appointment_date" binding:"required"` // YYYY-MM-DD
	AppointmentTime string `json:"appointment_time" binding:"required"` // HH:MM
	Complaint       string `json:"complaint" binding:"required"`
}

// CheckInPatientRequest is the request body for patient check-in
type CheckInPatientRequest struct {
	AppointmentID string `json:"appointment_id" binding:"required"`
	BedID         string `json:"bed_id" binding:"required"`
}

// RegisterPatient handles patient registration
// POST /api/receptionist/register-patient
func RegisterPatient(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req RegisterPatientRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
			return
		}

		// Get user ID from context (set by auth middleware)
		// For now, we'll assume it's passed in headers or context
		receptionistID := c.GetString("user_id") // Set by auth middleware
		if receptionistID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		service := services.NewReceptionistService(db)
		patientID, userID, err := service.RegisterPatient(services.RegisterPatientInput{
			FirstName:   req.FirstName,
			LastName:    req.LastName,
			Email:       req.Email,
			Phone:       req.Phone,
			DateOfBirth: req.DateOfBirth,
			Gender:      req.Gender,
			CityID:      req.CityID,
			Address:     req.Address,
			Allergies:   req.Allergies,
			Conditions:  req.Conditions,
		})

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message":    "Patient registered successfully",
			"patient_id": patientID,
			"user_id":    userID,
		})
	}
}

// BookAppointment handles appointment booking
// POST /api/receptionist/book-appointment
func BookAppointment(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req BookAppointmentRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
			return
		}

		receptionistID := c.GetString("user_id") // Set by auth middleware
		if receptionistID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		service := services.NewReceptionistService(db)
		appointmentID, err := service.BookAppointment(services.BookAppointmentInput{
			PatientID:       req.PatientID,
			DoctorID:        req.DoctorID,
			AppointmentDate: req.AppointmentDate,
			AppointmentTime: req.AppointmentTime,
			Complaint:       req.Complaint,
		})

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message":        "Appointment booked successfully",
			"appointment_id": appointmentID,
		})
	}
}

// CheckInPatient handles patient check-in
// POST /api/receptionist/check-in
func CheckInPatient(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CheckInPatientRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
			return
		}

		receptionistID := c.GetString("user_id") // Set by auth middleware
		if receptionistID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		service := services.NewReceptionistService(db)
		err := service.CheckInPatient(services.CheckInPatientInput{
			AppointmentID:  req.AppointmentID,
			BedID:          req.BedID,
			ReceptionistID: receptionistID,
		})

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Patient checked in successfully",
		})
	}
}

// GetPendingAppointments returns upcoming appointments for receptionist queue
// GET /api/receptionist/pending-appointments?limit=20&offset=0
func GetPendingAppointments(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		receptionistID := c.GetString("user_id") // Set by auth middleware
		if receptionistID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		// Parse pagination parameters
		limit := 20
		offset := 0

		if limitStr := c.Query("limit"); limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
				limit = l
			}
		}

		if offsetStr := c.Query("offset"); offsetStr != "" {
			if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
				offset = o
			}
		}

		service := services.NewReceptionistService(db)
		appointments, err := service.GetPendingAppointments(limit, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch appointments"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"appointments": appointments,
			"count":        len(appointments),
		})
	}
}
