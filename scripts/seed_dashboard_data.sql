-- Dashboard Data Seeding Script for Shalby Hospital HMS
-- This script inserts sample data for the admin dashboard
-- Last updated: March 29, 2026

-- ============================================
-- 1. INSERT SAMPLE PATIENTS
-- ============================================
INSERT INTO patients (id, user_id, date_of_birth, blood_group, allergies, chronic_conditions, emergency_contact, emergency_phone, address, created_at, updated_at)
SELECT 
  gen_random_uuid(),
  u.id,
  CURRENT_DATE - (RANDOM() * 20000)::int,
  (ARRAY['A+', 'A-', 'B+', 'B-', 'O+', 'O-', 'AB+', 'AB-'])[floor(random() * 8) + 1],
  'None',
  'None',
  'Emergency Contact ' || u.name,
  '+91-9999-99999' || SUBSTRING(u.id::text FROM 1 FOR 4),
  'Patient Address ' || u.id::text,
  NOW(),
  NOW()
FROM users u
WHERE u.role = 'Patient'
  AND NOT EXISTS (SELECT 1 FROM patients p WHERE p.user_id = u.id)
LIMIT 50;

-- ============================================
-- 2. INSERT SAMPLE DOCTORS
-- ============================================
INSERT INTO doctors (id, user_id, specialization_id, department_id, qualification, license_number, years_of_experience, bio, is_active, created_at, updated_at)
SELECT
  gen_random_uuid(),
  u.id,
  s.id,
  d.id,
  'MBBS, MD',
  'LIC-' || SUBSTRING(u.id::text FROM 1 FOR 8),
  (RANDOM() * 25 + 1)::int,
  'Experienced doctor in ' || s.name,
  true,
  NOW(),
  NOW()
FROM users u
CROSS JOIN (SELECT DISTINCT id, name FROM specializations LIMIT 5) s
CROSS JOIN (SELECT DISTINCT id FROM departments LIMIT 3) d
WHERE u.role = 'Doctor'
  AND NOT EXISTS (SELECT 1 FROM doctors doc WHERE doc.user_id = u.id)
LIMIT 30;

-- ============================================
-- 3. INSERT SAMPLE BEDS
-- ============================================
INSERT INTO beds (id, bed_number, room_number, floor, department_id, bed_type, status, created_at, updated_at)
SELECT
  gen_random_uuid(),
  'BED-' || SUBSTRING(d.id::text FROM 1 FOR 4) || '-' || (ROW_NUMBER() OVER (PARTITION BY d.id ORDER BY d.id))::text,
  'ROOM-' || (ROW_NUMBER() OVER (PARTITION BY d.id ORDER BY d.id) / 4)::text,
  ((ROW_NUMBER() OVER (PARTITION BY d.id ORDER BY d.id) - 1) / 4 + 1)::int,
  d.id,
  (ARRAY['General', 'Private', 'ICU', 'HDU'])[floor(random() * 4) + 1],
  (ARRAY['available', 'occupied', 'maintenance'])[floor(random() * 3 + 1)],
  NOW(),
  NOW()
FROM (
  SELECT DISTINCT id FROM departments
) d
CROSS JOIN LATERAL (SELECT 1 FROM generate_series(1, 20)) num
WHERE NOT EXISTS (SELECT 1 FROM beds WHERE beds.department_id = d.id);

-- ============================================
-- 4. INSERT SAMPLE APPOINTMENTS
-- ============================================
INSERT INTO appointments (id, patient_id, doctor_id, appointment_date, time_slot, status, reason, notes, created_at, updated_at)
SELECT
  gen_random_uuid(),
  p.id,
  d.id,
  CURRENT_DATE + (RANDOM() * 30)::int,
  ARRAY['09:00 AM', '10:00 AM', '11:00 AM', '02:00 PM', '03:00 PM', '04:00 PM'][floor(random() * 6) + 1],
  (ARRAY['Scheduled', 'Checked In', 'Ready for Doctor', 'In Consultation', 'Completed', 'Cancelled'])[floor(random() * 6 + 1)],
  'Routine Check-up',
  'Patient arrived on time',
  NOW() - (RANDOM() * 100)::int * INTERVAL '1 day',
  NOW()
FROM (SELECT DISTINCT patient_id as id FROM appointments UNION SELECT id FROM patients LIMIT 40) p
CROSS JOIN (SELECT DISTINCT id FROM doctors LIMIT 15) d
WHERE NOT EXISTS (
  SELECT 1 FROM appointments a 
  WHERE a.patient_id = p.id AND a.doctor_id = d.id
)
LIMIT 100;

