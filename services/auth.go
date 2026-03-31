package services

import (
	"database/sql"
	"encoding/base64"
	"errors"
	"log"
	"time"

	"shalby_backend/config"
	"shalby_backend/models"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Login authenticates a user and returns a JWT token
func Login(db *sql.DB, cfg *config.Config, req models.LoginRequest) (*models.LoginResponse, error) {
	// Step 1: Query user by email
	var user models.User
	err := db.QueryRow(
		"SELECT u.id, u.name, u.email, u.password_hash, u.is_active FROM users u WHERE u.email = $1 AND u.is_active = true",
		req.Email,
	).Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.IsActive)

	if err == sql.ErrNoRows {
		return nil, errors.New("Invalid email or password")
	}
	if err != nil {
		log.Println("Database error during login:", err)
		return nil, errors.New("Invalid email or password")
	}

	// Step 2: Decode base64-encoded password
	decodedPassword, err := base64.StdEncoding.DecodeString(req.Password)
	if err != nil {
		log.Println("Error decoding password:", err)
		return nil, errors.New("Invalid request format")
	}

	// Step 3-4: Compare passwords using bcrypt
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), decodedPassword)
	if err != nil {
		return nil, errors.New("Invalid email or password")
	}

	// Step 5: Query user's roles
	rows, err := db.Query(
		"SELECT r.name FROM roles r JOIN user_roles ur ON r.id = ur.role_id WHERE ur.user_id = $1",
		user.ID,
	)
	if err != nil {
		log.Println("Error querying roles:", err)
		return nil, errors.New("Failed to retrieve user roles")
	}
	defer rows.Close()

	var roles []string
	for rows.Next() {
		var role string
		if err := rows.Scan(&role); err != nil {
			log.Println("Error scanning role:", err)
			return nil, errors.New("Failed to retrieve user roles")
		}
		roles = append(roles, role)
	}

	user.Roles = roles

	// Step 6: Determine active role (use first role available)
	activeRole := "Patient"
	if len(roles) > 0 {
		activeRole = roles[0]
	}

	// Step 7: Generate JWT token
	expirationTime := time.Now().Add(time.Duration(cfg.JWTExpiry) * time.Hour)
	claims := jwt.MapClaims{
		"sub":         user.ID,
		"name":        user.Name,
		"email":       user.Email,
		"roles":       roles,
		"active_role": activeRole,
		"exp":         expirationTime.Unix(),
		"iat":         time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(cfg.JWTSecret))
	if err != nil {
		log.Println("Error generating token:", err)
		return nil, errors.New("Failed to generate authentication token")
	}

	// Step 8: Return response
	return &models.LoginResponse{
		Token: tokenString,
		User: models.User{
			ID:        user.ID,
			Name:      user.Name,
			Email:     user.Email,
			IsActive:  user.IsActive,
			CreatedAt: user.CreatedAt,
			Roles:     roles,
		},
	}, nil
}

