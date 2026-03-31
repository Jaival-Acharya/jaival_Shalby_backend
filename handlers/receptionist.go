package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"shalby_backend/models"
	"shalby_backend/services"

	"github.com/gin-gonic/gin"
)

// EmergencyContact holds emergency contact information
type EmergencyContact struct {
	Name     string `json:"name" binding:"required"`
	Relation string `json:"relation" binding:"required"`
	Phone    string `json:"phone" binding:"required"`
}

// RegisterPatientRequest is the request body for patient registration
type RegisterPatientRequest struct {
	FirstName         string             `json:"first_name" binding:"required"`
	LastName          string             `json:"last_name" binding:"required"`
	Email             string             `json:"email" binding:"required,email"`
	Phone             string             `json:"phone" binding:"required"`
	Password          string             `json:"password"`                         // Optional - auto-generated if not provided
	DateOfBirth       string             `json:"date_of_birth" binding:"required"` // YYYY-MM-DD
	Gender            string             `json:"gender" binding:"required"`        // M/F/O
	BloodGroup        string             `json:"blood_group"`                      // e.g., A+, O-
	CityID            int                `json:"city_id"`                          // Optional
	Address           string             `json:"address" binding:"required"`
	Allergies         []string           `json:"allergies"`
	Conditions        []string           `json:"conditions"`
	EmergencyContacts []EmergencyContact `json:"emergencyContacts"` // Emergency contact(s)
}

// BookAppointmentRequest is the request body for booking an appointment
type BookAppointmentRequest struct {
	PatientID       string `json:"patient_id" binding:"required"`
	DoctorID        string `json:"doctor_id" binding:"required"`
	AppointmentDate string `json:"appointment_date" binding:"required"` // YYYY-MM-DD
	AppointmentTime string `json:"appointment_time" binding:"required"` // HH:MM
	Complaint       string `json:"complaint" binding:"required"`
	AppointmentType string `json:"appointment_type" binding:"required"` // e.g. "First Visit" or "Follow-up"
}

// CheckInPatientRequest is the request body for patient check-in
type CheckInPatientRequest struct {
	AppointmentID string `json:"appointment_id" binding:"required"`
	BedID         string `json:"bed_id"` // Optional - bed can be assigned later
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
		receptionistID := c.GetString("user_id")
		if receptionistID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		service := services.NewReceptionistService(db)
		patientID, userID, err := service.RegisterPatient(services.RegisterPatientInput{
			FirstName:         req.FirstName,
			LastName:          req.LastName,
			Email:             req.Email,
			Phone:             req.Phone,
			Password:          req.Password,
			DateOfBirth:       req.DateOfBirth,
			Gender:            req.Gender,
			BloodGroup:        req.BloodGroup,
			CityID:            req.CityID,
			Address:           req.Address,
			Allergies:         req.Allergies,
			Conditions:        req.Conditions,
			EmergencyContacts: convertEmergencyContacts(req.EmergencyContacts),
		})

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Fetch the newly registered patient's full data
		query := `
			SELECT 
				u.name as patient_name,
				u.email,
				COALESCE(p.phone, '') as phone,
				p.date_of_birth,
				p.gender,
				p.blood_group,
				(SELECT COUNT(*) FROM appointments WHERE patient_id = p.id) as appointment_count
			FROM patients p
			JOIN users u ON p.user_id = u.id
			WHERE p.id = $1
		`

		var patientName, email, phone, gender, bloodGroup string
		var dateOfBirth time.Time
		var appointmentCount int

		err = db.QueryRow(query, patientID).Scan(
			&patientName, &email, &phone, &dateOfBirth, &gender, &bloodGroup, &appointmentCount,
		)

		if err != nil {
			// Even if fetch fails, return success with basic data
			c.JSON(http.StatusCreated, gin.H{
				"message": "Patient registered successfully",
				"patient": map[string]interface{}{
					"id":                patientID,
					"name":              req.FirstName + " " + req.LastName,
					"email":             req.Email,
					"phone":             req.Phone,
					"date_of_birth":     req.DateOfBirth,
					"gender":            req.Gender,
					"blood_group":       req.BloodGroup,
					"appointment_count": 0,
				},
				"patient_id": patientID,
				"user_id":    userID,
			})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "Patient registered successfully",
			"patient": map[string]interface{}{
				"id":                patientID,
				"name":              patientName,
				"email":             email,
				"phone":             phone,
				"date_of_birth":     dateOfBirth.Format("2006-01-02"),
				"gender":            gender,
				"blood_group":       bloodGroup,
				"appointment_count": appointmentCount,
			},
			"patient_id": patientID,
			"user_id":    userID,
		})
	}
}

