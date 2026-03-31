package services

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// AdminService handles business logic for admin operations
type AdminService struct {
	db *sql.DB
}

// NewAdminService creates a new admin service instance
func NewAdminService(db *sql.DB) *AdminService {
	return &AdminService{db: db}
}

// DashboardStats holds aggregated system statistics
type DashboardStats struct {
	TotalPatients         int        `json:"total_patients"`
	TotalDoctors          int        `json:"total_doctors"`
	TodayAppointments     int        `json:"today_appointments"`
	CheckedInPatients     int        `json:"checked_in_patients"`
	AvailableBeds         int        `json:"available_beds"`
	OccupiedBeds          int        `json:"occupied_beds"`
	PendingAppointments   int        `json:"pending_appointments"`
	CompletedAppointments int        `json:"completed_appointments"`
	AveragePatientsPerDay float64    `json:"average_patients_per_day"`
	SystemUptime          string     `json:"system_uptime"`
	LastBackupTime        *time.Time `json:"last_backup_time"`
}

// AdminProfile represents admin user profile information
type AdminProfile struct {
	ID         string `json:"id"`
	AdminID    string `json:"admin_id"`
	Name       string `json:"name"`
	Email      string `json:"email"`
	Phone      string `json:"phone"`
	Address    string `json:"address"`
	DOB        string `json:"dob"`
	Bio        string `json:"bio"`
	Role       string `json:"role"`
	Department string `json:"department"`
	Status     string `json:"status"`
	JoinDate   string `json:"join_date"`
}

// GetDashboardStats returns overall system statistics with fallback values
func (as *AdminService) GetDashboardStats() (*DashboardStats, error) {
	stats := &DashboardStats{}

	// Total Patients - fallback: 0
	err := as.db.QueryRow("SELECT COUNT(*) FROM patients").Scan(&stats.TotalPatients)
	if err != nil {
		log.Println("Note: could not fetch total patients, using fallback:", err)
		stats.TotalPatients = 0
	}

	// Total Doctors - fallback: 0
	err = as.db.QueryRow("SELECT COUNT(*) FROM doctors WHERE is_active = true").Scan(&stats.TotalDoctors)
	if err != nil {
		log.Println("Note: could not fetch total doctors, using fallback:", err)
		stats.TotalDoctors = 0
	}

	// Today's Appointments - fallback: 0
	err = as.db.QueryRow(`
		SELECT COUNT(*) FROM appointments
		WHERE DATE(appointment_date) = CURRENT_DATE
	`).Scan(&stats.TodayAppointments)
	if err != nil {
		log.Println("Note: could not fetch today appointments, using fallback:", err)
		stats.TodayAppointments = 0
	}

	// Checked In Patients - fallback: 0
	err = as.db.QueryRow(`
		SELECT COUNT(*) FROM appointments
		WHERE DATE(appointment_date) = CURRENT_DATE AND status = 'Checked In'
	`).Scan(&stats.CheckedInPatients)
	if err != nil {
		log.Println("Note: could not fetch checked in patients, using fallback:", err)
		stats.CheckedInPatients = 0
	}

	// Available Beds - fallback: 0
	err = as.db.QueryRow("SELECT COUNT(*) FROM beds WHERE status = 'available'").Scan(&stats.AvailableBeds)
	if err != nil {
		log.Println("Note: could not fetch available beds, using fallback:", err)
		stats.AvailableBeds = 0
	}

	// Occupied Beds - fallback: 0
	err = as.db.QueryRow("SELECT COUNT(*) FROM beds WHERE status = 'occupied'").Scan(&stats.OccupiedBeds)
	if err != nil {
		log.Println("Note: could not fetch occupied beds, using fallback:", err)
		stats.OccupiedBeds = 0
	}

	// Pending Appointments - fallback: 0
	err = as.db.QueryRow(`
		SELECT COUNT(*) FROM appointments
		WHERE status IN ('Scheduled', 'Checked In', 'Ready for Doctor')
	`).Scan(&stats.PendingAppointments)
	if err != nil {
		log.Println("Note: could not fetch pending appointments, using fallback:", err)
		stats.PendingAppointments = 0
	}

	// Completed Appointments - fallback: 0
	err = as.db.QueryRow(`
		SELECT COUNT(*) FROM appointments
		WHERE status = 'Completed'
	`).Scan(&stats.CompletedAppointments)
	if err != nil {
		log.Println("Note: could not fetch completed appointments, using fallback:", err)
		stats.CompletedAppointments = 0
	}

	// Average Patients Per Day (last 30 days) - fallback: 0
	err = as.db.QueryRow(`
		SELECT AVG(daily_count)::float FROM (
			SELECT COUNT(*) as daily_count
			FROM appointments
			WHERE DATE(appointment_date) BETWEEN CURRENT_DATE - INTERVAL '30 days' AND CURRENT_DATE
			GROUP BY DATE(appointment_date)
		) as daily_stats
	`).Scan(&stats.AveragePatientsPerDay)
	if err != nil {
		log.Println("Note: could not fetch average patients, using fallback:", err)
		stats.AveragePatientsPerDay = 0
	}

	// System Uptime (from settings or default)
	stats.SystemUptime = "Operational"

	return stats, nil
}

