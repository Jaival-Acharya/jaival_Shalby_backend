-- COMPREHENSIVE DATABASE MIGRATION SCRIPT
-- This script fixes all database schema issues for full system functionality
-- Date: 2026-03-30

-- =====================================================
-- 1. ADD DEPARTMENT CHECK-IN COLUMNS (Migration 011)
-- =====================================================
ALTER TABLE appointments
ADD COLUMN IF NOT EXISTS bed_id UUID REFERENCES beds(id),
ADD COLUMN IF NOT EXISTS checked_in_at TIMESTAMP,
ADD COLUMN IF NOT EXISTS checked_in_by UUID REFERENCES users(id);

CREATE INDEX IF NOT EXISTS idx_appointments_bed ON appointments(bed_id);
CREATE INDEX IF NOT EXISTS idx_appointments_checked_in_at ON appointments(checked_in_at);

-- =====================================================
-- 2. FIX DEPARTMENT VISIBILITY (Migration 010)
-- =====================================================

-- Update room departments to have type='room'
UPDATE departments 
SET type = 'room' 
WHERE name IN ('General Ward', 'ICU', 'Orthopedic', 'Emergency', 'Cardiology Ward', 'Emergency Ward')
  AND (type IS NULL OR type != 'room');

-- Update staff departments to have type='staff' and add descriptions
UPDATE departments 
SET type = 'staff', description = COALESCE(description, 'Medical Department')
WHERE name IN ('Cardiology', 'Pediatrics', 'Orthopedics', 'Neurology', 'Gastroenterology', 'General')
  AND (type IS NULL OR type != 'staff');

-- Ensure all departments have the type field set
UPDATE departments 
SET type = CASE 
  WHEN name LIKE '%Ward%' OR name IN ('ICU', 'Emergency') THEN 'room'
  ELSE 'staff'
END
WHERE type IS NULL OR type NOT IN ('room', 'staff');

-- Connect doctors to departments via department_id (if not already set)
UPDATE doctors d
SET department_id = (
  SELECT id FROM departments 
  WHERE name = d.department LIMIT 1
)
WHERE department_id IS NULL 
  AND d.department IS NOT NULL
  AND EXISTS (
    SELECT 1 FROM departments WHERE name = d.department
  );

-- Create any missing departments as staff type
INSERT INTO departments (name, type, is_active)
SELECT DISTINCT d.department, 'staff', true
FROM doctors d
WHERE d.department IS NOT NULL
  AND NOT EXISTS (
    SELECT 1 FROM departments WHERE name = d.department
  )
ON CONFLICT (name) DO NOTHING;

-- Now update doctors again with the newly created departments
UPDATE doctors d
SET department_id = (
  SELECT id FROM departments 
  WHERE name = d.department LIMIT 1
)
WHERE department_id IS NULL 
  AND d.department IS NOT NULL;

-- =====================================================
-- 3. VERIFICATION QUERIES
-- =====================================================

-- Show all columns in appointments table  
SELECT column_name, data_type, is_nullable
FROM information_schema.columns
WHERE table_name = 'appointments'
ORDER BY ordinal_position;

-- Show department types and their counts
SELECT type, COUNT(*) as count, GROUP_CONCAT(name) as departments
FROM departments
GROUP BY type
ORDER BY type;

-- Show doctors with department_id assignment
SELECT COUNT(*) as total_doctors,
       COUNT(department_id) as doctors_with_dept_id,
       COUNT(CASE WHEN department_id IS NULL THEN 1 END) as doctors_without_dept_id
FROM doctors;

-- Show all beds with their departments
SELECT COUNT(*) as total_beds,
       COUNT(DISTINCT department_id) as unique_departments
FROM beds;
