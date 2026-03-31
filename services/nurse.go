package services

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// NurseService handles business logic for nurse operations
type NurseService struct {
	db *sql.DB
}

// NewNurseService creates a new nurse service instance
func NewNurseService(db *sql.DB) *NurseService {
	return &NurseService{db: db}
}

// RecordVitalsInput holds vital signs data to be recorded
type RecordVitalsInput struct {
	AppointmentID             string
	PatientID                 string
	TemperatureCelsius        float64
	BloodPressureSystolic     int
	BloodPressureDiastolic    int
	PulseHeartRateBpm         int
	RespiratoryRateBreathsMin int
	OxygenSaturationPercent   float64
	HeightCm                  float64
	WeightKg                  float64
	Notes                     string
}

// RecordVitals saves vital signs for a patient and calculates BMI
func (ns *NurseService) RecordVitals(input RecordVitalsInput, nurseID string) (vitalID string, err error) {
	err = validateVitals(input)
	if err != nil {
		return "", err
	}

	tx, err := ns.db.Begin()
	if err != nil {
		return "", fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	var appointmentStatus string
	var patientIDFromAppt string
	apptQuery := "SELECT status, patient_id FROM appointments WHERE id = $1"
	err = tx.QueryRow(apptQuery, input.AppointmentID).Scan(&appointmentStatus, &patientIDFromAppt)
	if err == sql.ErrNoRows {
		return "", errors.New("appointment not found")
	}
	if err != nil {
		return "", fmt.Errorf("failed to fetch appointment: %w", err)
	}

	if appointmentStatus != "Checked In" {
		return "", fmt.Errorf("appointment is %s, vitals can only be recorded for 'Checked In' patients", appointmentStatus)
	}

	if patientIDFromAppt != input.PatientID {
		return "", errors.New("patient ID does not match appointment")
	}

	var bmi sql.NullFloat64
	if input.HeightCm > 0 && input.WeightKg > 0 {
		bmiValue := input.WeightKg / ((input.HeightCm / 100) * (input.HeightCm / 100))
		bmi = sql.NullFloat64{Float64: bmiValue, Valid: true}
	}

	vitalQuery := `
		INSERT INTO patient_vitals (
			patient_id, 
			recorded_at, 
			recorded_by, 
			temperature_celsius, 
			blood_pressure_systolic, 
			blood_pressure_diastolic, 
			pulse_bpm,
			height_cm,
			weight_kg,
			bmi
		)
		VALUES ($1, NOW(), $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`

	err = tx.QueryRow(
		vitalQuery,
		input.PatientID,
		nurseID,
		input.TemperatureCelsius,
		input.BloodPressureSystolic,
		input.BloodPressureDiastolic,
		input.PulseHeartRateBpm,
		input.HeightCm,
		input.WeightKg,
		bmi,
	).Scan(&vitalID)
	if err != nil {
		return "", fmt.Errorf("failed to record vitals: %w", err)
	}

	updateApptQuery := `
		UPDATE appointments
		SET status = 'Ready for Doctor'
		WHERE id = $1
	`
	_, err = tx.Exec(updateApptQuery, input.AppointmentID)
	if err != nil {
		return "", fmt.Errorf("failed to update appointment status: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return "", fmt.Errorf("failed to commit transaction: %w", err)
	}

	return vitalID, nil
}

// GetCheckedInPatients returns all patients with "Checked In" appointments
func (ns *NurseService) GetCheckedInPatients(limit int, offset int) ([]map[string]interface{}, error) {
	query := `
		SELECT 
			a.id as appointment_id,
			a.patient_id,
			u.first_name,
			u.last_name,
			u.phone,
			p.date_of_birth,
			p.gender,
			a.appointment_date,
			a.appointment_time,
			a.chief_complaint,
			a.checked_in_at,
			b.room_number,
			b.bed_number
		FROM appointments a
		JOIN patients p ON a.patient_id = p.id
		JOIN users u ON p.user_id = u.id
		LEFT JOIN beds b ON a.bed_id = b.id
		WHERE a.status = 'Checked In'
		ORDER BY a.checked_in_at ASC
		LIMIT $1 OFFSET $2
	`

	rows, err := ns.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch checked-in patients: %w", err)
	}
	defer rows.Close()

	var patients []map[string]interface{}
	for rows.Next() {
		var appointmentID, patientID, firstName, lastName, phone string
		var dateOfBirth time.Time
		var gender, complaint string
		var appointmentDate time.Time
		var appointmentTime string
		var checkedInAt time.Time
		var roomNumber, bedNumber sql.NullString

		err := rows.Scan(
			&appointmentID, &patientID, &firstName, &lastName, &phone,
			&dateOfBirth, &gender, &appointmentDate, &appointmentTime, &complaint,
			&checkedInAt, &roomNumber, &bedNumber,
		)
		if err != nil {
			continue
		}

		bedInfo := "TBD"
		if roomNumber.Valid && bedNumber.Valid {
			bedInfo = fmt.Sprintf("%s-%s", roomNumber.String, bedNumber.String)
		}

		patients = append(patients, map[string]interface{}{
			"appointment_id":   appointmentID,
			"patient_id":       patientID,
			"patient_name":     firstName + " " + lastName,
			"patient_phone":    phone,
			"date_of_birth":    dateOfBirth,
			"gender":           gender,
			"appointment_date": appointmentDate,
			"appointment_time": appointmentTime,
			"complaint":        complaint,
			"checked_in_at":    checkedInAt,
			"bed_info":         bedInfo,
		})
	}

	return patients, nil
}

// GetPatientVitals returns the latest vital signs for a patient
func (ns *NurseService) GetPatientVitals(patientID string, limit int) ([]map[string]interface{}, error) {
	query := `
		SELECT 
			id,
			recorded_at,
			temperature_celsius,
			blood_pressure_systolic,
			blood_pressure_diastolic,
			pulse_bpm,
			height_cm,
			weight_kg,
			bmi
		FROM patient_vitals
		WHERE patient_id = $1
		ORDER BY recorded_at DESC
		LIMIT $2
	`

	rows, err := ns.db.Query(query, patientID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch vitals: %w", err)
	}
	defer rows.Close()

	var vitals []map[string]interface{}
	for rows.Next() {
		var id string
		var recordedAt time.Time
		var temp float64
		var sysBP, diastBP, pulse int
		var height, weight sql.NullFloat64
		var bmi sql.NullFloat64

		err := rows.Scan(&id, &recordedAt, &temp, &sysBP, &diastBP, &pulse, &height, &weight, &bmi)
		if err != nil {
			continue
		}

		vitals = append(vitals, map[string]interface{}{
			"id":                  id,
			"recorded_at":         recordedAt,
			"temperature_celsius": temp,
			"blood_pressure":      fmt.Sprintf("%d/%d", sysBP, diastBP),
			"pulse_bpm":           pulse,
			"height_cm":           height.Float64,
			"weight_kg":           weight.Float64,
			"bmi":                 bmi.Float64,
		})
	}

	return vitals, nil
}

// GetRecentVitals returns the most recent vitals recorded across all patients
func (ns *NurseService) GetRecentVitals(limit int) ([]map[string]interface{}, error) {
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	query := `
		SELECT 
			DISTINCT ON (p.id) 
			p.id as patient_id,
			u.name as patient_name,
			pv.id as vital_id,
			pv.recorded_at,
			pv.temperature_celsius,
			pv.blood_pressure_systolic,
			pv.blood_pressure_diastolic,
			pv.pulse_bpm,
			pv.height_cm,
			pv.weight_kg,
			pv.bmi
		FROM patient_vitals pv
		JOIN patients p ON pv.patient_id = p.id
		JOIN users u ON p.user_id = u.id
		ORDER BY p.id, pv.recorded_at DESC
		LIMIT $1
	`

	rows, err := ns.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch recent vitals: %w", err)
	}
	defer rows.Close()

	var vitals []map[string]interface{}
	for rows.Next() {
		var patientID, patientName, vitalID string
		var recordedAt time.Time
		var temp float64
		var sysBP, diastBP, pulse int
		var height, weight, bmi sql.NullFloat64

		err := rows.Scan(
			&patientID, &patientName, &vitalID, &recordedAt, &temp,
			&sysBP, &diastBP, &pulse, &height, &weight, &bmi,
		)
		if err != nil {
			continue
		}

		vitals = append(vitals, map[string]interface{}{
			"vitalId":     vitalID,
			"patientId":   patientID,
			"patientName": patientName,
			"recordedAt":  recordedAt,
			"temperature": temp,
			"bpSystolic":  sysBP,
			"bpDiastolic": diastBP,
			"pulse":       pulse,
			"heightCm":    height.Float64,
			"weightKg":    weight.Float64,
			"bmi":         bmi.Float64,
		})
	}

	return vitals, nil
}