// SystemSetting represents a system configuration setting
type SystemSetting struct {
	ID    string `json:"id"`
	Key   string `json:"key"`
	Value string `json:"value"`
}

// GetSystemSettings retrieves all system settings
func (as *AdminService) GetSystemSettings() ([]SystemSetting, error) {
	query := "SELECT id, setting_key, setting_value FROM system_settings ORDER BY setting_key"
	log.Printf("DEBUG (Service): Executing query: %s\n", query)

	rows, err := as.db.Query(query)
	if err != nil {
		// Table doesn't exist or other error - return empty slice
		log.Printf("ERROR (Service): system_settings query error: %v\n", err)
		log.Println("ERROR (Service): system_settings table may not exist yet")
		return []SystemSetting{}, nil
	}
	defer rows.Close()

	var settings []SystemSetting
	rowCount := 0
	for rows.Next() {
		rowCount++
		var setting SystemSetting
		err := rows.Scan(&setting.ID, &setting.Key, &setting.Value)
		if err != nil {
			log.Printf("DEBUG (Service): Error scanning setting row %d: %v\n", rowCount, err)
			continue
		}
		log.Printf("DEBUG (Service): Scanned row %d - Key: %s, Value: %s, ID: %s\n", rowCount, setting.Key, setting.Value, setting.ID)
		settings = append(settings, setting)
	}

	log.Printf("DEBUG (Service): GetSystemSettings scanned %d total rows, returning %d settings\n", rowCount, len(settings))

	if err = rows.Err(); err != nil {
		log.Printf("ERROR (Service): Error during row iteration: %v\n", err)
	}

	return settings, nil
}

// UpdateSystemSetting updates a system configuration value
func (as *AdminService) UpdateSystemSetting(key string, value string) error {
	var exists int
	checkQuery := "SELECT 1 FROM system_settings WHERE setting_key = $1"
	err := as.db.QueryRow(checkQuery, key).Scan(&exists)

	if err == sql.ErrNoRows {
		// Insert new setting
		log.Printf("DEBUG (Service): %s does not exist, inserting new setting\n", key)
		insertQuery := `
			INSERT INTO system_settings (id, setting_key, setting_value, updated_at)
			VALUES (gen_random_uuid(), $1, $2, NOW())
		`
		result, err := as.db.Exec(insertQuery, key, value)
		if err != nil {
			log.Printf("ERROR: Failed to insert setting %s: %v\n", key, err)
			return fmt.Errorf("failed to create setting: %w", err)
		}
		log.Printf("DEBUG (Service): Successfully inserted setting %s\n", key)
		_ = result
	} else if err == nil {
		// Update existing setting
		log.Printf("DEBUG (Service): %s exists, updating value\n", key)
		updateQuery := `
			UPDATE system_settings
			SET setting_value = $1, updated_at = NOW()
			WHERE setting_key = $2
		`
		result, err := as.db.Exec(updateQuery, value, key)
		if err != nil {
			log.Printf("ERROR: Failed to update setting %s: %v\n", key, err)
			return fmt.Errorf("failed to update setting: %w", err)
		}
		log.Printf("DEBUG (Service): Successfully updated setting %s\n", key)
		_ = result
	} else {
		log.Printf("ERROR: Failed to check setting %s: %v\n", key, err)
		return fmt.Errorf("failed to check setting: %w", err)
	}

	return nil
}