// convertEmergencyContacts converts from handler request type to service type
func convertEmergencyContacts(contacts []EmergencyContact) []services.EmergencyContactInput {
	result := make([]services.EmergencyContactInput, len(contacts))
	for i, c := range contacts {
		result[i] = services.EmergencyContactInput{
			Name:     c.Name,
			Relation: c.Relation,
			Phone:    c.Phone,
		}
	}
	return result
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
			AppointmentType: req.AppointmentType,
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

// GetAppointmentsByFilter returns appointments filtered by date range and status
// GET /api/receptionist/appointments?filter=today|upcoming|past&limit=20&offset=0
func GetAppointmentsByFilter(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		receptionistID := c.GetString("user_id")
		if receptionistID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		filter := c.DefaultQuery("filter", "today") // today, upcoming, past
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
		appointments, err := service.GetAppointmentsByFilter(filter, limit, offset)
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

// GetReceptionistPatients returns all registered patients
// GET /api/receptionist/patients?search=&limit=20&offset=0
func GetReceptionistPatients(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		receptionistID := c.GetString("user_id")
		if receptionistID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		search := c.Query("search")
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
		patients, err := service.GetAllPatients(search, limit, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch patients"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"patients": patients,
			"count":    len(patients),
		})
	}
}

// GetReceptionistDoctors returns all active doctors with their schedules
// GET /api/receptionist/doctors
func GetReceptionistDoctors(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		receptionistID := c.GetString("user_id")
		if receptionistID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		service := services.NewReceptionistService(db)
		doctors, err := service.GetAllDoctorsWithSchedules()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch doctors"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"doctors": doctors,
		})
	}
}

// GetAllBeds returns all beds organized by room and department
// GET /api/receptionist/beds
func GetAllBeds(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		receptionistID := c.GetString("user_id")
		if receptionistID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		service := services.NewReceptionistService(db)
		beds, err := service.GetAllBeds()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch beds"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"beds": beds,
		})
	}
}

// ReassignBedRequest is the request body for bed reassignment
type ReassignBedRequest struct {
	AppointmentID string `json:"appointment_id" binding:"required"`
	NewBedID      string `json:"new_bed_id" binding:"required"`
}

// ReassignBed reassigns a patient to a different bed
// PUT /api/receptionist/beds/:bedId/reassign
func ReassignBed(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		receptionistID := c.GetString("user_id")
		if receptionistID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		currentBedID := c.Param("bedId")
		var req ReassignBedRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
			return
		}

		service := services.NewReceptionistService(db)
		err := service.ReassignBedToAppointment(req.AppointmentID, currentBedID, req.NewBedID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Bed reassignment successful",
		})
	}
}

// CreateNurseTaskRequest is the request body for creating a nurse task
type CreateNurseTaskRequest struct {
	AppointmentID string `json:"appointment_id"`                 // Optional
	PatientID     string `json:"patient_id"`                     // Optional
	NurseID       string `json:"nurse_id" binding:"required"`    // Required - assigned nurse
	TaskType      string `json:"task_type" binding:"required"`   // Required
	Description   string `json:"description" binding:"required"` // Required
	Priority      string `json:"priority"`                       // Optional
	DueBy         string `json:"due_by"`                         // Optional
}

// CreateNurseTask creates a new nurse task/request
// POST /api/receptionist/nurse-tasks
func CreateNurseTask(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		receptionistID := c.GetString("user_id")
		if receptionistID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		var req CreateNurseTaskRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
			return
		}

		service := services.NewReceptionistService(db)
		taskID, err := service.CreateNurseTask(req.AppointmentID, req.PatientID, req.TaskType, req.Description, req.NurseID, receptionistID, req.Priority, req.DueBy)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "Nurse task created successfully",
			"task_id": taskID,
		})
	}
}

// GetNurseTasks returns pending and active nurse tasks
// GET /api/receptionist/nurse-tasks
func GetNurseTasks(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		receptionistID := c.GetString("user_id")
		if receptionistID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		service := services.NewReceptionistService(db)
		tasks, err := service.GetNurseTasks()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch nurse tasks"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"tasks": tasks,
		})
	}
}

// UpdateNurseTaskRequest is the request body for updating nurse task details
type UpdateNurseTaskRequest struct {
	TaskType    string `json:"task_type"`
	PatientID   string `json:"patient_id"`
	Description string `json:"description"`
	Priority    string `json:"priority"`
	DueBy       string `json:"due_by"`
}

