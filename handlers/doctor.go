package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"shalby_backend/models"
	"shalby_backend/services"

	"github.com/gin-gonic/gin"
)

// getDoctorIDFromContext retrieves doctor ID from user_id in JWT context
func getDoctorIDFromContext(c *gin.Context, db *sql.DB) (string, error) {
	userID := c.GetString("user_id")
	if userID == "" {
		return "", fmt.Errorf("user_id not found in context")
	}

	var doctorID string
	err := db.QueryRow(
		`SELECT id FROM doctors WHERE user_id = $1 AND is_active = true`,
		userID,
	).Scan(&doctorID)

	if err == sql.ErrNoRows {
		return "", fmt.Errorf("no doctor profile found for this user")
	}
	if err != nil {
		return "", fmt.Errorf("database error: " + err.Error())
	}
	return doctorID, nil
}

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

// GetDoctorProfile returns the doctor's own profile
func GetDoctorProfile(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user_id from auth context
		userID := c.GetString("user_id")
		if userID == "" {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: "Unauthorized: user ID not found",
			})
			return
		}

		// Query doctor by user_id
		// var doctor models.Doctor
		// err := db.QueryRow(
		// 	`SELECT d.id, d.user_id, d.specialization, d.department, d.license_number,
		// 	        d.consultation_fee, d.joining_date, d.is_active, d.created_at
		// 	 FROM doctors d WHERE d.user_id = $1`,
		// 	userID,
		// ).Scan(&doctor.ID, &doctor.UserID, &doctor.Specialization, &doctor.Department,
		// 	&doctor.LicenseNumber, &doctor.ConsultationFee, &doctor.JoiningDate,
		// 	&doctor.IsActive, &doctor.CreatedAt)
		var doctor models.Doctor
		var joiningDate sql.NullString

		err := db.QueryRow(`
			SELECT d.id, d.user_id, u.name, u.email,
				d.specialization, d.department, d.license_number,
				d.consultation_fee, d.joining_date, d.is_active, d.created_at
			FROM doctors d
			JOIN users u ON d.user_id = u.id
			WHERE d.user_id = $1
		`, userID).Scan(
			&doctor.ID, &doctor.UserID, &doctor.Name, &doctor.Email,
			&doctor.Specialization, &doctor.Department, &doctor.LicenseNumber,
			&doctor.ConsultationFee, &joiningDate, &doctor.IsActive, &doctor.CreatedAt,
		)

		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: "Doctor profile not found. Please contact administrator.",
			})
			return
		}

		if err != nil {
			log.Println("Database error fetching doctor profile:", err)
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: "Failed to fetch doctor profile",
			})
			return
		}

		c.JSON(http.StatusOK, models.SuccessResponse{
			Message: "Doctor profile fetched successfully",
			Data:    doctor,
		})
	}
}

// GetDoctorAppointments returns appointments for the doctor
func GetDoctorAppointments(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		if userID == "" {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: "Unauthorized: user ID not found",
			})
			return
		}

		// Get doctor_id from doctors table
		var doctorID string
		err := db.QueryRow("SELECT id FROM doctors WHERE user_id = $1", userID).Scan(&doctorID)
		if err != nil {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: "Doctor profile not found",
			})
			return
		}

		status := c.Query("status")
		date := c.Query("date")
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

		// Build query based on filters
		query := `
			SELECT a.id, a.patient_id, u.name, a.doctor_id, du.name, d.specialization,
			       a.appointment_date, a.appointment_time, a.duration_minutes, a.type,
			       a.status, a.priority, a.chief_complaint, a.cancellation_reason, a.created_at, a.updated_at
			FROM appointments a
			JOIN patients p ON a.patient_id = p.id
			JOIN users u ON p.user_id = u.id
			JOIN doctors d ON a.doctor_id = d.id
			JOIN users du ON d.user_id = du.id
			WHERE a.doctor_id = $1
		`
		args := []interface{}{doctorID}
		argCount := 2

		if status != "" && status != "all" {
			query += ` AND a.status::TEXT = $` + strconv.Itoa(argCount)
			args = append(args, status)
			argCount++
		}

		if date != "" {
			query += ` AND DATE(a.appointment_date) = $` + strconv.Itoa(argCount)
			args = append(args, date)
			argCount++
		}

		query += ` ORDER BY a.appointment_date DESC LIMIT $` + strconv.Itoa(argCount) + ` OFFSET $` + strconv.Itoa(argCount+1)
		args = append(args, limit, offset)

		rows, err := db.Query(query, args...)
		if err != nil {
			log.Println("Error querying appointments:", err)
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: "Failed to fetch appointments",
			})
			return
		}
		defer rows.Close()

		var appointments []models.Appointment
		for rows.Next() {
			var apt models.Appointment
			var chief, cancellation sql.NullString

			if err := rows.Scan(&apt.ID, &apt.PatientID, &apt.PatientName, &apt.DoctorID,
				&apt.DoctorName, &apt.DoctorSpecialty, &apt.AppointmentDate, &apt.AppointmentTime,
				&apt.DurationMinutes, &apt.Type, &apt.Status, &apt.Priority, &chief,
				&cancellation, &apt.CreatedAt, &apt.UpdatedAt); err != nil {
				log.Println("Error scanning appointment:", err)
				continue
			}

			if chief.Valid {
				apt.ChiefComplaint = chief
			}
			if cancellation.Valid {
				apt.CancellationReason = cancellation
			}

			appointments = append(appointments, apt)
		}

		c.JSON(http.StatusOK, models.SuccessResponse{
			Message: "Appointments fetched successfully",
			Data:    appointments,
		})
	}
}

