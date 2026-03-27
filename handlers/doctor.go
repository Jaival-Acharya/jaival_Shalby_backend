package handlers

import (
	"database/sql"
	"log"
	"net/http"

	"shalby_backend/models"
	"shalby_backend/services"

	"github.com/gin-gonic/gin"
)

// GetAllDoctors returns list of all doctors
func GetAllDoctors(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		doctors, err := services.GetAllDoctors(db)
		if err != nil {
			log.Println("Error fetching doctors:", err)
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: "Failed to fetch doctors",
			})
			return
		}

		c.JSON(http.StatusOK, models.SuccessResponse{
			Message: "Doctors fetched successfully",
			Data:    doctors,
		})
	}
}

// GetDoctorByID returns a single doctor with full details
func GetDoctorByID(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		doctorID := c.Param("id")

		doctor, err := services.GetDoctorByID(db, doctorID)
		if err != nil {
			log.Println("Error fetching doctor:", err)
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: "Doctor not found",
			})
			return
		}

		c.JSON(http.StatusOK, models.SuccessResponse{
			Message: "Doctor fetched successfully",
			Data:    doctor,
		})
	}
}

// CreateDoctor creates a new doctor
func CreateDoctor(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.DoctorRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error: "Invalid request format: " + err.Error(),
			})
			return
		}

		doctor, err := services.CreateDoctor(db, req)
		if err != nil {
			log.Println("Error creating doctor:", err)
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: err.Error(),
			})
			return
		}

		c.JSON(http.StatusCreated, models.SuccessResponse{
			Message: "Doctor created successfully",
			Data:    doctor,
		})
	}
}

// UpdateDoctor updates a doctor
func UpdateDoctor(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		doctorID := c.Param("id")
		var req models.DoctorUpdateRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error: "Invalid request format: " + err.Error(),
			})
			return
		}

		doctor, err := services.UpdateDoctor(db, doctorID, req)
		if err != nil {
			log.Println("Error updating doctor:", err)
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, models.SuccessResponse{
			Message: "Doctor updated successfully",
			Data:    doctor,
		})
	}
}

// DeleteDoctor deletes a doctor
func DeleteDoctor(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		doctorID := c.Param("id")

		err := services.DeleteDoctor(db, doctorID)
		if err != nil {
			log.Println("Error deleting doctor:", err)
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, models.SuccessResponse{
			Message: "Doctor deleted successfully",
		})
	}
}
