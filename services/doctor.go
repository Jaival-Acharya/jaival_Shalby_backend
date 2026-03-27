package services

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"shalby_backend/models"

	"golang.org/x/crypto/bcrypt"
)

// generateDefaultPassword creates a password from the first name
func generateDefaultPassword(fullName string) string {
	// Extract first name from full name
	parts := strings.Fields(strings.TrimSpace(fullName))
	if len(parts) == 0 {
		return "User@123"
	}
	firstName := strings.ToLower(parts[0])
	return firstName + "@123"
}

// GetAllDoctors retrieves all doctors with their user information
func GetAllDoctors(db *sql.DB) (*models.DoctorListResponse, error) {
	query := `
	SELECT 
		d.id, d.user_id, u.name, u.email,
		d.specialization, d.department, d.license_number, 
		d.consultation_fee, d.joining_date, d.is_active, d.created_at
	FROM doctors d
	JOIN users u ON d.user_id = u.id
	ORDER BY u.name ASC
	`

	rows, err := db.Query(query)
	if err != nil {
		log.Println("Error querying doctors:", err)
		return nil, err
	}
	defer rows.Close()

	doctors := []models.DoctorResponse{}
	for rows.Next() {
		var doc models.Doctor
		var joinDate sql.NullString

		if err := rows.Scan(
			&doc.ID, &doc.UserID, &doc.Name, &doc.Email,
			&doc.Specialization, &doc.Department, &doc.LicenseNumber,
			&doc.ConsultationFee, &joinDate, &doc.IsActive, &doc.CreatedAt,
		); err != nil {
			log.Println("Error scanning doctor:", err)
			continue
		}

		// TODO: Avatar URL generation - currently not storing avatar in database
		// hash := md5.Sum([]byte(doc.Email))
		// avatarURL := fmt.Sprintf("https://i.pravatar.cc/150?u=%x", hash)

		doctors = append(doctors, models.DoctorResponse{
			ID:              doc.ID,
			Name:            doc.Name,
			Email:           doc.Email,
			Specialization:  doc.Specialization,
			Department:      doc.Department,
			LicenseNumber:   doc.LicenseNumber,
			ConsultationFee: doc.ConsultationFee,
			JoiningDate:     joinDate.String,
			IsActive:        doc.IsActive,
			// AvatarURL:       avatarURL, // TODO: Avatar storage not implemented yet
			CreatedAt: doc.CreatedAt,
		})
	}

	return &models.DoctorListResponse{
		Doctors: doctors,
		Total:   len(doctors),
	}, nil
}

// CreateDoctor creates a new doctor with user account
func CreateDoctor(db *sql.DB, req models.DoctorRequest) (*models.DoctorResponse, error) {
	// Auto-generate password if not provided
	password := req.Password
	if password == "" {
		password = generateDefaultPassword(req.Name)
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Println("Error hashing password:", err)
		return nil, fmt.Errorf("failed to hash password")
	}

	// TODO: Generate avatar URL - currently not storing avatar in database
	// hash := md5.Sum([]byte(req.Email))
	// avatarURL := fmt.Sprintf("https://i.pravatar.cc/150?u=%x", hash)

	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		log.Println("Error starting transaction:", err)
		return nil, err
	}
	defer tx.Rollback()

	// Create user
	var userID string
	userQuery := `
	INSERT INTO users (name, email, password_hash, phone, address, is_active)
	VALUES ($1, $2, $3, $4, $5, true)
	RETURNING id
	`

	err = tx.QueryRow(userQuery, req.Name, req.Email, string(hashedPassword), req.Phone, req.Address).Scan(&userID)
	if err != nil {
		log.Println("Error creating user:", err)
		return nil, fmt.Errorf("failed to create user account")
	}

	// Assign Doctor role
	roleQuery := `
	INSERT INTO user_roles (user_id, role_id)
	VALUES ($1, (SELECT id FROM roles WHERE name = 'Doctor'))
	`

	_, err = tx.Exec(roleQuery, userID)
	if err != nil {
		log.Println("Error assigning doctor role:", err)
		return nil, fmt.Errorf("failed to assign doctor role")
	}

	// Create doctor record
	var doctorID string
	doctorQuery := `
	INSERT INTO doctors (user_id, specialization, department, license_number, consultation_fee, joining_date, is_active)
	VALUES ($1, $2, $3, $4, $5, NULLIF($6, '')::date, true)
	RETURNING id
	`

	err = tx.QueryRow(doctorQuery,
		userID, req.Specialization, req.Department, req.LicenseNumber, req.ConsultationFee, req.JoiningDate,
	).Scan(&doctorID)
	if err != nil {
		log.Println("Error creating doctor:", err)
		return nil, fmt.Errorf("failed to create doctor record")
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		log.Println("Error committing transaction:", err)
		return nil, err
	}

	return &models.DoctorResponse{
		ID:              doctorID,
		Name:            req.Name,
		Email:           req.Email,
		Specialization:  req.Specialization,
		Department:      req.Department,
		LicenseNumber:   req.LicenseNumber,
		ConsultationFee: req.ConsultationFee,
		IsActive:        true,
		// AvatarURL:       avatarURL, // TODO: Avatar storage not implemented yet
		CreatedAt: time.Now(),
	}, nil
}