func validateVitals(input RecordVitalsInput) error {
	if input.TemperatureCelsius < 35 || input.TemperatureCelsius > 42 {
		return fmt.Errorf("temperature %.1f°C is outside acceptable range (35-42°C)", input.TemperatureCelsius)
	}

	if input.BloodPressureSystolic < 80 || input.BloodPressureSystolic > 200 {
		return fmt.Errorf("systolic BP %d mmHg is outside acceptable range (80-200)", input.BloodPressureSystolic)
	}

	if input.BloodPressureDiastolic < 40 || input.BloodPressureDiastolic > 130 {
		return fmt.Errorf("diastolic BP %d mmHg is outside acceptable range (40-130)", input.BloodPressureDiastolic)
	}

	if input.PulseHeartRateBpm < 30 || input.PulseHeartRateBpm > 200 {
		return fmt.Errorf("heart rate %d bpm is outside acceptable range (30-200)", input.PulseHeartRateBpm)
	}

	if input.RespiratoryRateBreathsMin > 0 && (input.RespiratoryRateBreathsMin < 8 || input.RespiratoryRateBreathsMin > 60) {
		return fmt.Errorf("respiratory rate %d breaths/min is outside acceptable range (8-60)", input.RespiratoryRateBreathsMin)
	}

	if input.OxygenSaturationPercent > 0 && (input.OxygenSaturationPercent < 70 || input.OxygenSaturationPercent > 100) {
		return fmt.Errorf("oxygen saturation %.1f%% is outside acceptable range (70-100%%)", input.OxygenSaturationPercent)
	}

	return nil
}

