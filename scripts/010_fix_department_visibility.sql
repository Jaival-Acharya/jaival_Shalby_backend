-- Migration 010: Fix Department Visibility
-- Date: 2026-03-30
-- Purpose: 
--   1. Set type='room' for bed-related departments
--   2. Add descriptions to staff departments
--   3. Ensure doctors are linked to departments via department_id

BEGIN;

-- 1. Update room departments to have type='room'
UPDATE departments 
SET type = 'room' 
WHERE name IN ('General Ward', 'ICU', 'Orthopedic', 'Emergency', 'Cardiology Ward', 'Emergency Ward')
  AND (type IS NULL OR type != 'room');

-- 2. Update staff departments to have type='staff' and add descriptions
UPDATE departments 
SET type = 'staff', description = COALESCE(description, 'Medical Department')
WHERE name IN ('Cardiology', 'Pediatrics', 'Orthopedics', 'Neurology', 'Gastroenterology', 'Neurology', 'General')
  AND (type IS NULL OR type != 'staff');

-- 3. Ensure all departments have the type field set
UPDATE departments 
SET type = CASE 
  WHEN name LIKE '%Ward%' OR name IN ('ICU', 'Emergency') THEN 'room'
  ELSE 'staff'
END
WHERE type IS NULL OR type NOT IN ('room', 'staff');

-- 4. Connect doctors to departments via department_id (if not already set)
-- For doctors with department name but no department_id
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

-- 5. If no matching department exists, create one as staff type
INSERT INTO departments (name, type, is_active)
SELECT DISTINCT d.department, 'staff', true
FROM doctors d
WHERE d.department IS NOT NULL
  AND NOT EXISTS (
    SELECT 1 FROM departments WHERE name = d.department
  )
ON CONFLICT (name) DO NOTHING;

-- 6. Now update doctors again with the newly created departments
UPDATE doctors d
SET department_id = (
  SELECT id FROM departments 
  WHERE name = d.department LIMIT 1
)
WHERE department_id IS NULL 
  AND d.department IS NOT NULL;

COMMIT;

-- Verify the changes
SELECT COUNT(*) as total_departments,
       SUM(CASE WHEN type = 'room' THEN 1 ELSE 0 END) as room_departments,
       SUM(CASE WHEN type = 'staff' THEN 1 ELSE 0 END) as staff_departments
FROM departments;

SELECT COUNT(*) as total_doctors,
       SUM(CASE WHEN department_id IS NOT NULL THEN 1 ELSE 0 END) as doctors_with_dept_id
FROM doctors;

-- Show departments with their types
SELECT id, name, type, description, is_active 
FROM departments 
ORDER BY type DESC, name ASC;