// UpdateNurseTask updates nurse task details
// PUT /api/receptionist/nurse-tasks/:taskId
func UpdateNurseTask(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		taskID := c.Param("taskId")
		if taskID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Task ID is required"})
			return
		}

		receptionistID := c.GetString("user_id")
		if receptionistID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		var req UpdateNurseTaskRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		service := services.NewReceptionistService(db)
		err := service.UpdateNurseTask(taskID, req.TaskType, req.PatientID, req.Description, req.Priority, req.DueBy)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Task updated successfully",
			"task_id": taskID,
		})
	}
}

// DeleteTask deletes a nurse task
// DELETE /api/receptionist/nurse-tasks/:taskId
func DeleteTask(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		taskID := c.Param("taskId")
		if taskID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Task ID is required"})
			return
		}

		receptionistID := c.GetString("user_id")
		if receptionistID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		service := services.NewReceptionistService(db)
		err := service.DeleteTask(taskID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Task deleted successfully",
			"task_id": taskID,
		})
	}
}

// UpdateTaskStatusRequest is the request body for updating task status
type UpdateTaskStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

// UpdateTaskStatus updates the status of a nurse task
// PATCH /api/receptionist/nurse-tasks/:taskId/status
func UpdateTaskStatus(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		taskID := c.Param("taskId")
		if taskID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Task ID is required"})
			return
		}

		receptionistID := c.GetString("user_id")
		if receptionistID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		var req UpdateTaskStatusRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		service := services.NewReceptionistService(db)
		err := service.UpdateTaskStatus(taskID, req.Status)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Task status updated successfully",
			"task_id": taskID,
			"status":  req.Status,
		})
	}
}

// DischargeBedRequest is the request body for discharging a patient from a bed
type DischargeBedRequest struct {
	Notes string `json:"notes"` // Optional discharge notes
}

// DischargeBedPatient discharges a patient from a bed
// POST /api/receptionist/beds/:bedId/discharge
func DischargeBedPatient(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		bedID := c.Param("bedId")
		if bedID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Bed ID is required"})
			return
		}

		receptionistID := c.GetString("user_id")
		if receptionistID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		var req DischargeBedRequest
		c.ShouldBindJSON(&req) // Optional body

		service := services.NewReceptionistService(db)
		err := service.DischargeBedPatient(bedID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Patient discharged successfully",
			"bed_id":  bedID,
		})
	}
}

// GetBedDischargeInfo returns information about a patient in a bed for discharge
// GET /api/receptionist/beds/:bedId/discharge-info
func GetBedDischargeInfo(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		bedID := c.Param("bedId")
		if bedID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Bed ID is required"})
			return
		}

		receptionistID := c.GetString("user_id")
		if receptionistID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		service := services.NewReceptionistService(db)
		info, err := service.GetBedDischargeInfo(bedID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": info,
		})
	}
}

// GetReceptionistProfile returns the receptionist user's profile information
func GetReceptionistProfile(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		if userID == "" {
			log.Println("Error: User not authenticated")
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: "User not authenticated",
			})
			return
		}

		service := services.NewReceptionistService(db)
		profile, err := service.GetProfile(userID)
		if err != nil {
			log.Printf("ERROR: Failed to get receptionist profile: %v\n", err)
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: "Failed to get profile",
			})
			return
		}

		c.JSON(http.StatusOK, models.SuccessResponse{
			Message: "Receptionist profile fetched successfully",
			Data:    profile,
		})
	}
}

// UpdateReceptionistProfile updates the receptionist user's profile information
func UpdateReceptionistProfile(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		if userID == "" {
			log.Println("Error: User not authenticated")
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: "User not authenticated",
			})
			return
		}

		var updateData map[string]interface{}
		if err := c.BindJSON(&updateData); err != nil {
			log.Println("Error parsing profile update request:", err)
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error: "Invalid request data",
			})
			return
		}

		log.Printf("DEBUG: Updating receptionist profile for user %s with data: %v\n", userID, updateData)
		service := services.NewReceptionistService(db)

		// Update the profile
		profile, err := service.UpdateProfile(userID, updateData)
		if err != nil {
			log.Printf("ERROR: Failed to update receptionist profile: %v\n", err)
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: fmt.Sprintf("Failed to update profile: %v", err),
			})
			return
		}

		c.JSON(http.StatusOK, models.SuccessResponse{
			Message: "Receptionist profile updated successfully",
			Data:    profile,
		})
	}
}
