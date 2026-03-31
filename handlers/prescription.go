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

// CreatePrescription handles POST /api/doctor/prescriptions
func CreatePrescription(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req services.CreatePrescriptionRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error: "Invalid request format: " + err.Error(),
			})
			return
		}

		prescription, err := services.CreatePrescription(db, req)
		if err != nil {
			log.Println("Error creating prescription:", err)
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: err.Error(),
			})
			return
		}

		c.JSON(http.StatusCreated, models.SuccessResponse{
			Message: "Prescription created successfully",
			Data:    prescription,
		})
	}
}

// GetPrescriptionByID handles GET /api/doctor/prescriptions/:id
func GetPrescriptionByID(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		prescriptionID := c.Param("id")

		prescription, err := services.GetPrescriptionByID(db, prescriptionID)
		if err != nil {
			log.Println("Error fetching prescription:", err)
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: "Prescription not found",
			})
			return
		}

		c.JSON(http.StatusOK, models.SuccessResponse{
			Message: "Prescription fetched successfully",
			Data:    prescription,
		})
	}
}

// GetPrescriptionsByDoctor handles GET /api/doctor/prescriptions
func GetPrescriptionsByDoctor(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		doctorID := c.GetString("doctor_id") // From auth context
		status := c.Query("status")
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

		prescriptions, count, err := services.GetPrescriptionsByDoctor(db, doctorID, status, limit, offset)
		if err != nil {
			log.Println("Error fetching prescriptions:", err)
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: "Failed to fetch prescriptions",
			})
			return
		}

		c.JSON(http.StatusOK, models.SuccessResponse{
			Message: "Prescriptions fetched successfully",
			Data: gin.H{
				"prescriptions": prescriptions,
				"count":         count,
			},
		})
	}
}

// GetPendingPrescriptions handles GET /api/pharmacist/pending-prescriptions
func GetPendingPrescriptions(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
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

		prescriptions, count, err := services.GetPendingPrescriptions(db, limit, offset)
		if err != nil {
			log.Println("Error fetching pending prescriptions:", err)
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: "Failed to fetch pending prescriptions",
			})
			return
		}

		c.JSON(http.StatusOK, models.SuccessResponse{
			Message: "Pending prescriptions fetched successfully",
			Data: gin.H{
				"prescriptions": prescriptions,
				"count":         count,
			},
		})
	}
}

// DispensePrescription handles POST /api/pharmacist/dispense/:id
func DispensePrescription(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		prescriptionID := c.Param("id")
		pharmacistID := c.GetString("user_id") // From auth context

		var req struct {
			Items []struct {
				MedicineID        string `json:"medicineId"`
				QuantityDispensed int    `json:"quantityDispensed"`
			} `json:"items"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error: "Invalid request format: " + err.Error(),
			})
			return
		}

		// Convert request items to service struct format
		var serviceItems []struct {
			MedicineID        string
			QuantityDispensed int
		}
		for _, item := range req.Items {
			serviceItems = append(serviceItems, struct {
				MedicineID        string
				QuantityDispensed int
			}{
				MedicineID:        item.MedicineID,
				QuantityDispensed: item.QuantityDispensed,
			})
		}

		prescription, err := services.DispensePrescription(db, prescriptionID, pharmacistID, serviceItems)
		if err != nil {
			log.Println("Error dispensing prescription:", err)
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error: err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, models.SuccessResponse{
			Message: "Prescription dispensed successfully",
			Data:    prescription,
		})
	}
}
