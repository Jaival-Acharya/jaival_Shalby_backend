package services

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
)

// PrescriptionItem represents a medicine in prescription
type PrescriptionItem struct {
	ID                  string `json:"id"`
	PrescriptionID      string `json:"prescriptionId"`
	MedicineID          string `json:"medicineId"`
	MedicineName        string `json:"medicineName"`
	Strength            string `json:"strength"`
	DosageForm          string `json:"dosageForm"`
	Frequency           string `json:"frequency"`
	Timing              string `json:"timing"`
	Duration            string `json:"duration"`
	SpecialInstructions string `json:"specialInstructions"`
	QuantityDispensed   int    `json:"quantityDispensed"`
}

// PrescriptionDetails contains prescription with items
type PrescriptionDetails struct {
	ID              string             `json:"id"`
	ConsultationID  string             `json:"consultationId"`
	PatientID       string             `json:"patientId"`
	PatientName     string             `json:"patientName"`
	DoctorID        string             `json:"doctorId"`
	DoctorName      string             `json:"doctorName"`
	Status          string             `json:"status"` // Pending, Dispensed
	Items           []PrescriptionItem `json:"items"`
	DispensingNotes string             `json:"dispensingNotes"`
	DispensedAt     *time.Time         `json:"dispensedAt"`
	DispensedBy     *string            `json:"dispensedBy"`
	CreatedAt       time.Time          `json:"createdAt"`
	UpdatedAt       time.Time          `json:"updatedAt"`
}

// CreatePrescriptionRequest contains data to create prescription
type CreatePrescriptionRequest struct {
	ConsultationID string                          `json:"consultationId" binding:"required"`
	PatientID      string                          `json:"patientId" binding:"required"`
	DoctorID       string                          `json:"doctorId" binding:"required"`
	Items          []CreatePrescriptionItemRequest `json:"items" binding:"required"`
}

// CreatePrescriptionItemRequest contains medicine for prescription
type CreatePrescriptionItemRequest struct {
	MedicineID          string `json:"medicineId" binding:"required"`
	MedicineName        string `json:"medicineName" binding:"required"`
	Strength            string `json:"strength"`
	DosageForm          string `json:"dosageForm" binding:"required"`
	Frequency           string `json:"frequency" binding:"required"`
	Timing              string `json:"timing"`
	Duration            string `json:"duration" binding:"required"`
	SpecialInstructions string `json:"specialInstructions"`
	Quantity            int    `json:"quantity"`
}