// UpdateDoctor updates doctor information
func UpdateDoctor(db *sql.DB, doctorID string, req models.DoctorUpdateRequest) (*models.DoctorResponse, error) {
	// Get the user_id for this doctor first
	var userID string
	err := db.QueryRow(`SELECT user_id FROM doctors WHERE id = $1`, doctorID).Scan(&userID)
	if err != nil {
		log.Println("Error finding doctor:", err)
		return nil, fmt.Errorf("doctor not found")
	}

	// Update user table (name, email, phone, address)
	userQuery := `UPDATE users SET name = $1, email = $2, phone = $3, address = $4 WHERE id = $5`
	_, err = db.Exec(userQuery, req.Name, req.Email, req.Phone, req.Address, userID)
	if err != nil {
		log.Println("Error updating user:", err)
		return nil, fmt.Errorf("failed to update user information")
	}

	// Update doctor table
	doctorQuery := `
	UPDATE doctors 
	SET specialization = $1, department = $2, license_number = $3, consultation_fee = $4, joining_date = NULLIF($5, '')::date
	WHERE id = $6
	`

	_, err = db.Exec(doctorQuery, req.Specialization, req.Department, req.LicenseNumber, req.ConsultationFee, req.JoiningDate, doctorID)
	if err != nil {
		log.Println("Error updating doctor:", err)
		return nil, fmt.Errorf("failed to update doctor")
	}

	return GetDoctorByIDForResponse(db, doctorID)
}

// GetDoctorByID retrieves a specific doctor by ID with education and work experience
func GetDoctorByID(db *sql.DB, doctorID string) (*models.DoctorDetailResponse, error) {
	query := `
	SELECT 
		d.id, d.user_id, u.name, u.email, u.phone, u.address,
		d.specialization, d.department, d.license_number,
		d.consultation_fee, d.joining_date, d.is_active, d.created_at
	FROM doctors d
	JOIN users u ON d.user_id = u.id
	WHERE d.id = $1
	`

	var doc models.Doctor
	var phone, address sql.NullString
	var joinDate sql.NullString

	err := db.QueryRow(query, doctorID).Scan(
		&doc.ID, &doc.UserID, &doc.Name, &doc.Email, &phone, &address,
		&doc.Specialization, &doc.Department, &doc.LicenseNumber,
		&doc.ConsultationFee, &joinDate, &doc.IsActive, &doc.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("doctor not found")
	}
	if err != nil {
		log.Println("Error querying doctor:", err)
		return nil, err
	}

	// Fetch education and work experience
	education, errEdu := GetDoctorEducation(db, doctorID)
	if errEdu != nil && errEdu != sql.ErrNoRows {
		log.Println("Error fetching education:", errEdu)
	}

	workExperience, errExp := GetDoctorWorkExperience(db, doctorID)
	if errExp != nil && errExp != sql.ErrNoRows {
		log.Println("Error fetching work experience:", errExp)
	}

	response := &models.DoctorDetailResponse{
		ID:              doc.ID,
		Name:            doc.Name,
		Email:           doc.Email,
		Phone:           phone.String,
		Address:         address.String,
		Specialization:  doc.Specialization,
		Department:      doc.Department,
		LicenseNumber:   doc.LicenseNumber,
		ConsultationFee: doc.ConsultationFee,
		JoiningDate:     joinDate.String,
		IsActive:        doc.IsActive,
		CreatedAt:       doc.CreatedAt,
		Education:       education,
		WorkExperience:  workExperience,
	}

	return response, nil
}

