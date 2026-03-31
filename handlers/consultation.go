package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"

	"shalby_backend/models"
	"shalby_backend/services"

	"github.com/gin-gonic/gin"
)

// CreateConsultation handles POST /api/doctor/consultations/:appointmentId
func CreateConsultation(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		appointmentID := c.Param("appointmentId")
		doctorID := c.GetString("doctor_id") // From auth context
		patientID := c.Query("patientId")

		var req services.CreateConsultationRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error: "Invalid request format: " + err.Error(),
			})
			return
		}

		consultation, err := services.CreateConsultation(db, appointmentID, doctorID, patientID, req)
		if err != nil {
			log.Println("Error creating consultation:", err)
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: err.Error(),
			})
			return
		}

		c.JSON(http.StatusCreated, models.SuccessResponse{
			Message: "Consultation created successfully",
			Data:    consultation,
		})
	}
}

// GetConsultationByAppointmentID handles GET /api/doctor/appointments/:appointmentId/consultation
func GetConsultationByAppointmentID(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		appointmentID := c.Param("appointmentId")

		consultation, err := services.GetConsultationByAppointmentID(db, appointmentID)
		if err != nil {
			log.Println("Error fetching consultation:", err)
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: "Failed to fetch consultation",
			})
			return
		}

		if consultation == nil {
			c.JSON(http.StatusOK, nil)
			return
		}

		c.JSON(http.StatusOK, models.SuccessResponse{
			Message: "Consultation fetched successfully",
			Data:    consultation,
		})
	}
}

// GetConsultationByID handles GET /api/doctor/consultations/:id
func GetConsultationByID(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		consultationID := c.Param("id")

		consultation, err := services.GetConsultationByID(db, consultationID)
		if err != nil {
			log.Println("Error fetching consultation:", err)
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: "Consultation not found",
			})
			return
		}

		c.JSON(http.StatusOK, models.SuccessResponse{
			Message: "Consultation fetched successfully",
			Data:    consultation,
		})
	}
}

// GetConsultationHistory handles GET /api/doctor/patients/:patientId/consultations
func GetConsultationHistory(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		patientID := c.Param("patientId")

		limit := 20
		offset := 0

		if l := c.Query("limit"); l != "" {
			if parsed, err := strconv.Atoi(l); err == nil {
				limit = parsed
			}
		}

		if o := c.Query("offset"); o != "" {
			if parsed, err := strconv.Atoi(o); err == nil {
				offset = parsed
			}
		}

		consultations, count, err := services.GetConsultationHistory(db, patientID, limit, offset)
		if err != nil {
			log.Println("Error fetching consultations:", err)
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: "Failed to fetch consultations",
			})
			return
		}

		c.JSON(http.StatusOK, models.SuccessResponse{
			Message: "Consultations fetched successfully",
			Data: gin.H{
				"consultations": consultations,
				"count":         count,
			},
		})
	}
}