// Report represents a generated report
type Report struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	ReportType  string    `json:"report_type"`
	GeneratedAt time.Time `json:"generated_at"`
	GeneratedBy string    `json:"generated_by"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	RecordCount int       `json:"record_count"`
}

// GetReports retrieves generated reports
func (as *AdminService) GetReports(reportType string, limit int, offset int) ([]Report, error) {
	query := `
		SELECT id, title, report_type, generated_at, generated_by, start_date, end_date, record_count
		FROM reports
		WHERE 1=1
	`
	args := []interface{}{}

	if reportType != "" {
		query += " AND report_type = $1"
		args = append(args, reportType)
	}

	query += " ORDER BY generated_at DESC LIMIT $" + fmt.Sprintf("%d", len(args)+1) + " OFFSET $" + fmt.Sprintf("%d", len(args)+2)
	args = append(args, limit, offset)

	rows, err := as.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch reports: %w", err)
	}
	defer rows.Close()

	var reports []Report
	for rows.Next() {
		var r Report
		err := rows.Scan(&r.ID, &r.Title, &r.ReportType, &r.GeneratedAt, &r.GeneratedBy, &r.StartDate, &r.EndDate, &r.RecordCount)
		if err != nil {
			continue
		}
		reports = append(reports, r)
	}

	return reports, nil
}

// GetAdminProfile retrieves the admin profile from database
func (as *AdminService) GetAdminProfile(adminID string) (*AdminProfile, error) {
	query := `
		SELECT id, admin_id, name, email, phone, address, dob, bio, role, department, status, join_date
		FROM admin_profiles
		WHERE admin_id = $1
	`

	profile := &AdminProfile{}
	var joinDate time.Time

	err := as.db.QueryRow(query, adminID).Scan(
		&profile.ID,
		&profile.AdminID,
		&profile.Name,
		&profile.Email,
		&profile.Phone,
		&profile.Address,
		&profile.DOB,
		&profile.Bio,
		&profile.Role,
		&profile.Department,
		&profile.Status,
		&joinDate,
	)

	if err == sql.ErrNoRows {
		log.Printf("DEBUG: No admin profile found for %s, returning nil\n", adminID)
		return nil, nil
	}

	if err != nil {
		log.Printf("ERROR: Failed to get admin profile: %v\n", err)
		return nil, fmt.Errorf("failed to get admin profile: %w", err)
	}

	profile.JoinDate = joinDate.Format("2006-01-02")
	log.Printf("DEBUG: Retrieved admin profile for %s: %+v\n", adminID, profile)
	return profile, nil
}

// UpdateAdminProfile updates the admin profile in database
func (as *AdminService) UpdateAdminProfile(adminID string, profile *AdminProfile) (*AdminProfile, error) {
	query := `
		UPDATE admin_profiles
		SET name = $1, email = $2, phone = $3, address = $4, dob = $5, bio = $6, updated_at = NOW()
		WHERE admin_id = $7
		RETURNING id, admin_id, name, email, phone, address, dob, bio, role, department, status, join_date
	`

	var joinDate time.Time
	err := as.db.QueryRow(
		query,
		profile.Name,
		profile.Email,
		profile.Phone,
		profile.Address,
		profile.DOB,
		profile.Bio,
		adminID,
	).Scan(
		&profile.ID,
		&profile.AdminID,
		&profile.Name,
		&profile.Email,
		&profile.Phone,
		&profile.Address,
		&profile.DOB,
		&profile.Bio,
		&profile.Role,
		&profile.Department,
		&profile.Status,
		&joinDate,
	)

	if err != nil {
		log.Printf("ERROR: Failed to update admin profile: %v\n", err)
		return nil, fmt.Errorf("failed to update admin profile: %w", err)
	}

	profile.JoinDate = joinDate.Format("2006-01-02")
	log.Printf("DEBUG: Successfully updated admin profile for %s\n", adminID)
	return profile, nil
}

// CreateStaffInput holds data for creating staff members
type CreateStaffInput struct {
	Name           string // Full name (single field)
	Email          string
	Phone          string
	Password       string // Optional - will be auto-generated if empty
	Role           string // Doctor, Nurse, Receptionist, Pharmacist, Admin
	Department     string // For Doctor/Nurse roles
	Specialization string // For Doctor role
	LicenseNumber  string // For Doctor role
	Address        string
}

// CreateStaff creates a new staff member with role assignment
func (as *AdminService) CreateStaff(input CreateStaffInput) (staffID string, userID string, password string, err error) {
	tx, err := as.db.Begin()
	if err != nil {
		return "", "", "", fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Check if email already exists
	var existingID string
	checkQuery := "SELECT id FROM users WHERE email = $1 LIMIT 1"
	err = tx.QueryRow(checkQuery, input.Email).Scan(&existingID)
	if err == nil {
		return "", "", "", errors.New("email already exists")
	}
	if err != sql.ErrNoRows {
		return "", "", "", fmt.Errorf("email check failed: %w", err)
	}

	// Generate password if not provided
	password = input.Password
	if password == "" {
		password = generateDefaultPassword(input.Name)
	}

	// Hash password using bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user with correct schema (single name field, phone, address)
	// Insert user with phone and address
	userQuery := `
		INSERT INTO users (name, email, password_hash, phone, address, is_active)
		VALUES ($1, $2, $3, $4, $5, true)
		RETURNING id
	`
	userID = ""
	err = tx.QueryRow(userQuery, input.Name, input.Email, string(hashedPassword), input.Phone, input.Address).Scan(&userID)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to create user: %w", err)
	}

	// Assign role to user
	var roleID string
	roleQuery := "SELECT id FROM roles WHERE name = $1"
	err = tx.QueryRow(roleQuery, input.Role).Scan(&roleID)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to get role ID: %w", err)
	}

	userRoleQuery := `
		INSERT INTO user_roles (user_id, role_id)
		VALUES ($1, $2)
	`
	_, err = tx.Exec(userRoleQuery, userID, roleID)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to assign role: %w", err)
	}

	// For Doctor role, create doctor record
	if input.Role == "Doctor" {
		doctorQuery := `
			INSERT INTO doctors (user_id, department, specialization, license_number, is_active, created_at)
			VALUES ($1, $2, $3, $4, true, NOW())
			RETURNING id
		`
		err = tx.QueryRow(doctorQuery, userID, input.Department, input.Specialization, input.LicenseNumber).Scan(&staffID)
		if err != nil {
			return "", "", "", fmt.Errorf("failed to create doctor record: %w", err)
		}
	} else if input.Role == "Nurse" {
		// Create nurse record if needed (optional - can use just user table)
		staffID = userID
	} else if input.Role == "Receptionist" {
		// No additional table needed for receptionist
		staffID = userID
	} else if input.Role == "Pharmacist" {
		// No additional table needed for pharmacist
		staffID = userID
	} else if input.Role == "Admin" {
		// No additional table needed for admin
		staffID = userID
	} else {
		staffID = userID
	}

	err = tx.Commit()
	if err != nil {
		return "", "", "", fmt.Errorf("failed to commit transaction: %w", err)
	}

	return staffID, userID, password, nil
}

// StaffMember represents a staff member with full details
type StaffMember struct {
	ID             string    `json:"id"`
	UserID         string    `json:"user_id"`
	Name           string    `json:"name"`
	Email          string    `json:"email"`
	Phone          string    `json:"phone"`
	Role           string    `json:"role"`
	Department     string    `json:"department"`
	Specialization string    `json:"specialization"`
	IsActive       bool      `json:"is_active"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// GetStaffByRole returns staff members filtered by role