// Signup creates a new patient account
func Signup(db *sql.DB, req models.SignupRequest) (*models.SignupResponse, error) {
	// Step 1: Check if email already exists
	var existingID string
	err := db.QueryRow("SELECT id FROM users WHERE email = $1", req.Email).Scan(&existingID)
	if err == nil {
		return nil, errors.New("An account with this email already exists")
	}
	if err != sql.ErrNoRows {
		log.Println("Error checking email existence:", err)
		return nil, errors.New("Registration failed")
	}

	// Step 2: Decode base64-encoded password
	decodedPassword, err := base64.StdEncoding.DecodeString(req.Password)
	if err != nil {
		log.Println("Error decoding password:", err)
		return nil, errors.New("Invalid request format")
	}

	// Step 3: Hash password with bcrypt
	passwordHash, err := bcrypt.GenerateFromPassword(decodedPassword, bcrypt.DefaultCost)
	if err != nil {
		log.Println("Error hashing password:", err)
		return nil, errors.New("Registration failed")
	}

	// Step 4: Start transaction
	tx, err := db.Begin()
	if err != nil {
		log.Println("Error starting transaction:", err)
		return nil, errors.New("Registration failed")
	}
	defer tx.Rollback()

	// Step 5: Insert into users table
	var userID string
	err = tx.QueryRow(
		"INSERT INTO users (name, email, password_hash) VALUES ($1, $2, $3) RETURNING id",
		req.FullName, req.Email, string(passwordHash),
	).Scan(&userID)
	if err != nil {
		log.Println("Error inserting user:", err)
		return nil, errors.New("Registration failed")
	}

	// Step 6: Get Patient role ID
	var roleID int
	err = tx.QueryRow("SELECT id FROM roles WHERE name = 'Patient'").Scan(&roleID)
	if err != nil {
		log.Println("Error getting Patient role:", err)
		return nil, errors.New("Registration failed")
	}

	// Step 7: Insert into user_roles
	_, err = tx.Exec("INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2)", userID, roleID)
	if err != nil {
		log.Println("Error inserting user_role:", err)
		return nil, errors.New("Registration failed")
	}

	// Step 8: Insert into patients table
	var patientID string
	err = tx.QueryRow(
		"INSERT INTO patients (user_id, date_of_birth, gender, blood_group, phone) VALUES ($1, $2, $3, $4, $5) RETURNING id",
		userID, req.DateOfBirth, req.Gender, req.BloodGroup, req.Phone,
	).Scan(&patientID)
	if err != nil {
		log.Println("Error inserting patient:", err)
		return nil, errors.New("Registration failed")
	}

	// Step 9: Insert allergies
	if len(req.Allergies) > 0 {
		for _, allergy := range req.Allergies {
			_, err := tx.Exec(
				"INSERT INTO patient_allergies (patient_id, allergy) VALUES ($1, $2)",
				patientID, allergy,
			)
			if err != nil {
				log.Println("Error inserting allergy:", err)
				return nil, errors.New("Registration failed")
			}
		}
	}

	// Step 10: Insert chronic conditions
	if len(req.Conditions) > 0 {
		for _, condition := range req.Conditions {
			_, err := tx.Exec(
				"INSERT INTO patient_chronic_conditions (patient_id, condition) VALUES ($1, $2)",
				patientID, condition,
			)
			if err != nil {
				log.Println("Error inserting condition:", err)
				return nil, errors.New("Registration failed")
			}
		}
	}

	// Step 11: Insert emergency contact if provided
	if req.EmergencyContactName != "" {
		_, err := tx.Exec(
			"INSERT INTO patient_emergency_contacts (patient_id, name, phone, relation) VALUES ($1, $2, $3, $4)",
			patientID, req.EmergencyContactName, req.EmergencyContactPhone, req.EmergencyContactRelation,
		)
		if err != nil {
			log.Println("Error inserting emergency contact:", err)
			return nil, errors.New("Registration failed")
		}
	}

	// Step 12: Commit transaction
	err = tx.Commit()
	if err != nil {
		log.Println("Error committing transaction:", err)
		return nil, errors.New("Registration failed")
	}

	// Step 13: Return success response
	return &models.SignupResponse{
		Message: "Registration successful",
		UserID:  userID,
	}, nil
}

// Logout handles logout requests. For JWT stateless auth this confirms client-side token invalidation.
func Logout() error {
	return nil
}

// DecryptPassword decodes the base64-encoded password sent by the frontend.
func DecryptPassword(encodedPassword string) ([]byte, error) {
	decodedPassword, err := base64.StdEncoding.DecodeString(encodedPassword)
	if err != nil {
		return nil, errors.New("invalid request format")
	}

	return decodedPassword, nil
}

// HashPassword hashes plain/decrypted password bytes using bcrypt.
func HashPassword(password []byte) (string, error) {
	passwordHash, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
	if err != nil {
		return "", errors.New("failed to hash password")
	}

	return string(passwordHash), nil
}
