package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"shalby_backend/services"

	"github.com/gin-gonic/gin"
)

// RecordVitalsRequest is the request body for recording patient vitals
type RecordVitalsRequest struct {
	AppointmentID             string  `json:"appointment_id" binding:"required"`
	PatientID                 string  `json:"patient_id" binding:"required"`
	TemperatureCelsius        float64 `json:"temperature_celsius" binding:"required"`
	BloodPressureSystolic     int     `json:"blood_pressure_systolic" binding:"required"`
	BloodPressureDiastolic    int     `json:"blood_pressure_diastolic" binding:"required"`
	PulseHeartRateBpm         int     `json:"pulse_heart_rate_bpm" binding:"required"`
	RespiratoryRateBreathsMin int     `json:"respiratory_rate_breaths_min"`
	OxygenSaturationPercent   float64 `json:"oxygen_saturation_percent"`
	HeightCm                  float64 `json:"height_cm"`
	WeightKg                  float64 `json:"weight_kg"`
	Notes                     string  `json:"notes"`
}

// RecordVitals handles vital signs recording for a patient
// POST /api/nurse/record-vitals
func RecordVitals(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req RecordVitalsRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
			return
		}

		// Get nurse ID from context (set by auth middleware)
		nurseID := c.GetString("user_id")
		if nurseID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		service := services.NewNurseService(db)
		vitalID, err := service.RecordVitals(services.RecordVitalsInput{
			AppointmentID:             req.AppointmentID,
			PatientID:                 req.PatientID,
			TemperatureCelsius:        req.TemperatureCelsius,
			BloodPressureSystolic:     req.BloodPressureSystolic,
			BloodPressureDiastolic:    req.BloodPressureDiastolic,
			PulseHeartRateBpm:         req.PulseHeartRateBpm,
			RespiratoryRateBreathsMin: req.RespiratoryRateBreathsMin,
			OxygenSaturationPercent:   req.OxygenSaturationPercent,
			HeightCm:                  req.HeightCm,
			WeightKg:                  req.WeightKg,
			Notes:                     req.Notes,
		}, nurseID)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message":  "Vitals recorded successfully",
			"vital_id": vitalID,
			"status":   "Appointment status updated to 'Ready for Doctor'",
		})
	}
}

// GetCheckedInPatients returns patients currently checked in for vitals recording
// GET /api/nurse/checked-in-patients?limit=20&offset=0
func GetCheckedInPatients(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		nurseID := c.GetString("user_id")
		if nurseID == "" {
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

		service := services.NewNurseService(db)
		patients, err := service.GetCheckedInPatients(limit, offset)
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

// GetPatientVitals returns vital history for a patient
// GET /api/nurse/patient-vitals/:patient_id?limit=10
func GetPatientVitals(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		nurseID := c.GetString("user_id")
		if nurseID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		patientID := c.Param("patient_id")
		if patientID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Patient ID is required"})
			return
		}

		limit := 10
		if limitStr := c.Query("limit"); limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
				limit = l
			}
		}

		service := services.NewNurseService(db)
		vitals, err := service.GetPatientVitals(patientID, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch vitals"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"patient_id": patientID,
			"vitals":     vitals,
			"count":      len(vitals),
		})
	}
}

// GetRecentVitals returns the most recent vitals recorded across all patients
// GET /api/nurse/recent-vitals?limit=10
func GetRecentVitals(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		nurseID := c.GetString("user_id")
		if nurseID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		limit := 10
		if limitStr := c.Query("limit"); limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
				limit = l
			}
		}

		service := services.NewNurseService(db)
		vitals, err := service.GetRecentVitals(limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch recent vitals"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"vitals": vitals,
			"count":  len(vitals),
		})
	}
}

// GetNurseQueue returns today's appointment queue with vitals status
// GET /api/nurse/queue
func GetNurseQueue(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		nurseID := c.GetString("user_id")
		if nurseID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		today := time.Now().Format("2006-01-02")

		service := services.NewNurseService(db)
		queue, err := service.GetQueue(nurseID, today)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"queue": queue,
			"count": len(queue),
			"date":  today,
		})
	}
}

// GetMyTasks returns tasks assigned to the authenticated nurse
// GET /api/nurse/tasks
func GetMyTasks(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		nurseID := c.GetString("user_id")
		if nurseID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		service := services.NewNurseService(db)
		tasks, err := service.GetTasksForNurse(nurseID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"tasks": tasks,
			"count": len(tasks),
		})
	}
}