func (as *AdminService) GetStaffByRole(role string, limit int, offset int) ([]StaffMember, error) {
	query := `
		SELECT 
			u.id::text as id,
			u.id as user_id,
			u.name,
			u.email,
			u.phone,
			ro.name as role,
			COALESCE(dept.name, '') as department,
			COALESCE(s.name, '') as specialization,
			u.is_active,
			u.created_at,
			u.updated_at
		FROM users u
		JOIN user_roles ur ON u.id = ur.user_id
		JOIN roles ro ON ur.role_id = ro.id
		LEFT JOIN doctors d ON u.id = d.user_id
		LEFT JOIN departments dept ON d.department_id = dept.id
		LEFT JOIN specializations s ON d.specialization_id = s.id
		WHERE ro.name = $1
		ORDER BY u.name ASC
		LIMIT $2 OFFSET $3
	`

	rows, err := as.db.Query(query, role, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch staff: %w", err)
	}
	defer rows.Close()

	var staff []StaffMember
	for rows.Next() {
		var member StaffMember
		err := rows.Scan(
			&member.ID, &member.UserID, &member.Name,
			&member.Email, &member.Phone, &member.Role, &member.Department,
			&member.Specialization, &member.IsActive, &member.CreatedAt, &member.UpdatedAt,
		)
		if err != nil {
			continue
		}
		staff = append(staff, member)
	}

	return staff, nil
}

