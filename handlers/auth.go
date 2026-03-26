package handlers

import (
	"database/sql"
	"log"
	"net/http"

	"shalby_backend/config"
	"shalby_backend/models"
	"shalby_backend/services"

	"github.com/gin-gonic/gin"
)

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
