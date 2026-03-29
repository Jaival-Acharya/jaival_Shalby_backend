package services

import (
	"database/sql"
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