// GetQueue returns today's appointments with vitals status for the nurse queue
func (ns *NurseService) GetQueue(nurseID string, date string) ([]interface{}, error) {
	query := `
		SELECT 
			a.id as appointment_id,
			a.patient_id,
			u.name as patient_name,
			EXTRACT(YEAR FROM AGE(NOW(), p.date_of_birth))::integer as patient_age,
			p.gender,
			p.blood_group,
			a.doctor_id,
			d_user.name as doctor_name,
			a.appointment_date,
			a.appointment_time,
			a.chief_complaint,
			a.status,
			a.priority,
			COALESCE(pv.id IS NOT NULL, false) as has_vitals_today,
			COALESCE(pv.blood_pressure_systolic, 0) as bp_systolic,
			COALESCE(pv.blood_pressure_diastolic, 0) as bp_diastolic,
			COALESCE(pv.pulse_bpm, 0) as pulse,
			COALESCE(pv.temperature_celsius, 0) as temperature,
			COALESCE(pv.blood_sugar_mg_dl, 0) as blood_sugar,
			COALESCE(pv.recorded_at, NOW()) as vitals_recorded_at
		FROM appointments a
		JOIN patients p ON a.patient_id = p.id
		JOIN users u ON p.user_id = u.id
		JOIN doctors d ON a.doctor_id = d.id
		JOIN users d_user ON d.user_id = d_user.id
		LEFT JOIN patient_vitals pv ON p.id = pv.patient_id AND DATE(pv.recorded_at) = $1
		WHERE DATE(a.appointment_date) = $1 
			AND a.status IN ('Checked In', 'Ready for Doctor', 'In Consultation')
		ORDER BY a.appointment_time ASC
	`

	rows, err := ns.db.Query(query, date)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch queue: %w", err)
	}
	defer rows.Close()

	var queue []interface{}
	for rows.Next() {
		var appointmentID, patientID, patientName, gender, bloodGroup string
		var patientAge int
		var doctorID, doctorName, status, priority, chiefComplaint string
		var appointmentDate time.Time
		var appointmentTime string
		var hasVitals bool
		var bpSys, bpDia, pulse int
		var temp, bloodSugar float64
		var vitalsTime time.Time

		err := rows.Scan(
			&appointmentID, &patientID, &patientName, &patientAge, &gender, &bloodGroup,
			&doctorID, &doctorName, &appointmentDate, &appointmentTime, &chiefComplaint,
			&status, &priority, &hasVitals, &bpSys, &bpDia, &pulse, &temp, &bloodSugar, &vitalsTime,
		)
		if err != nil {
			continue
		}

		item := map[string]interface{}{
			"id":              patientID, // Add id field for frontend compatibility
			"appointmentId":   appointmentID,
			"patientId":       patientID,
			"patientName":     patientName,
			"patientAge":      patientAge,
			"gender":          gender,
			"bloodGroup":      bloodGroup,
			"doctorId":        doctorID,
			"doctorName":      doctorName,
			"appointmentDate": appointmentDate,
			"appointmentTime": appointmentTime,
			"complaint":       chiefComplaint, // Changed from chiefComplaint to complaint for frontend
			"status":          status,
			"priority":        priority,
			"hasVitalsToday":  hasVitals,
			"vitalsStatus":    map[bool]string{true: "recorded", false: "pending"}[hasVitals],
		}

		if hasVitals {
			item["latestVitals"] = map[string]interface{}{
				"bloodPressureSystolic":  bpSys,
				"bloodPressureDiastolic": bpDia,
				"pulseBpm":               pulse,
				"temperatureCelsius":     temp,
				"bloodSugarMgDl":         bloodSugar,
				"recordedAt":             vitalsTime,
			}
		}

		queue = append(queue, item)
	}

	return queue, nil
}

