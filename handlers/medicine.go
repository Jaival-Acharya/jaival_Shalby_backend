package handlers

import (
	"database/sql"
	"log"
	"net/http"

	"shalby_backend/models"
	"shalby_backend/services"

	"github.com/gin-gonic/gin"
)

// GetAllMedicines returns list of all medicines
func GetAllMedicines(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		medicines, err := services.GetAllMedicines(db)
		if err != nil {
			log.Println("Error fetching medicines:", err)
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: "Failed to fetch medicines",
			})
			return
		}

		c.JSON(http.StatusOK, models.SuccessResponse{
			Message: "Medicines fetched successfully",
			Data:    medicines,
		})
	}
}

// GetMedicineStats returns medicine statistics
func GetMedicineStats(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		stats, err := services.GetMedicineStats(db)
		if err != nil {
			log.Println("Error fetching medicine stats:", err)
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: "Failed to fetch medicine statistics",
			})
			return
		}

		c.JSON(http.StatusOK, models.SuccessResponse{
			Message: "Medicine stats fetched successfully",
			Data:    stats,
		})
	}
}

// CreateMedicine creates a new medicine
func CreateMedicine(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.CreateMedicineRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error: "Invalid request format: " + err.Error(),
			})
			return
		}

		// Convert CreateMedicineRequest to separate requests
		medReq := models.MedicineRequest{
			Name:         req.Name,
			GenericName:  req.GenericName,
			Category:     req.Category,
			DosageForm:   req.DosageForm,
			Manufacturer: req.Manufacturer,
			Price:        req.Price,
			IsActive:     req.IsActive,
		}

		invReq := models.MedicineInventoryRequest{
			StockQuantity: req.StockQuantity,
			Unit:          req.Unit,
			ReorderLevel:  req.ReorderLevel,
			ExpiryDate:    req.ExpiryDate,
		}

		medicine, err := services.CreateMedicine(db, medReq, invReq)
		if err != nil {
			log.Println("Error creating medicine:", err)
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: err.Error(),
			})
			return
		}

		c.JSON(http.StatusCreated, models.SuccessResponse{
			Message: "Medicine created successfully",
			Data:    medicine,
		})
	}
}

// UpdateMedicine updates a medicine
func UpdateMedicine(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		medicineID := c.Param("id")
		var req models.MedicineRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error: "Invalid request format: " + err.Error(),
			})
			return
		}

		medicine, err := services.UpdateMedicine(db, medicineID, req)
		if err != nil {
			log.Println("Error updating medicine:", err)
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, models.SuccessResponse{
			Message: "Medicine updated successfully",
			Data:    medicine,
		})
	}
}

// DeleteMedicine deletes a medicine
func DeleteMedicine(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		medicineID := c.Param("id")

		err := services.DeleteMedicine(db, medicineID)
		if err != nil {
			log.Println("Error deleting medicine:", err)
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, models.SuccessResponse{
			Message: "Medicine deleted successfully",
		})
	}
}