// GetDoctorByIDForResponse retrieves doctor data formatted as response
func GetDoctorByIDForResponse(db *sql.DB, doctorID string) (*models.DoctorResponse, error) {
	detailDoc, err := GetDoctorByID(db, doctorID)
	if err != nil {
		return nil, err
	}

	return &models.DoctorResponse{
		ID:              detailDoc.ID,
		Name:            detailDoc.Name,
		Email:           detailDoc.Email,
		Specialization:  detailDoc.Specialization,
		Department:      detailDoc.Department,
		LicenseNumber:   detailDoc.LicenseNumber,
		ConsultationFee: detailDoc.ConsultationFee,
		JoiningDate:     detailDoc.JoiningDate,
		IsActive:        detailDoc.IsActive,
		// AvatarURL:       detailDoc.AvatarURL, // TODO: Avatar storage not implemented yet
		CreatedAt: detailDoc.CreatedAt,
	}, nil
}

// DeleteDoctor soft deletes a doctor
func DeleteDoctor(db *sql.DB, doctorID string) error {
	query := `UPDATE doctors SET is_active = false WHERE id = $1`
	_, err := db.Exec(query, doctorID)
	if err != nil {
		log.Println("Error deleting doctor:", err)
		return fmt.Errorf("failed to delete doctor")
	}
	return nil
}

// GetDoctorSchedules retrieves all schedules for a doctor
func GetDoctorSchedules(db *sql.DB, doctorID string) ([]models.DoctorSchedule, error) {
	query := `
	SELECT id, doctor_id, day_of_week, start_time, end_time, 
	       slot_duration_minutes, max_patients_per_slot, is_active
	FROM doctor_schedules
	WHERE doctor_id = $1 AND is_active = true
	ORDER BY day_of_week ASC
	`

	rows, err := db.Query(query, doctorID)
	if err != nil {
		log.Println("Error querying schedules:", err)
		return nil, err
	}
	defer rows.Close()

	schedules := []models.DoctorSchedule{}
	for rows.Next() {
		var schedule models.DoctorSchedule
		if err := rows.Scan(
			&schedule.ID, &schedule.DoctorID, &schedule.DayOfWeek,
			&schedule.StartTime, &schedule.EndTime,
			&schedule.SlotDurationMinutes, &schedule.MaxPatientsPerSlot,
			&schedule.IsActive,
		); err != nil {
			log.Println("Error scanning schedule:", err)
			continue
		}
		schedules = append(schedules, schedule)
	}

	return schedules, nil
}

// CreateDoctorSchedule creates a new schedule for a doctor
func CreateDoctorSchedule(db *sql.DB, doctorID string, req models.DoctorScheduleRequest) (*models.DoctorSchedule, error) {
	query := `
	INSERT INTO doctor_schedules 
	(doctor_id, day_of_week, start_time, end_time, slot_duration_minutes, max_patients_per_slot, is_active)
	VALUES ($1, $2, $3, $4, $5, $6, true)
	RETURNING id, doctor_id, day_of_week, start_time, end_time, slot_duration_minutes, max_patients_per_slot, is_active
	`

	var schedule models.DoctorSchedule
	err := db.QueryRow(
		query, doctorID, req.DayOfWeek, req.StartTime, req.EndTime,
		req.SlotDurationMinutes, req.MaxPatientsPerSlot,
	).Scan(
		&schedule.ID, &schedule.DoctorID, &schedule.DayOfWeek,
		&schedule.StartTime, &schedule.EndTime,
		&schedule.SlotDurationMinutes, &schedule.MaxPatientsPerSlot,
		&schedule.IsActive,
	)

	if err != nil {
		log.Println("Error creating schedule:", err)
		return nil, fmt.Errorf("failed to create schedule")
	}

	return &schedule, nil
}