// GetDoctorAppointmentDetail returns detailed info for a single appointment
func GetDoctorAppointmentDetail(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		appointmentID := c.Param("appointmentId")
		userID := c.GetString("user_id")

		// Get doctor_id from doctors table
		var doctorID string
		err := db.QueryRow("SELECT id FROM doctors WHERE user_id = $1", userID).Scan(&doctorID)
		if err != nil {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: "Doctor profile not found",
			})
			return
		}

		var apt models.Appointment
		var chief, cancellation sql.NullString

		err = db.QueryRow(`
			SELECT a.id, a.patient_id, u.name, a.doctor_id, du.name, d.specialization,
			       a.appointment_date, a.appointment_time, a.duration_minutes, a.type,
			       a.status, a.priority, a.chief_complaint, a.cancellation_reason, a.created_at, a.updated_at
			FROM appointments a
			JOIN patients p ON a.patient_id = p.id
			JOIN users u ON p.user_id = u.id
			JOIN doctors d ON a.doctor_id = d.id
			JOIN users du ON d.user_id = du.id
			WHERE a.id = $1 AND a.doctor_id = $2
		`, appointmentID, doctorID).Scan(&apt.ID, &apt.PatientID, &apt.PatientName, &apt.DoctorID,
			&apt.DoctorName, &apt.DoctorSpecialty, &apt.AppointmentDate, &apt.AppointmentTime,
			&apt.DurationMinutes, &apt.Type, &apt.Status, &apt.Priority, &chief,
			&cancellation, &apt.CreatedAt, &apt.UpdatedAt)

		if err != nil {
			log.Println("Error querying appointment:", err)
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: "Appointment not found",
			})
			return
		}

		if chief.Valid {
			apt.ChiefComplaint = chief
		}
		if cancellation.Valid {
			apt.CancellationReason = cancellation
		}

		c.JSON(http.StatusOK, models.SuccessResponse{
			Message: "Appointment fetched successfully",
			Data:    apt,
		})
	}
}

// GetDoctorPatients returns all patients for the doctor
func GetDoctorPatients(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		if userID == "" {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: "Unauthorized: user ID not found",
			})
			return
		}

		// Get doctor_id from doctors table
		var doctorID string
		err := db.QueryRow("SELECT id FROM doctors WHERE user_id = $1", userID).Scan(&doctorID)
		if err != nil {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: "Doctor profile not found",
			})
			return
		}

		rows, err := db.Query(`
			SELECT DISTINCT p.id, u.name, u.email, p.date_of_birth, p.gender, p.blood_group, p.phone
			FROM patients p
			JOIN users u ON p.user_id = u.id
			JOIN appointments a ON a.patient_id = p.id
			WHERE a.doctor_id = $1
			ORDER BY u.name
		`, doctorID)
		if err != nil {
			log.Println("Error querying patients:", err)
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: "Failed to fetch patients",
			})
			return
		}
		defer rows.Close()

		var patients []services.PatientDetail
		for rows.Next() {
			var p services.PatientDetail
			var phone, dob sql.NullString

			if err := rows.Scan(&p.ID, &p.Name, &p.Email, &dob, &p.Gender, &p.Blood, &phone); err != nil {
				log.Println("Error scanning patient:", err)
				continue
			}

			if phone.Valid {
				p.Phone = phone.String
			}
			if dob.Valid {
				p.DateOfBirth = dob.String
			}

			patients = append(patients, p)
		}

		c.JSON(http.StatusOK, models.SuccessResponse{
			Message: "Patients fetched successfully",
			Data:    patients,
		})
	}
}

// GetDoctorDashboardStats returns doctor dashboard statistics
func GetDoctorDashboardStats(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		if userID == "" {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: "Unauthorized: user ID not found",
			})
			return
		}

		// Get doctor_id from doctors table
		var doctorID string
		err := db.QueryRow("SELECT id FROM doctors WHERE user_id = $1", userID).Scan(&doctorID)
		if err != nil {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: "Doctor profile not found",
			})
			return
		}

		stats := gin.H{}

		// pending_appointments_today — scheduled for today
		var pendingToday int
		db.QueryRow(`
			SELECT COUNT(*) FROM appointments
			WHERE doctor_id = $1
			AND DATE(appointment_date) = CURRENT_DATE
			AND status IN ('Scheduled', 'Checked In', 'Ready for Doctor', 'Confirmed')
		`, doctorID).Scan(&pendingToday)
		stats["pending_appointments_today"] = pendingToday

		// patients_seen_this_week — completed this week
		var seenThisWeek int
		db.QueryRow(`
			SELECT COUNT(DISTINCT patient_id) FROM appointments
			WHERE doctor_id = $1
			AND status = 'Completed'
			AND appointment_date >= DATE_TRUNC('week', CURRENT_DATE)
		`, doctorID).Scan(&seenThisWeek)
		stats["patients_seen_this_week"] = seenThisWeek

		// pending_prescriptions — pharmacy status pending
		var pendingRx int
		db.QueryRow(`
			SELECT COUNT(*) FROM prescriptions
			WHERE doctor_id = $1 AND pharmacy_status = 'Pending'
		`, doctorID).Scan(&pendingRx)
		stats["pending_prescriptions"] = pendingRx

		// new_patients_this_month — first appointments this month
		var newThisMonth int
		db.QueryRow(`
			SELECT COUNT(DISTINCT patient_id) FROM appointments
			WHERE doctor_id = $1
			AND appointment_date >= DATE_TRUNC('month', CURRENT_DATE)
		`, doctorID).Scan(&newThisMonth)
		stats["new_patients_this_month"] = newThisMonth

		c.JSON(http.StatusOK, models.SuccessResponse{
			Message: "Dashboard stats fetched successfully",
			Data:    stats,
		})
	}
}