// GetTasksForNurse returns tasks assigned to a specific nurse
func (ns *NurseService) GetTasksForNurse(nurseID string) ([]interface{}, error) {
	// Query from audit_logs where action is CREATE_NURSE_TASK and nurse_id in JSON details matches
	// Use JSON operator instead of LIKE for more reliable matching
	query := `
		SELECT 
			a.resource_id as id,
			a.details,
			a.created_at
		FROM audit_logs a
		WHERE a.action = 'CREATE_NURSE_TASK'
		AND a.created_at > NOW() - INTERVAL '30 days'
		AND (a.details::jsonb ->> 'nurse_id') = $1
		ORDER BY a.created_at DESC
	`

	rows, err := ns.db.Query(query, nurseID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tasks: %w", err)
	}
	defer rows.Close()

	var tasks []interface{}
	for rows.Next() {
		var id, details string
		var createdAt time.Time

		err := rows.Scan(&id, &details, &createdAt)
		if err != nil {
			continue
		}

		// Parse JSON details
		var detailsMap map[string]interface{}
		err = json.Unmarshal([]byte(details), &detailsMap)
		if err != nil {
			continue
		}

		task := map[string]interface{}{
			"id":             id,
			"title":          detailsMap["task_type"],
			"description":    detailsMap["description"],
			"task_type":      detailsMap["task_type"],
			"priority":       detailsMap["priority"],
			"patient_id":     detailsMap["patient_id"],
			"patient_name":   detailsMap["patient_id"], // Will show ID unless we join patients
			"due_by":         detailsMap["due_by"],
			"appointment_id": detailsMap["appointment_id"],
			"notes":          "",
			"status":         "Pending",
			"created_at":     createdAt,
			"created_by":     "", // From audit_logs user_id
		}

		tasks = append(tasks, task)
	}

	return tasks, nil
}