// GetDoctorLeaves retrieves all leaves for a doctor
func GetDoctorLeaves(db *sql.DB, doctorID string) ([]models.DoctorLeave, error) {
	query := `
	SELECT id, doctor_id, start_date, end_date, reason, created_at
	FROM doctor_leave
	WHERE doctor_id = $1
	ORDER BY start_date DESC
	`

	rows, err := db.Query(query, doctorID)
	if err != nil {
		log.Println("Error querying leaves:", err)
		return nil, err
	}
	defer rows.Close()

	leaves := []models.DoctorLeave{}
	for rows.Next() {
		var leave models.DoctorLeave
		if err := rows.Scan(
			&leave.ID, &leave.DoctorID, &leave.StartDate,
			&leave.EndDate, &leave.Reason, &leave.CreatedAt,
		); err != nil {
			log.Println("Error scanning leave:", err)
			continue
		}
		leaves = append(leaves, leave)
	}

	return leaves, nil
}

// CreateDoctorLeave creates a new leave period for a doctor
func CreateDoctorLeave(db *sql.DB, doctorID string, req models.DoctorLeaveRequest) (*models.DoctorLeave, error) {
	query := `
	INSERT INTO doctor_leave (doctor_id, start_date, end_date, reason)
	VALUES ($1, $2::date, $3::date, $4)
	RETURNING id, doctor_id, start_date, end_date, reason, created_at
	`

	var leave models.DoctorLeave
	err := db.QueryRow(query, doctorID, req.StartDate, req.EndDate, req.Reason).Scan(
		&leave.ID, &leave.DoctorID, &leave.StartDate,
		&leave.EndDate, &leave.Reason, &leave.CreatedAt,
	)

	if err != nil {
		log.Println("Error creating leave:", err)
		return nil, fmt.Errorf("failed to create leave")
	}

	return &leave, nil
}

// GetDoctorEducation retrieves education details for a doctor
func GetDoctorEducation(db *sql.DB, doctorID string) ([]map[string]interface{}, error) {
	query := `
	SELECT id, degree, school_university, year_graduated 
	FROM doctor_education 
	WHERE doctor_id = $1
	ORDER BY year_graduated DESC
	`

	rows, err := db.Query(query, doctorID)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	defer rows.Close()

	var education []map[string]interface{}
	for rows.Next() {
		var id, degree, school string
		var year sql.NullInt64

		if err := rows.Scan(&id, &degree, &school, &year); err != nil {
			continue
		}

		edu := map[string]interface{}{
			"id":                id,
			"degree":            degree,
			"school":            school,
			"school_university": school,
		}
		if year.Valid {
			edu["year"] = year.Int64
		}

		education = append(education, edu)
	}

	return education, nil
}

// GetDoctorWorkExperience retrieves work experience for a doctor
func GetDoctorWorkExperience(db *sql.DB, doctorID string) ([]map[string]interface{}, error) {
	query := `
	SELECT id, position, hospital_institution, period_text, start_year, end_year
	FROM doctor_work_experience 
	WHERE doctor_id = $1
	ORDER BY end_year DESC NULLS FIRST
	`

	rows, err := db.Query(query, doctorID)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	defer rows.Close()

	var experience []map[string]interface{}
	for rows.Next() {
		var id, position, hospital, periodText string
		var startYear, endYear sql.NullInt64

		if err := rows.Scan(&id, &position, &hospital, &periodText, &startYear, &endYear); err != nil {
			continue
		}

		exp := map[string]interface{}{
			"id":       id,
			"position": position,
			"place":    hospital,
			"period":   periodText,
		}
		if startYear.Valid {
			exp["start_year"] = startYear.Int64
		}
		if endYear.Valid {
			exp["end_year"] = endYear.Int64
		}

		experience = append(experience, exp)
	}

	return experience, nil
}
