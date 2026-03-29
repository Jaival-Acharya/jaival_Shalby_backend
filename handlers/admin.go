package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"shalby_backend/models"
	"shalby_backend/services"

	"github.com/gin-gonic/gin"
)

// GetAdminDashboard returns admin dashboard statistics
func GetAdminDashboard(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		adminService := services.NewAdminService(db)
		stats, err := adminService.GetDashboardStats()
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
		adminService := services.NewAdminService(db)
		settingsArray, err := adminService.GetSystemSettings()
		if err != nil {
			log.Println("Error fetching system settings:", err)
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: "Failed to fetch system settings",
			})
			return
		}

		log.Printf("DEBUG: Retrieved %d settings from database\n", len(settingsArray))

		// If no settings found, seed defaults
		if len(settingsArray) == 0 {
			log.Println("DEBUG: No settings found, seeding defaults...")
			defaultSettingsMap := map[string]string{
				"hospital_name":       "Shalby Hospital",
				"hospital_email":      "info@shalby.hospital",
				"hospital_phone":      "+91-8000-XXX-XXX",
				"hospital_address":    "Your Hospital Address",
				"hospital_website":    "https://shalby.hospital",
				"currency":            "INR",
				"timezone":            "Asia/Kolkata",
				"language":            "English",
				"notifications_email": "true",
				"notifications_sms":   "false",
				"notifications_push":  "true",
				"security_twoFactor":  "false",
				"security_autoLogout": "30",
				"mobile_offlineMode":  "true",
			}

			for key, value := range defaultSettingsMap {
				err := adminService.UpdateSystemSetting(key, value)
				if err != nil {
					log.Printf("ERROR: Failed to seed default setting %s: %v\n", key, err)
				}
			}

			// Re-fetch settings after seeding
			settingsArray, _ = adminService.GetSystemSettings()
			log.Printf("DEBUG: After seeding, now have %d settings\n", len(settingsArray))
		}

		// Convert array of settings to a flat map/object structure
		settingsMap := make(map[string]interface{})
		for _, setting := range settingsArray {
			settingsMap[setting.Key] = setting.Value
			log.Printf("DEBUG: Adding to response - Key: %s, Value: %s\n", setting.Key, setting.Value)
		}

		log.Printf("DEBUG: Final response map has %d entries\n", len(settingsMap))
		c.JSON(http.StatusOK, models.SuccessResponse{
			Message: "System settings fetched successfully",
			Data:    settingsMap,
		})
	}
}

// UpdateSystemSettings updates system settings from the admin form
func UpdateSystemSettings(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.SystemSettingsRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			log.Printf("ERROR: Invalid request format: %v\n", err)
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error: "Invalid request format: " + err.Error(),
			})
			return
		}

		log.Printf("DEBUG: Received request to update settings\n")
		log.Printf("DEBUG: Hospital Name: %s\n", req.HospitalName)
		log.Printf("DEBUG: Email: %s\n", req.Email)

		// Convert the structured request to individual settings updates with snake_case keys
		settingsMap := map[string]string{
			"hospital_name":       req.HospitalName,
			"hospital_email":      req.Email,
			"hospital_phone":      req.Phone,
			"hospital_address":    req.Address,
			"hospital_website":    req.Website,
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
		adminService := services.NewAdminService(db)
		var updateErrors []string
		for key, value := range settingsMap {
			log.Printf("DEBUG: Updating %s = %s\n", key, value)
			err := adminService.UpdateSystemSetting(key, value)
			if err != nil {
				log.Println("Error updating setting "+key+":", err)
				updateErrors = append(updateErrors, key+": "+err.Error())
				// Continue updating other settings even if one fails
			} else {
				log.Printf("DEBUG: Successfully updated %s\n", key)
			}
		}

		// If there were errors, return 400 Bad Request with error details
		if len(updateErrors) > 0 {
			log.Printf("DEBUG: %d settings failed to update\n", len(updateErrors))
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Some settings failed to update",
				"details": updateErrors,
				"data":    settingsMap,
			})
			return
		}

		log.Printf("DEBUG: All settings updated successfully, returning response\n")
		c.JSON(http.StatusOK, models.SuccessResponse{
			Message: "System settings updated successfully",
			Data:    settingsMap,
		})
	}
}