// GetAllStaff returns all staff members
func (as *AdminService) GetAllStaff(limit int, offset int) ([]StaffMember, error) {
	query := `
		SELECT 
			u.id::text as id,
			u.id as user_id,
			u.name,
			u.email,
			u.phone,
			ro.name as role,
			COALESCE(dept.name, '') as department,
			COALESCE(s.name, '') as specialization,
			u.is_active,
			u.created_at,
			u.updated_at
		FROM users u
		JOIN user_roles ur ON u.id = ur.user_id
		JOIN roles ro ON ur.role_id = ro.id
		LEFT JOIN doctors d ON u.id = d.user_id
		LEFT JOIN departments dept ON d.department_id = dept.id
		LEFT JOIN specializations s ON d.specialization_id = s.id
		WHERE ro.name IN ('Doctor', 'Nurse', 'Receptionist', 'Pharmacist', 'Admin')
		ORDER BY u.name ASC
		LIMIT $1 OFFSET $2
	`

	rows, err := as.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch staff: %w", err)
	}
	defer rows.Close()

	var staff []StaffMember
	for rows.Next() {
		var member StaffMember
		err := rows.Scan(
			&member.ID, &member.UserID, &member.Name,
			&member.Email, &member.Phone, &member.Role, &member.Department,
			&member.Specialization, &member.IsActive, &member.CreatedAt, &member.UpdatedAt,
		)
		if err != nil {
			continue
		}
		staff = append(staff, member)
	}

	return staff, nil
}

// UpdateStaffInput holds data for updating staff
type UpdateStaffInput struct {
	FirstName      string
	LastName       string
	Phone          string
	Department     string
	Specialization *string
}