-- ============================================
-- 5. INSERT SAMPLE PRESCRIPTIONS  
-- ============================================
INSERT INTO prescriptions (id, appointment_id, medicine_ids, instructions, created_at, updated_at)
SELECT
  gen_random_uuid(),
  a.id,
  ARRAY['med-' || SUBSTRING(gen_random_uuid()::text FROM 1 FOR 8)],
  'Take 1 tablet twice daily with water after meals',
  NOW(),
  NOW()
FROM appointments a
WHERE a.status = 'Completed'
  AND NOT EXISTS (SELECT 1 FROM prescriptions p WHERE p.appointment_id = a.id)
LIMIT 50;

-- ============================================
-- 6. INSERT SAMPLE MEDICINES
-- ============================================
INSERT INTO medicines (id, name, generic_name, category_id, dosage_form, strength, unit, stock_quantity, reorder_level, price, manufacturer, expiry_date, is_active, created_at, updated_at)
SELECT
  gen_random_uuid(),
  'Medicine ' || SUBSTRING(gen_random_uuid()::text FROM 1 FOR 8),
  'Generic Name ' || SUBSTRING(gen_random_uuid()::text FROM 1 FOR 6),
  c.id,
  (ARRAY['Tablet', 'Capsule', 'Syrup', 'Injection', 'Ointment'])[floor(random() * 5 + 1)],
  (ARRAY['250mg', '500mg', '1000mg', '2mg', '5mg'])[floor(random() * 5 + 1)],
  'units',
  (RANDOM() * 1000 + 50)::int,
  100,
  (RANDOM() * 500 + 10)::numeric(10, 2),
  'Pharma Company ' || (floor(random() * 10) + 1)::text,
  CURRENT_DATE + (RANDOM() * 365)::int,
  true,
  NOW(),
  NOW()
FROM (SELECT DISTINCT id FROM medicine_categories LIMIT 10) c
CROSS JOIN LATERAL (SELECT 1 FROM generate_series(1, 10)) num
WHERE NOT EXISTS (SELECT 1 FROM medicines WHERE medicines.category_id = c.id LIMIT 10);

-- ============================================
-- 7. INSERT SAMPLE VITAL READINGS
-- ============================================
INSERT INTO vital_readings (id, patient_id, temperature, blood_pressure_systolic, blood_pressure_diastolic, pulse_rate, respiratory_rate, oxygen_saturation, weight, height, recorded_at, created_at, updated_at)
SELECT
  gen_random_uuid(),
  p.id,
  (96 + RANDOM() * 4)::numeric(5, 1),
  (110 + RANDOM() * 40)::int,
  (70 + RANDOM() * 30)::int,
  (60 + RANDOM() * 40)::int,
  (12 + RANDOM() * 8)::int,
  (95 + RANDOM() * 5)::numeric(5, 1),
  (50 + RANDOM() * 50)::numeric(6, 2),
  (150 + RANDOM() * 40)::numeric(6, 2),
  NOW() - (RANDOM() * 30)::int * INTERVAL '1 day',
  NOW(),
  NOW()
FROM (SELECT DISTINCT id FROM patients LIMIT 60) p
WHERE NOT EXISTS (SELECT 1 FROM vital_readings vr WHERE vr.patient_id = p.id);

-- ============================================
-- 8. VERIFY DATA INSERTION
-- ============================================
SELECT
  (SELECT COUNT(*) FROM patients) as total_patients,
  (SELECT COUNT(*) FROM doctors WHERE is_active = true) as active_doctors,
  (SELECT COUNT(*) FROM appointments) as total_appointments,
  (SELECT COUNT(*) FROM appointments WHERE DATE(appointment_date) = CURRENT_DATE) as today_appointments,
  (SELECT COUNT(*) FROM beds WHERE status = 'available') as available_beds,
  (SELECT COUNT(*) FROM beds WHERE status = 'occupied') as occupied_beds,
  (SELECT COUNT(*) FROM prescriptions) as total_prescriptions,
  (SELECT COUNT(*) FROM medicines WHERE is_active = true) as active_medicines;

-- ============================================
-- Summary
-- ============================================
-- This script creates:
-- - 50 sample patients
-- - 30 sample doctors 
-- - 20 beds per department (auto-calculated)
-- - 100 appointments with various statuses
-- - 50 prescriptions (linked to completed appointments)
-- - 100 medicines with different categories
-- - 60 vital readings
--
-- The dashboard will now display real data instead of empty values
-- Fallback values ensure no errors even if some data is missing
-- ============================================