// CreatePrescription creates a new prescription
func CreatePrescription(db *sql.DB, req CreatePrescriptionRequest) (*PrescriptionDetails, error) {
	prescriptionID := uuid.New().String()
	now := time.Now()

	// Validate medicines exist
	for _, item := range req.Items {
		if err := ValidatePrescriptionMedicine(db, item.MedicineID); err != nil {
			return nil, fmt.Errorf("medicine validation failed: %v", err)
		}
	}

	// Get doctor and patient names
	var doctorName, patientName string
	err := db.QueryRow(`
		SELECT u.name FROM users u
		JOIN doctors d ON u.id = d.user_id
		WHERE d.id = $1
	`, req.DoctorID).Scan(&doctorName)
	if err != nil {
		log.Println("Error getting doctor:", err)
		return nil, fmt.Errorf("failed to get doctor")
	}

	err = db.QueryRow(`
		SELECT u.name FROM users u
		JOIN patients p ON u.id = p.user_id
		WHERE p.id = $1
	`, req.PatientID).Scan(&patientName)
	if err != nil {
		log.Println("Error getting patient:", err)
		return nil, fmt.Errorf("failed to get patient")
	}

	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		log.Println("Error starting transaction:", err)
		return nil, fmt.Errorf("transaction error")
	}
	defer tx.Rollback()

	// Create prescription
	_, err = tx.Exec(`
		INSERT INTO prescriptions 
		(id, consultation_id, patient_id, doctor_id, pharmacy_status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, prescriptionID, req.ConsultationID, req.PatientID, req.DoctorID, "Pending", now, now)

	if err != nil {
		log.Println("Error creating prescription:", err)
		return nil, fmt.Errorf("failed to create prescription")
	}

	// Create prescription items
	items := make([]PrescriptionItem, 0)
	for _, itemReq := range req.Items {
		itemID := uuid.New().String()
		_, err := tx.Exec(`
			INSERT INTO prescription_items 
			(id, prescription_id, medicine_id, medicine_name, strength, dosage_form, frequency, timing, duration, special_instructions, quantity_dispensed)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		`, itemID, prescriptionID, itemReq.MedicineID, itemReq.MedicineName, itemReq.Strength, itemReq.DosageForm,
			itemReq.Frequency, itemReq.Timing, itemReq.Duration, itemReq.SpecialInstructions, 0)

		if err != nil {
			log.Println("Error creating prescription item:", err)
			tx.Rollback()
			return nil, fmt.Errorf("failed to add medicine to prescription")
		}

		items = append(items, PrescriptionItem{
			ID:                  itemID,
			PrescriptionID:      prescriptionID,
			MedicineID:          itemReq.MedicineID,
			MedicineName:        itemReq.MedicineName,
			Strength:            itemReq.Strength,
			DosageForm:          itemReq.DosageForm,
			Frequency:           itemReq.Frequency,
			Timing:              itemReq.Timing,
			Duration:            itemReq.Duration,
			SpecialInstructions: itemReq.SpecialInstructions,
		})
	}

	if err = tx.Commit(); err != nil {
		log.Println("Error committing transaction:", err)
		return nil, fmt.Errorf("transaction commit failed")
	}

	return &PrescriptionDetails{
		ID:             prescriptionID,
		ConsultationID: req.ConsultationID,
		PatientID:      req.PatientID,
		PatientName:    patientName,
		DoctorID:       req.DoctorID,
		DoctorName:     doctorName,
		Status:         "Pending",
		Items:          items,
		CreatedAt:      now,
		UpdatedAt:      now,
	}, nil
}

// GetPrescriptionByID gets a prescription with items
func GetPrescriptionByID(db *sql.DB, prescriptionID string) (*PrescriptionDetails, error) {
	detail := &PrescriptionDetails{}

	var dispensedAt sql.NullTime

	err := db.QueryRow(`
		SELECT p.id, p.consultation_id, p.patient_id, u2.name, p.doctor_id, u1.name,
		       p.pharmacy_status, p.dispensed_at, p.dispensed_by, p.created_at, p.updated_at
		FROM prescriptions p
		JOIN doctors d ON p.doctor_id = d.id
		JOIN users u1 ON d.user_id = u1.id
		JOIN patients pat ON p.patient_id = pat.id
		JOIN users u2 ON pat.user_id = u2.id
		WHERE p.id = $1
	`, prescriptionID).Scan(
		&detail.ID, &detail.ConsultationID, &detail.PatientID, &detail.PatientName,
		&detail.DoctorID, &detail.DoctorName, &detail.Status, &dispensedAt, &detail.DispensedBy,
		&detail.CreatedAt, &detail.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("prescription not found")
	}
	if err != nil {
		log.Println("Error querying prescription:", err)
		return nil, fmt.Errorf("failed to get prescription")
	}

	if dispensedAt.Valid {
		detail.DispensedAt = &dispensedAt.Time
	}

	// Get items
	rows, err := db.Query(`
		SELECT id, prescription_id, medicine_id, medicine_name, strength, dosage_form,
		       frequency, timing, duration, special_instructions, quantity_dispensed
		FROM prescription_items
		WHERE prescription_id = $1
	`, prescriptionID)
	if err != nil {
		log.Println("Error querying items:", err)
		return nil, fmt.Errorf("failed to get prescription items")
	}
	defer rows.Close()

	var items []PrescriptionItem
	for rows.Next() {
		item := PrescriptionItem{}
		if err := rows.Scan(&item.ID, &item.PrescriptionID, &item.MedicineID, &item.MedicineName,
			&item.Strength, &item.DosageForm, &item.Frequency, &item.Timing,
			&item.Duration, &item.SpecialInstructions, &item.QuantityDispensed); err != nil {
			log.Println("Error scanning item:", err)
			continue
		}
		items = append(items, item)
	}

	detail.Items = items
	return detail, nil
}

// GetPrescriptionsByDoctor gets prescriptions created by a doctor
func GetPrescriptionsByDoctor(db *sql.DB, doctorID, status string, limit, offset int) ([]PrescriptionDetails, int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM prescriptions WHERE doctor_id = $1`
	args := []interface{}{doctorID}

	if status != "" && status != "all" {
		query += ` AND pharmacy_status = $2`
		args = append(args, status)
	}

	err := db.QueryRow(query, args...).Scan(&count)
	if err != nil {
		log.Println("Error counting prescriptions:", err)
		return nil, 0, err
	}

	query = `
		SELECT p.id, p.consultation_id, p.patient_id, u2.name, p.doctor_id, u1.name,
		       p.pharmacy_status, p.dispensed_at, p.dispensed_by, p.created_at, p.updated_at
		FROM prescriptions p
		JOIN doctors d ON p.doctor_id = d.id
		JOIN users u1 ON d.user_id = u1.id
		JOIN patients pat ON p.patient_id = pat.id
		JOIN users u2 ON pat.user_id = u2.id
		WHERE p.doctor_id = $1
	`

	if status != "" && status != "all" {
		query += ` AND p.pharmacy_status = $2`
	}

	query += ` ORDER BY p.created_at DESC LIMIT $` + fmt.Sprintf("%d", len(args)+1) + ` OFFSET $` + fmt.Sprintf("%d", len(args)+2)
	args = append(args, limit, offset)

	rows, err := db.Query(query, args...)
	if err != nil {
		log.Println("Error querying prescriptions:", err)
		return nil, 0, err
	}
	defer rows.Close()

	var prescriptions []PrescriptionDetails
	for rows.Next() {
		detail := PrescriptionDetails{}
		var dispensedAt sql.NullTime

		if err := rows.Scan(&detail.ID, &detail.ConsultationID, &detail.PatientID, &detail.PatientName,
			&detail.DoctorID, &detail.DoctorName, &detail.Status, &dispensedAt, &detail.DispensedBy,
			&detail.CreatedAt, &detail.UpdatedAt); err != nil {
			log.Println("Error scanning prescription:", err)
			continue
		}

		if dispensedAt.Valid {
			detail.DispensedAt = &dispensedAt.Time
		}

		// Get items
		itemRows, err := db.Query(`
			SELECT id, prescription_id, medicine_id, medicine_name, strength, dosage_form,
			       frequency, timing, duration, special_instructions, quantity_dispensed
			FROM prescription_items
			WHERE prescription_id = $1
		`, detail.ID)
		if err == nil {
			var items []PrescriptionItem
			for itemRows.Next() {
				item := PrescriptionItem{}
				if err := itemRows.Scan(&item.ID, &item.PrescriptionID, &item.MedicineID, &item.MedicineName,
					&item.Strength, &item.DosageForm, &item.Frequency, &item.Timing,
					&item.Duration, &item.SpecialInstructions, &item.QuantityDispensed); err == nil {
					items = append(items, item)
				}
			}
			itemRows.Close()
			detail.Items = items
		}

		prescriptions = append(prescriptions, detail)
	}

	return prescriptions, count, nil
}