// UpdateTaskStatus updates a task's status and/or notes
func (ns *NurseService) UpdateTaskStatus(taskID, nurseID, status, note string) error {
	if taskID == "" || nurseID == "" || status == "" {
		return errors.New("task ID, nurse ID, and status are required")
	}

	var assignedTo string
	var currentNotes string
	err := ns.db.QueryRow("SELECT assigned_to, COALESCE(notes, '') FROM nurse_tasks WHERE id = $1", taskID).Scan(&assignedTo, &currentNotes)
	if err == sql.ErrNoRows {
		return errors.New("task not found")
	}
	if err != nil {
		return fmt.Errorf("failed to fetch task: %w", err)
	}

	if assignedTo != nurseID {
		return errors.New("you are not assigned to this task")
	}

	var newNotes string
	if note != "" {
		newNotes = currentNotes
		if currentNotes != "" {
			newNotes += "\n"
		}
		newNotes += fmt.Sprintf("[%s] %s", time.Now().Format("2006-01-02 15:04:05"), note)
	}

	query := `
		UPDATE nurse_tasks
		SET status = $1,
			notes = CASE WHEN $2 != '' THEN $2 ELSE notes END,
			updated_at = NOW(),
			completed_at = CASE WHEN $1 = 'Completed' THEN NOW() ELSE completed_at END
		WHERE id = $3
	`

	_, err = ns.db.Exec(query, status, newNotes, taskID)
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	return nil
}

// GetAdmittedPatients returns all patients currently in beds
func (ns *NurseService) GetAdmittedPatients() ([]interface{}, error) {
	query := `
		SELECT 
			p.id as patient_id,
			u.name as patient_name,
			b.id as bed_id,
			b.room_number,
			b.bed_number,
			d.name as department_name,
			a.doctor_id,
			d_user.name as doctor_name,
			COALESCE(c.diagnosis, '') as diagnosis,
			a.checked_in_at as admitted_since,
			COALESCE(pv.blood_pressure_systolic, 0) as bp_systolic,
			COALESCE(pv.blood_pressure_diastolic, 0) as bp_diastolic,
			COALESCE(pv.pulse_bpm, 0) as pulse,
			COALESCE(pv.temperature_celsius, 0) as temperature,
			COALESCE(pv.height_cm, 0) as height,
			COALESCE(pv.weight_kg, 0) as weight,
			COALESCE(pv.blood_sugar_mg_dl, 0) as blood_sugar,
			COALESCE(pv.bmi, 0) as bmi,
			COALESCE(pv.recorded_at, NOW()) as vitals_recorded_at,
			ROW_NUMBER() OVER (PARTITION BY p.id ORDER BY pv.recorded_at DESC) as vital_rank
		FROM appointments a
		JOIN patients p ON a.patient_id = p.id
		JOIN users u ON p.user_id = u.id
		JOIN beds b ON a.bed_id = b.id AND b.status = 'occupied'
		JOIN departments d ON b.department_id = d.id
		JOIN doctors doc ON a.doctor_id = doc.id
		JOIN users d_user ON doc.user_id = d_user.id
		LEFT JOIN consultations c ON a.id = c.appointment_id
		LEFT JOIN patient_vitals pv ON p.id = pv.patient_id
		WHERE a.status NOT IN ('Cancelled', 'Completed')
		AND a.checked_in_at IS NOT NULL
	`

	rows, err := ns.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch admitted patients: %w", err)
	}
	defer rows.Close()

	var patients []interface{}
	patientMap := make(map[string]interface{})

	for rows.Next() {
		var patientID, patientName, bedID, roomNumber, bedNumber, deptName string
		var doctorID, doctorName, diagnosis string
		var admittedSince time.Time
		var bpSys, bpDia, pulse int
		var temp, height, weight, bloodSugar, bmi float64
		var vitalsTime time.Time
		var vitalRank int

		err := rows.Scan(
			&patientID, &patientName, &bedID, &roomNumber, &bedNumber, &deptName,
			&doctorID, &doctorName, &diagnosis, &admittedSince,
			&bpSys, &bpDia, &pulse, &temp, &height, &weight, &bloodSugar, &bmi, &vitalsTime, &vitalRank,
		)
		if err != nil {
			continue
		}

		if vitalRank != 1 {
			continue
		}

		if _, exists := patientMap[patientID]; exists {
			patients = append(patients, patientMap[patientID])
			continue
		}

		alertLevel := "none"
		alertMessage := ""
		if bpSys > 160 {
			alertLevel = "danger"
			alertMessage = fmt.Sprintf("High BP: %d/%d", bpSys, bpDia)
		} else if (pulse > 100 || pulse < 60) && pulse != 0 {
			alertLevel = "warning"
			alertMessage = fmt.Sprintf("Abnormal Pulse: %d bpm", pulse)
		} else if temp > 38.5 {
			alertLevel = "warning"
			alertMessage = fmt.Sprintf("High Fever: %.1f°C", temp)
		}

		patient := map[string]interface{}{
			"patientId":      patientID,
			"patientName":    patientName,
			"bedId":          bedID,
			"roomNumber":     roomNumber,
			"bedNumber":      bedNumber,
			"departmentName": deptName,
			"doctorId":       doctorID,
			"doctorName":     doctorName,
			"diagnosis":      diagnosis,
			"admittedSince":  admittedSince,
			"alertLevel":     alertLevel,
			"alertMessage":   alertMessage,
			"latestVitals": map[string]interface{}{
				"bloodPressureSystolic":  bpSys,
				"bloodPressureDiastolic": bpDia,
				"pulseBpm":               pulse,
				"temperatureCelsius":     temp,
				"heightCm":               height,
				"weightKg":               weight,
				"bloodSugarMgDl":         bloodSugar,
				"bmi":                    bmi,
				"recordedAt":             vitalsTime,
			},
		}

		patientMap[patientID] = patient
		patients = append(patients, patient)
	}

	return patients, nil
}

