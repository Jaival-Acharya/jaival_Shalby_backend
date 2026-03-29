package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

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