// UpdateMyTask updates a task's status and notes
// PUT /api/nurse/tasks/:id
func UpdateMyTask(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		nurseID := c.GetString("user_id")
		if nurseID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		taskID := c.Param("id")
		if taskID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Task ID is required"})
			return
		}

		var req struct {
			Status string `json:"status"`
			Note   string `json:"note,omitempty"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
			return
		}

		if req.Status == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Status is required"})
			return
		}

		service := services.NewNurseService(db)
		err := service.UpdateTaskStatus(taskID, nurseID, req.Status, req.Note)
		if err != nil {
			if err.Error() == "task not found" {
				c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
				return
			}
			if err.Error() == "you are not assigned to this task" {
				c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Task updated successfully",
			"taskId":  taskID,
		})
	}
}

// RecordAdmittedVitalsRequest is the request body for recording vitals for admitted patients
type RecordAdmittedVitalsRequest struct {
	PatientID   string  `json:"patientId" binding:"required"`
	HeightCm    float64 `json:"heightCm"`
	WeightKg    float64 `json:"weightKg"`
	BpSystolic  int     `json:"bpSystolic" binding:"required"`
	BpDiastolic int     `json:"bpDiastolic" binding:"required"`
	BloodSugar  float64 `json:"bloodSugar"`
	Temperature float64 `json:"temperature" binding:"required"`
	Pulse       int     `json:"pulse" binding:"required"`
	Notes       string  `json:"notes"`
}

// RecordAdmittedPatientVitals records vitals for an admitted patient
// POST /api/nurse/vitals/admitted/:patientId
func RecordAdmittedPatientVitals(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		patientID := c.Param("patientId")
		nurseID := c.GetString("user_id")
		if nurseID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		var req RecordAdmittedVitalsRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
			return
		}

		// Insert vitals record
		query := `
		INSERT INTO patient_vitals (
			patient_id,
			recorded_by,
			height_cm,
			weight_kg,
			blood_pressure_systolic,
			blood_pressure_diastolic,
			blood_sugar_mg_dl,
			temperature_celsius,
			pulse_bpm,
			recorded_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
		RETURNING id
		`

		var vitalID string
		err := db.QueryRow(
			query,
			patientID,
			nurseID,
			req.HeightCm,
			req.WeightKg,
			req.BpSystolic,
			req.BpDiastolic,
			req.BloodSugar,
			req.Temperature,
			req.Pulse,
		).Scan(&vitalID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record vitals: " + err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message":  "Vitals recorded successfully",
			"vital_id": vitalID,
		})
	}
}

// GetAdmittedPatients returns all currently admitted patients in beds
// GET /api/nurse/patients/admitted
func GetAdmittedPatients(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		nurseID := c.GetString("user_id")
		if nurseID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		service := services.NewNurseService(db)
		patients, err := service.GetAdmittedPatients()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"patients": patients,
			"count":    len(patients),
		})
	}
}

// GetNurseStats returns dashboard statistics for the nurse
// GET /api/nurse/stats
func GetNurseStats(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		nurseID := c.GetString("user_id")
		if nurseID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		today := time.Now().Format("2006-01-02")

		service := services.NewNurseService(db)
		stats, err := service.GetNurseStats(nurseID, today)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, stats)
	}
}

// GetNurseProfile returns the authenticated nurse's profile information
// GET /api/nurse/profile
func GetNurseProfile(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		nurseID := c.GetString("user_id")
		if nurseID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		var nurse struct {
			ID    string `db:"id"`
			Name  string `db:"name"`
			Email string `db:"email"`
			Phone string `db:"phone"`
			Role  string `db:"role"`
		}

		query := `SELECT id, name, email, phone, role FROM users WHERE id = $1 AND role = 'Nurse'`
		err := db.QueryRow(query, nurseID).Scan(
			&nurse.ID,
			&nurse.Name,
			&nurse.Email,
			&nurse.Phone,
			&nurse.Role,
		)

		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Nurse profile not found"})
			return
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch profile"})
			return
		}

		c.JSON(http.StatusOK, nurse)
	}
}

// UpdateNurseProfileRequest is the request body for updating nurse profile
type UpdateNurseProfileRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone"`
}

// UpdateNurseProfile updates the nurse's profile information
// PATCH /api/nurse/profile
func UpdateNurseProfile(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		nurseID := c.GetString("user_id")
		if nurseID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		var req UpdateNurseProfileRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		query := `UPDATE users SET name = $1, email = $2, phone = $3 WHERE id = $4 AND role = 'Nurse'`
		_, err := db.Exec(query, req.Name, req.Email, req.Phone, nurseID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Profile updated successfully",
			"id":      nurseID,
		})
	}
}