// GetNurseStats returns dashboard statistics for a nurse
func (ns *NurseService) GetNurseStats(nurseID string, date string) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	var waitingCount int
	err := ns.db.QueryRow(`
		SELECT COUNT(DISTINCT a.patient_id)
		FROM appointments a
		WHERE DATE(a.appointment_date) = $1
			AND a.status IN ('Checked In', 'Ready for Doctor', 'In Consultation')
			AND NOT EXISTS (
				SELECT 1 FROM patient_vitals pv 
				WHERE pv.patient_id = a.patient_id AND DATE(pv.recorded_at) = $1
			)
	`, date).Scan(&waitingCount)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get waiting count: %w", err)
	}
	stats["patientsWaitingForVitals"] = waitingCount

	var recordedCount int
	err = ns.db.QueryRow(`
		SELECT COUNT(*) FROM patient_vitals 
		WHERE recorded_by = $1 AND DATE(recorded_at) = $2
	`, nurseID, date).Scan(&recordedCount)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get recorded count: %w", err)
	}
	stats["vitalsRecordedToday"] = recordedCount

	var pendingCount int
	err = ns.db.QueryRow(`
		SELECT COUNT(*) FROM nurse_tasks 
		WHERE assigned_to = $1 AND status IN ('Pending', 'In Progress')
	`, nurseID).Scan(&pendingCount)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get pending count: %w", err)
	}
	stats["pendingTasks"] = pendingCount

	var bedsCount int
	err = ns.db.QueryRow(`
		SELECT COUNT(DISTINCT a.patient_id)
		FROM appointments a
		JOIN beds b ON a.bed_id = b.id
		WHERE b.status = 'occupied' AND a.status NOT IN ('Cancelled', 'Completed')
	`).Scan(&bedsCount)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get beds count: %w", err)
	}
	stats["patientsInBeds"] = bedsCount

	return stats, nil
}
