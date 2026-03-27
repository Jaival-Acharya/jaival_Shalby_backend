package handlers

import (
	"database/sql"
	"log"
	"net/http"

	"shalby_backend/models"
	"shalby_backend/services"

	"github.com/gin-gonic/gin"
)

// GetAdminDashboard returns admin dashboard statistics
func GetAdminDashboard(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		stats, err := services.GetDashboardStats(db)
		if err != nil {
			log.Println("Error fetching dashboard stats:", err)
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: "Failed to fetch dashboard statistics",
			})
			return
		}

		c.JSON(http.StatusOK, models.SuccessResponse{
			Message: "Dashboard statistics fetched successfully",
			Data:    stats,
		})
	}
}

// GetSystemSettings returns current system settings
func GetSystemSettings(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		settings, err := services.GetSystemSettings(db)
		if err != nil {
			log.Println("Error fetching system settings:", err)
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: "Failed to fetch system settings",
			})
			return
		}

		c.JSON(http.StatusOK, models.SuccessResponse{
			Message: "System settings fetched successfully",
			Data:    settings,
		})
	}
}

// UpdateSystemSettings updates system settings from the admin form
func UpdateSystemSettings(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.SystemSettingsRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error: "Invalid request format: " + err.Error(),
			})
			return
		}

		// Convert the structured request to individual settings updates
		settingsMap := map[string]string{
			"hospitalName":        req.HospitalName,
			"email":               req.Email,
			"phone":               req.Phone,
			"address":             req.Address,
			"website":             req.Website,
			"currency":            req.Currency,
			"timezone":            req.Timezone,
			"language":            req.Language,
			"notifications_email": "false",
			"notifications_sms":   "false",
			"notifications_push":  "false",
			"security_twoFactor":  "false",
			"security_autoLogout": req.Security.AutoLogout,
			"mobile_offlineMode":  "false",
		}

		// Convert booleans to strings
		if req.Notifications.Email {
			settingsMap["notifications_email"] = "true"
		}
		if req.Notifications.SMS {
			settingsMap["notifications_sms"] = "true"
		}
		if req.Notifications.Push {
			settingsMap["notifications_push"] = "true"
		}
		if req.Security.TwoFactor {
			settingsMap["security_twoFactor"] = "true"
		}
		if req.Mobile.OfflineMode {
			settingsMap["mobile_offlineMode"] = "true"
		}

		// Update all settings
		for key, value := range settingsMap {
			err := services.UpdateSystemSetting(db, key, value)
			if err != nil {
				log.Println("Error updating setting "+key+":", err)
				// Continue updating other settings even if one fails
			}
		}

		c.JSON(http.StatusOK, models.SuccessResponse{
			Message: "System settings updated successfully",
			Data:    settingsMap,
		})
	}
}

// GetReports returns system reports
func GetReports(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		startDate := c.Query("startDate")
		endDate := c.Query("endDate")

		reports, err := services.GetReports(db, startDate, endDate)
		if err != nil {
			log.Println("Error fetching reports:", err)
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: "Failed to fetch reports",
			})
			return
		}

		c.JSON(http.StatusOK, models.SuccessResponse{
			Message: "Reports fetched successfully",
			Data:    reports,
		})
	}
}