// GetPendingPrescriptions gets all pending prescriptions
func GetPendingPrescriptions(db *sql.DB, limit, offset int) ([]PrescriptionDetails, int64, error) {
	var count int64
	db.QueryRow(`SELECT COUNT(*) FROM prescriptions WHERE pharmacy_status = 'Pending'`).Scan(&count)

	rows, err := db.Query(`
		SELECT p.id, p.consultation_id, p.patient_id, u2.name, p.doctor_id, u1.name,
		       p.pharmacy_status, p.dispensed_at, p.dispensed_by, p.created_at, p.updated_at
		FROM prescriptions p
		JOIN doctors d ON p.doctor_id = d.id
		JOIN users u1 ON d.user_id = u1.id
		JOIN patients pat ON p.patient_id = pat.id
		JOIN users u2 ON pat.user_id = u2.id
		WHERE p.pharmacy_status = 'Pending'
		ORDER BY p.created_at ASC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		log.Println("Error querying prescriptions:", err)
		return nil, 0, err
	}
	defer rows.Close()

	var prescriptions []PrescriptionDetails
	for rows.Next() {
		detail := PrescriptionDetails{}
		var dispensedAt sql.NullTime

		if err := rows.Scan(&detail.ID, &detail.ConsultationID, &detail.PatientID, &detail.PatientName,
			&detail.DoctorID, &detail.DoctorName, &detail.Status, &dispensedAt, &detail.DispensedBy,
			&detail.CreatedAt, &detail.UpdatedAt); err != nil {
			log.Println("Error scanning prescription:", err)
			continue
		}

		if dispensedAt.Valid {
			detail.DispensedAt = &dispensedAt.Time
		}

		prescriptions = append(prescriptions, detail)
	}

	return prescriptions, count, nil
}

// DispensePrescription marks prescription as dispensed and deducts stock
func DispensePrescription(db *sql.DB, prescriptionID string, dispensedByID string, items []struct {
	MedicineID        string
	QuantityDispensed int
}) (*PrescriptionDetails, error) {
	// Get prescription
	prescription, err := GetPrescriptionByID(db, prescriptionID)
	if err != nil {
		return nil, err
	}

	if prescription.Status != "Pending" {
		return nil, fmt.Errorf("prescription is not pending")
	}

	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		return nil, fmt.Errorf("transaction error")
	}
	defer tx.Rollback()

	now := time.Now()

	// Update each item and deduct stock
	itemMap := make(map[string]int)
	for _, item := range items {
		itemMap[item.MedicineID] = item.QuantityDispensed
	}

	for _, pItem := range prescription.Items {
		qty := itemMap[pItem.MedicineID]
		if qty <= 0 {
			qty = 1 // Default to 1 if not specified
		}

		// Check stock first
		var currentStock int
		err := tx.QueryRow(`
			SELECT stock_quantity FROM medicine_inventory WHERE medicine_id = $1
		`, pItem.MedicineID).Scan(&currentStock)
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to get stock for medicine %s", pItem.MedicineID)
		}

		if currentStock < qty {
			tx.Rollback()
			return nil, fmt.Errorf("insufficient stock for %s: have %d, need %d", pItem.MedicineName, currentStock, qty)
		}

		// Deduct stock
		_, err = tx.Exec(`
			UPDATE medicine_inventory
			SET stock_quantity = stock_quantity - $1, updated_at = NOW()
			WHERE medicine_id = $2
		`, qty, pItem.MedicineID)
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("stock deduction failed")
		}

		// Update prescription item
		_, err = tx.Exec(`
			UPDATE prescription_items
			SET quantity_dispensed = $1
			WHERE prescription_id = $2 AND medicine_id = $3
		`, qty, prescriptionID, pItem.MedicineID)
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to update item")
		}
	}

	// Update prescription status
	_, err = tx.Exec(`
		UPDATE prescriptions
		SET pharmacy_status = 'Dispensed', dispensed_by = $1, dispensed_at = $2, updated_at = $3
		WHERE id = $4
	`, dispensedByID, now, now, prescriptionID)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to update prescription status")
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("transaction commit failed")
	}

	return GetPrescriptionByID(db, prescriptionID)
}

// DeductMedicineStock deducts medicine from inventory
func DeductMedicineStock(db *sql.DB, medicineID string, quantity int) error {
	var currentStock int
	err := db.QueryRow(`
		SELECT stock_quantity FROM medicine_inventory WHERE medicine_id = $1
	`, medicineID).Scan(&currentStock)

	if err != nil {
		log.Println("Error getting stock:", err)
		return fmt.Errorf("failed to get medicine stock")
	}

	if currentStock < quantity {
		return fmt.Errorf("insufficient stock: have %d, need %d", currentStock, quantity)
	}

	_, err = db.Exec(`
		UPDATE medicine_inventory
		SET stock_quantity = stock_quantity - $1, updated_at = NOW()
		WHERE medicine_id = $2
	`, quantity, medicineID)

	return err
}

// ValidatePrescriptionMedicine validates medicine exists and has stock
func ValidatePrescriptionMedicine(db *sql.DB, medicineID string) error {
	var stock int
	err := db.QueryRow(`
		SELECT COALESCE(stock_quantity, 0) FROM medicine_inventory WHERE medicine_id = $1
	`, medicineID).Scan(&stock)

	if err == sql.ErrNoRows {
		return fmt.Errorf("medicine not found")
	}
	if err != nil {
		return fmt.Errorf("failed to validate medicine")
	}

	return nil
}