// GetAdminProfile returns the admin user's profile information
func GetAdminProfile(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		adminService := services.NewAdminService(db)
		settingsArray, err := adminService.GetSystemSettings()

		if err != nil {
			log.Printf("ERROR: Failed to get system settings: %v, using defaults\n", err)
			// If there's an error, return defaults
			adminProfile := gin.H{
				"id":         "ADM001",
				"name":       "System Administrator",
				"email":      "admin@shalby.hospital",
				"phone":      "+91-8000-XXX-XXX",
				"address":    "Shalby Hospital, Address",
				"dob":        "",
				"bio":        "Senior Healthcare Administrator with expertise in hospital management systems.",
				"role":       "Administrator",
				"department": "Administration",
				"status":     "Active",
				"joinDate":   "2024-01-01",
			}

			c.JSON(http.StatusOK, models.SuccessResponse{
				Message: "Admin profile fetched successfully",
				Data:    adminProfile,
			})
			return
		}

		// Convert array of settings to a map
		settingsMap := make(map[string]string)
		for _, setting := range settingsArray {
			settingsMap[setting.Key] = setting.Value
		}

		// Build admin profile from settings
		adminProfile := gin.H{
			"id":         "ADM001",
			"name":       getSettingOrDefault(settingsMap, "admin_name", "System Administrator"),
			"email":      getSettingOrDefault(settingsMap, "admin_email", "admin@shalby.hospital"),
			"phone":      getSettingOrDefault(settingsMap, "admin_phone", "+91-8000-XXX-XXX"),
			"address":    getSettingOrDefault(settingsMap, "admin_address", "Shalby Hospital, Address"),
			"dob":        getSettingOrDefault(settingsMap, "admin_dob", ""),
			"bio":        getSettingOrDefault(settingsMap, "admin_bio", "Senior Healthcare Administrator with expertise in hospital management systems."),
			"role":       "Administrator",
			"department": "Administration",
			"status":     "Active",
			"joinDate":   "2024-01-01",
		}

		c.JSON(http.StatusOK, models.SuccessResponse{
			Message: "Admin profile fetched successfully",
			Data:    adminProfile,
		})
	}
}

// getSettingOrDefault returns a setting value or a default if not found
func getSettingOrDefault(settings map[string]string, key string, defaultValue string) string {
	if value, ok := settings[key]; ok && value != "" {
		return value
	}
	return defaultValue
}

// UpdateAdminProfile updates the admin user's profile information
func UpdateAdminProfile(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var updateData map[string]interface{}
		if err := c.BindJSON(&updateData); err != nil {
			log.Println("Error parsing profile update request:", err)
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error: "Invalid request data",
			})
			return
		}

		log.Printf("DEBUG: Updating admin profile with data: %v\n", updateData)
		adminService := services.NewAdminService(db)

		// Map field names to setting keys
		fieldToSettingKey := map[string]string{
			"name":    "admin_name",
			"email":   "admin_email",
			"phone":   "admin_phone",
			"address": "admin_address",
			"dob":     "admin_dob",
			"bio":     "admin_bio",
		}

		// Update each field as a system setting
		for field, settingKey := range fieldToSettingKey {
			if value, ok := updateData[field]; ok {
				valueStr := fmt.Sprintf("%v", value)
				err := adminService.UpdateSystemSetting(settingKey, valueStr)
				if err != nil {
					log.Printf("ERROR: Failed to update setting %s: %v\n", settingKey, err)
				}
			}
		}

		// Fetch updated settings
		settingsArray, err := adminService.GetSystemSettings()
		if err != nil {
			log.Println("Error fetching updated settings:", err)
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: "Failed to update profile",
			})
			return
		}

		// Convert to map
		settingsMap := make(map[string]string)
		for _, setting := range settingsArray {
			settingsMap[setting.Key] = setting.Value
		}

		// Build response with updated profile
		updatedProfile := gin.H{
			"id":         "ADM001",
			"name":       getSettingOrDefault(settingsMap, "admin_name", "System Administrator"),
			"email":      getSettingOrDefault(settingsMap, "admin_email", "admin@shalby.hospital"),
			"phone":      getSettingOrDefault(settingsMap, "admin_phone", "+91-8000-XXX-XXX"),
			"address":    getSettingOrDefault(settingsMap, "admin_address", "Shalby Hospital, Address"),
			"dob":        getSettingOrDefault(settingsMap, "admin_dob", ""),
			"bio":        getSettingOrDefault(settingsMap, "admin_bio", "Senior Healthcare Administrator with expertise in hospital management systems."),
			"role":       "Administrator",
			"department": "Administration",
			"status":     "Active",
			"joinDate":   "2024-01-01",
		}

		c.JSON(http.StatusOK, models.SuccessResponse{
			Message: "Admin profile updated successfully",
			Data:    updatedProfile,
		})
	}
}

// GetReports returns system reports
func GetReports(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		adminService := services.NewAdminService(db)
		reports, err := adminService.GetReports("", 100, 0)
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
