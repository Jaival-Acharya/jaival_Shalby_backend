package services

import (
	"database/sql"

	"shalby_backend/models"
)

// GetDashboardStats returns comprehensive admin dashboard statistics
func GetDashboardStats(db *sql.DB) (*models.AdminDashboardStats, error) {
	stats := &models.AdminDashboardStats{}

	// Total appointments
	err := db.QueryRow("SELECT COUNT(*) FROM appointments").Scan(&stats.TotalAppointments)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	// Completed appointments
	err = db.QueryRow("SELECT COUNT(*) FROM appointments WHERE status = 'Completed'").Scan(&stats.CompletedAppointments)
	if err != nil && err != sql.ErrNoRows {
		stats.CompletedAppointments = 0
	}

	// Scheduled appointments
	err = db.QueryRow("SELECT COUNT(*) FROM appointments WHERE status = 'Scheduled'").Scan(&stats.ScheduledAppointments)
	if err != nil && err != sql.ErrNoRows {
		stats.ScheduledAppointments = 0
	}

	// Active patients
	err = db.QueryRow("SELECT COUNT(*) FROM patients").Scan(&stats.ActivePatients)
	if err != nil && err != sql.ErrNoRows {
		stats.ActivePatients = 0
	}

	// Total doctors
	err = db.QueryRow("SELECT COUNT(*) FROM doctors WHERE is_active = true").Scan(&stats.TotalDoctors)
	if err != nil && err != sql.ErrNoRows {
		stats.TotalDoctors = 0
	}

	// Total medicines
	err = db.QueryRow("SELECT COUNT(*) FROM medicines WHERE is_active = true").Scan(&stats.TotalMedicines)
	if err != nil && err != sql.ErrNoRows {
		stats.TotalMedicines = 0
	}

	// Low stock alerts
	err = db.QueryRow("SELECT COUNT(*) FROM medicine_inventory WHERE stock_quantity <= reorder_level").Scan(&stats.LowStockAlerts)
	if err != nil && err != sql.ErrNoRows {
		stats.LowStockAlerts = 0
	}

	return stats, nil
}

// GetSystemSettings returns current system settings
func GetSystemSettings(db *sql.DB) (map[string]string, error) {
	// Return default settings
	// TODO: When system_settings table is created and populated, fetch from there
	settings := map[string]string{
		"hospital_name":          "Shalby Healthcare",
		"hospital_email":         "info@shalby.com",
		"hospital_phone":         "+1 (555) 123-4567",
		"hospital_address":       "Healthcare City, Medical District, NY 10001",
		"hospital_website":       "www.shalby.com",
		"currency":               "USD",
		"timezone":               "America/New_York",
		"appointment_duration":   "30",
		"appointment_buffer":     "10",
		"max_daily_appointments": "50",
	}

	// Try to fetch from database, but don't fail if table doesn't exist
	rows, err := db.Query(`SELECT setting_key, setting_value FROM system_settings`)
	if err != nil {
		// Table might not exist - return defaults
		return settings, nil
	}
	defer rows.Close()

	dbSettings := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			continue
		}
		dbSettings[key] = value
	}

	// Merge database settings with defaults (DB overrides defaults)
	for key, value := range dbSettings {
		settings[key] = value
	}

	return settings, nil
}

// UpdateSystemSetting updates or creates a system setting
func UpdateSystemSetting(db *sql.DB, key string, value string) error {
	query := `
	INSERT INTO system_settings (setting_key, setting_value, updated_at)
	VALUES ($1, $2, NOW())
	ON CONFLICT (setting_key) DO UPDATE SET setting_value = $2, updated_at = NOW()
	`

	_, err := db.Exec(query, key, value)
	if err != nil {
		return err
	}

	return nil
}

// GetReports returns reports data
func GetReports(db *sql.DB, startDate, endDate string) (map[string]interface{}, error) {
	reports := map[string]interface{}{
		"startDate": startDate,
		"endDate":   endDate,
	}

	// Get total appointments
	var totalAppointments int
	err := db.QueryRow(`
		SELECT COUNT(*) FROM appointments 
		WHERE appointment_date >= $1 AND appointment_date <= $2
	`, startDate, endDate).Scan(&totalAppointments)

	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	reports["totalAppointments"] = totalAppointments

	// Get appointments by status
	rows, err := db.Query(`
		SELECT status, COUNT(*) as count FROM appointments 
		WHERE appointment_date >= $1 AND appointment_date <= $2
		GROUP BY status
	`, startDate, endDate)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	statusCounts := make(map[string]int)
	for rows.Next() {
		var status string
		var count int
		err := rows.Scan(&status, &count)
		if err != nil {
			return nil, err
		}
		statusCounts[status] = count
	}

	reports["appointmentsByStatus"] = statusCounts

	// Get top medicines
	var topMedicines []map[string]interface{}
	rows, err = db.Query(`
		SELECT m.name, COUNT(pi.medicine_id) as usage_count
		FROM medicines m
		LEFT JOIN prescription_items pi ON m.id = pi.medicine_id
		GROUP BY m.id, m.name
		ORDER BY usage_count DESC
		LIMIT 5
	`)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		var count int
		err := rows.Scan(&name, &count)
		if err != nil {
			return nil, err
		}
		topMedicines = append(topMedicines, map[string]interface{}{
			"name":  name,
			"count": count,
		})
	}

	reports["topMedicines"] = topMedicines

	return reports, nil
}
