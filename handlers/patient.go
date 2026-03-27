package handlers

import (
	"database/sql"
	"log"
	"net/http"

	"shalby_backend/models"
	"shalby_backend/services"

	"github.com/gin-gonic/gin"
)

// GetAllPatients returns list of all patients
func GetAllPatients(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		patients, err := services.GetAllPatients(db)
		if err != nil {
			log.Println("Error fetching patients:", err)
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: "Failed to fetch patients",
			})
			return
		}

		c.JSON(http.StatusOK, models.SuccessResponse{
			Message: "Patients fetched successfully",
			Data:    patients,
		})
	}
}

// GetPatientDetails returns detailed patient information
func GetPatientDetails(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		patientID := c.Param("id")

		patient, err := services.GetPatientDetails(db, patientID)
		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, models.ErrorResponse{
					Error: "Patient not found",
				})
				return
			}
			log.Println("Error fetching patient:", err)
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: "Failed to fetch patient",
			})
			return
		}

		c.JSON(http.StatusOK, models.SuccessResponse{
			Message: "Patient details fetched successfully",
			Data:    patient,
		})
	}
}

// CreatePatient creates a new patient
func CreatePatient(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.PatientCreateRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error: "Invalid request format: " + err.Error(),
			})
			return
		}

		patient, err := services.CreatePatient(db, req)
		if err != nil {
			log.Println("Error creating patient:", err)
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: err.Error(),
			})
			return
		}

		c.JSON(http.StatusCreated, models.SuccessResponse{
			Message: "Patient created successfully",
			Data:    patient,
		})
	}
}

// UpdatePatient updates patient information
func UpdatePatient(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		patientID := c.Param("id")
		var req map[string]interface{}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error: "Invalid request format: " + err.Error(),
			})
			return
		}

		patient, err := services.UpdatePatient(db, patientID, req)
		if err != nil {
			log.Println("Error updating patient:", err)
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, models.SuccessResponse{
			Message: "Patient updated successfully",
			Data:    patient,
		})
	}
}

// DeletePatient deletes a patient
func DeletePatient(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		patientID := c.Param("id")

		err := services.DeletePatient(db, patientID)
		if err != nil {
			log.Println("Error deleting patient:", err)
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, models.SuccessResponse{
			Message: "Patient deleted successfully",
		})
	}
}