// UpdateStaff updates staff member information
func (as *AdminService) UpdateStaff(userID string, input UpdateStaffInput) error {
	tx, err := as.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Update user info
	updateUserQuery := `
		UPDATE users
		SET first_name = $1, last_name = $2, phone = $3, updated_at = NOW()
		WHERE id = $4
	`
	_, err = tx.Exec(updateUserQuery, input.FirstName, input.LastName, input.Phone, userID)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	// Check if doctor and update specialization if provided
	if input.Specialization != nil && *input.Specialization != "" {
		var specID int
		getSpecQuery := "SELECT id FROM specializations WHERE name = $1"
		err = tx.QueryRow(getSpecQuery, *input.Specialization).Scan(&specID)
		if err == nil {
			updateDoctorQuery := `
				UPDATE doctors
				SET specialization_id = $1, updated_at = NOW()
				WHERE user_id = $2
			`
			tx.Exec(updateDoctorQuery, specID, userID)
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeactivateStaff deactivates a staff member
func (as *AdminService) DeactivateStaff(userID string) error {
	tx, err := as.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Deactivate doctor
	doctorQuery := "UPDATE doctors SET is_active = false, updated_at = NOW() WHERE user_id = $1"
	tx.Exec(doctorQuery, userID)

	// Deactivate nurse
	nurseQuery := "UPDATE nurses SET is_active = false, updated_at = NOW() WHERE user_id = $1"
	tx.Exec(nurseQuery, userID)

	// Deactivate receptionist
	receptionistQuery := "UPDATE receptionists SET is_active = false, updated_at = NOW() WHERE user_id = $1"
	tx.Exec(receptionistQuery, userID)

	// Update user status
	updateUserQuery := "UPDATE users SET updated_at = NOW() WHERE id = $1"
	_, err = tx.Exec(updateUserQuery, userID)
	if err != nil {
		return fmt.Errorf("failed to deactivate staff: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// ActivateStaff activates a deactivated staff member
func (as *AdminService) ActivateStaff(userID string) error {
	tx, err := as.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Activate doctor
	doctorQuery := "UPDATE doctors SET is_active = true, updated_at = NOW() WHERE user_id = $1"
	tx.Exec(doctorQuery, userID)

	// Activate nurse
	nurseQuery := "UPDATE nurses SET is_active = true, updated_at = NOW() WHERE user_id = $1"
	tx.Exec(nurseQuery, userID)

	// Activate receptionist
	receptionistQuery := "UPDATE receptionists SET is_active = true, updated_at = NOW() WHERE user_id = $1"
	tx.Exec(receptionistQuery, userID)

	// Update user status
	updateUserQuery := "UPDATE users SET updated_at = NOW() WHERE id = $1"
	_, err = tx.Exec(updateUserQuery, userID)
	if err != nil {
		return fmt.Errorf("failed to activate staff: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DoctorDepartment represents a doctor in a specific department
type DoctorDepartment struct {
	ID                    string `json:"id"`
	UserID                string `json:"user_id"`
	Name                  string `json:"name"`
	Email                 string `json:"email"`
	Specialization        string `json:"specialization"`
	Phone                 string `json:"phone"`
	IsAvailable           bool   `json:"is_available"`
	DepartmentDescription string `json:"department_description"`
}

// GetDoctorsInDepartment returns doctors for a specific department
func (as *AdminService) GetDoctorsInDepartment(departmentID string) ([]DoctorDepartment, error) {
	// First, try to get the department name from the ID (if input is a UUID)
	departmentName := departmentID

	// Try to look up department name by ID
	var deptName string
	err := as.db.QueryRow(
		"SELECT name FROM departments WHERE id::text = $1 LIMIT 1",
		departmentID,
	).Scan(&deptName)

	if err == nil {
		// Found a department by ID, use its name
		departmentName = deptName
	}
	// If not found, assume departmentID is already the department name

	query := `
		SELECT 
			d.id::text,
			u.id,
			u.name,
			u.email,
			d.specialization,
			u.phone,
			d.is_active,
			COALESCE(dept.description, '') as department_description
		FROM doctors d
		JOIN users u ON d.user_id = u.id
		LEFT JOIN departments dept ON d.department_id = dept.id OR d.department = dept.name
		WHERE (d.department_id::text = $1 OR d.department = $2) AND d.is_active = true
		ORDER BY u.name ASC
	`

	rows, err := as.db.Query(query, departmentID, departmentName)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch doctors: %w", err)
	}
	defer rows.Close()

	var doctors []DoctorDepartment
	for rows.Next() {
		var doc DoctorDepartment
		err := rows.Scan(
			&doc.ID, &doc.UserID, &doc.Name,
			&doc.Email, &doc.Specialization, &doc.Phone, &doc.IsAvailable,
			&doc.DepartmentDescription,
		)
		if err != nil {
			continue
		}
		doctors = append(doctors, doc)
	}

	return doctors, nil
}
