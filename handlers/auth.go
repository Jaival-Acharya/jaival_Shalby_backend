package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"shalby_backend/config"
	"shalby_backend/models"
	"shalby_backend/services"

	"github.com/gin-gonic/gin"
)

// In-memory OTP storage with expiration
type OTPEntry struct {
	OTP       string
	ExpiresAt time.Time
}

var (
	otpStore = make(map[string]OTPEntry)
	otpMutex = &sync.Mutex{}
)

// CleanExpiredOTPs removes expired OTPs from memory
func CleanExpiredOTPs() {
	otpMutex.Lock()
	defer otpMutex.Unlock()

	now := time.Now()
	for email, entry := range otpStore {
		if now.After(entry.ExpiresAt) {
			delete(otpStore, email)
		}
	}
}

func generateSixDigitOTP() string {
	seed := time.Now().UnixNano() % 900000
	otp := int(seed) + 100000
	return strconv.Itoa(otp)
}

// LoginHandler handles user login
func LoginHandler(db *sql.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.LoginRequest

		// Bind and validate JSON
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error: "Invalid request format: " + err.Error(),
			})
			return
		}

		// Call login service
		response, err := services.Login(db, cfg, req)
		if err != nil {
			log.Println("Login error:", err.Error())
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, response)
	}
}

// SignupHandler handles patient registration
func SignupHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.SignupRequest

		// Bind and validate JSON
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error: "Invalid request format: " + err.Error(),
			})
			return
		}

		// Call signup service
		response, err := services.Signup(db, req)
		if err != nil {
			log.Println("Signup error:", err.Error())

			// Check if it's a conflict error (email exists)
			if err.Error() == "An account with this email already exists" {
				c.JSON(http.StatusConflict, models.ErrorResponse{
					Error: err.Error(),
				})
				return
			}

			// Other errors return 500
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: "Registration failed",
			})
			return
		}

		c.JSON(http.StatusCreated, response)
	}
}

// LogoutHandler handles user logout requests
func LogoutHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := services.Logout(); err != nil {
			log.Println("Logout error:", err.Error())
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: "Logout failed",
			})
			return
		}

		c.JSON(http.StatusOK, models.SuccessResponse{
			Message: "Logout successful",
		})
	}
}

func RequestPasswordResetHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.PasswordResetRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error: "Invalid request format: " + err.Error(),
			})
			return
		}

		var exists bool
		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", req.Email).Scan(&exists)
		if err != nil {
			log.Println("Password reset user lookup error:", err.Error())
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to process request"})
			return
		}

		if !exists {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "No account found with this email"})
			return
		}

		// Generate OTP
		otp := generateSixDigitOTP()
		expiresAt := time.Now().Add(5 * time.Minute)

		// Store in memory
		otpMutex.Lock()
		otpStore[req.Email] = OTPEntry{
			OTP:       otp,
			ExpiresAt: expiresAt,
		}
		otpMutex.Unlock()

		log.Printf("Password reset OTP for %s: %s", req.Email, otp)

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": fmt.Sprintf("OTP generated successfully. Use this OTP for reset: %s", otp),
		})
	}
}

func VerifyPasswordOTPHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.VerifyPasswordOTPRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error: "Invalid request format: " + err.Error(),
			})
			return
		}

		// Clean expired OTPs
		CleanExpiredOTPs()

		// Retrieve from memory
		otpMutex.Lock()
		entry, exists := otpStore[req.Email]
		otpMutex.Unlock()

		if !exists {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Invalid OTP"})
			return
		}

		if time.Now().After(entry.ExpiresAt) {
			otpMutex.Lock()
			delete(otpStore, req.Email)
			otpMutex.Unlock()
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "OTP has expired"})
			return
		}

		if entry.OTP != req.OTP {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Invalid OTP"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "OTP verified successfully",
		})
	}
}

func ResetPasswordHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.ResetPasswordRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error: "Invalid request format: " + err.Error(),
			})
			return
		}

		// Clean expired OTPs
		CleanExpiredOTPs()

		// Retrieve from memory
		otpMutex.Lock()
		entry, exists := otpStore[req.Email]
		otpMutex.Unlock()

		if !exists {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Invalid OTP"})
			return
		}

		if time.Now().After(entry.ExpiresAt) {
			otpMutex.Lock()
			delete(otpStore, req.Email)
			otpMutex.Unlock()
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "OTP has expired"})
			return
		}

		if entry.OTP != req.OTP {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Invalid OTP"})
			return
		}

		// Decrypt and hash the new password
		decryptedPassword, err := services.DecryptPassword(req.Password)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid password encryption"})
			return
		}

		hashedPassword, err := services.HashPassword(decryptedPassword)
		if err != nil {
			log.Println("Reset password hashing error:", err.Error())
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to reset password"})
			return
		}

		// Update password in database
		_, err = db.Exec("UPDATE users SET password_hash = $1 WHERE email = $2", hashedPassword, req.Email)
		if err != nil {
			log.Println("Reset password update error:", err.Error())
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to reset password"})
			return
		}

		// Clean up OTP after successful reset
		otpMutex.Lock()
		delete(otpStore, req.Email)
		otpMutex.Unlock()

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Password updated successfully",
		})
	}
}
