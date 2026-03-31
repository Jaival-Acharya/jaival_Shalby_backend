-- Seed Hospital Beds
-- This script creates departments and adds 24 hospital beds to the system
-- Organized by department and status

BEGIN;

-- Insert departments if they don't exist
INSERT INTO departments (name, description, type, is_active) VALUES
('General Ward', 'General patient ward for routine care and recovery', 'room', true),
('ICU', 'Intensive Care Unit for critical patients requiring close monitoring', 'room', true),
('Orthopedic', 'Orthopedic ward for bone and joint care', 'room', true),
('Emergency', 'Emergency department for urgent care', 'room', true)
ON CONFLICT (name) DO UPDATE SET type = 'room', description = EXCLUDED.description;

-- Clear existing beds (if needed)
DELETE FROM beds WHERE id IS NOT NULL;

-- Insert beds for General Ward (Room 101-110)
INSERT INTO beds (room_number, bed_number, department_id, capacity, status, created_at) 
SELECT 'ROOM-101', 'A', d.id, 1, 'available'::bed_status_enum, NOW() FROM departments d WHERE d.name = 'General Ward' LIMIT 1;
INSERT INTO beds (room_number, bed_number, department_id, capacity, status, created_at) 
SELECT 'ROOM-102', 'A', d.id, 1, 'available'::bed_status_enum, NOW() FROM departments d WHERE d.name = 'General Ward' LIMIT 1;
INSERT INTO beds (room_number, bed_number, department_id, capacity, status, created_at) 
SELECT 'ROOM-103', 'A', d.id, 1, 'available'::bed_status_enum, NOW() FROM departments d WHERE d.name = 'General Ward' LIMIT 1;
INSERT INTO beds (room_number, bed_number, department_id, capacity, status, created_at) 
SELECT 'ROOM-104', 'A', d.id, 2, 'available'::bed_status_enum, NOW() FROM departments d WHERE d.name = 'General Ward' LIMIT 1;
INSERT INTO beds (room_number, bed_number, department_id, capacity, status, created_at) 
SELECT 'ROOM-105', 'A', d.id, 1, 'reserved'::bed_status_enum, NOW() FROM departments d WHERE d.name = 'General Ward' LIMIT 1;
INSERT INTO beds (room_number, bed_number, department_id, capacity, status, created_at) 
SELECT 'ROOM-106', 'A', d.id, 1, 'available'::bed_status_enum, NOW() FROM departments d WHERE d.name = 'General Ward' LIMIT 1;
INSERT INTO beds (room_number, bed_number, department_id, capacity, status, created_at) 
SELECT 'ROOM-107', 'A', d.id, 1, 'available'::bed_status_enum, NOW() FROM departments d WHERE d.name = 'General Ward' LIMIT 1;
INSERT INTO beds (room_number, bed_number, department_id, capacity, status, created_at) 
SELECT 'ROOM-108', 'A', d.id, 2, 'occupied'::bed_status_enum, NOW() FROM departments d WHERE d.name = 'General Ward' LIMIT 1;
INSERT INTO beds (room_number, bed_number, department_id, capacity, status, created_at) 
SELECT 'ROOM-109', 'A', d.id, 1, 'available'::bed_status_enum, NOW() FROM departments d WHERE d.name = 'General Ward' LIMIT 1;
INSERT INTO beds (room_number, bed_number, department_id, capacity, status, created_at) 
SELECT 'ROOM-110', 'A', d.id, 1, 'maintenance'::bed_status_enum, NOW() FROM departments d WHERE d.name = 'General Ward' LIMIT 1;

-- Insert beds for ICU (Room 201-205)
INSERT INTO beds (room_number, bed_number, department_id, capacity, status, created_at) 
SELECT 'ROOM-201', 'A', d.id, 1, 'available'::bed_status_enum, NOW() FROM departments d WHERE d.name = 'ICU' LIMIT 1;
INSERT INTO beds (room_number, bed_number, department_id, capacity, status, created_at) 
SELECT 'ROOM-202', 'A', d.id, 1, 'reserved'::bed_status_enum, NOW() FROM departments d WHERE d.name = 'ICU' LIMIT 1;
INSERT INTO beds (room_number, bed_number, department_id, capacity, status, created_at) 
SELECT 'ROOM-203', 'A', d.id, 1, 'occupied'::bed_status_enum, NOW() FROM departments d WHERE d.name = 'ICU' LIMIT 1;
INSERT INTO beds (room_number, bed_number, department_id, capacity, status, created_at) 
SELECT 'ROOM-204', 'A', d.id, 1, 'available'::bed_status_enum, NOW() FROM departments d WHERE d.name = 'ICU' LIMIT 1;
INSERT INTO beds (room_number, bed_number, department_id, capacity, status, created_at) 
SELECT 'ROOM-205', 'A', d.id, 1, 'available'::bed_status_enum, NOW() FROM departments d WHERE d.name = 'ICU' LIMIT 1;

-- Insert beds for Orthopedic Ward (Room 301-305)
INSERT INTO beds (room_number, bed_number, department_id, capacity, status, created_at) 
SELECT 'ROOM-301', 'A', d.id, 1, 'available'::bed_status_enum, NOW() FROM departments d WHERE d.name = 'Orthopedic' LIMIT 1;
INSERT INTO beds (room_number, bed_number, department_id, capacity, status, created_at) 
SELECT 'ROOM-302', 'A', d.id, 1, 'available'::bed_status_enum, NOW() FROM departments d WHERE d.name = 'Orthopedic' LIMIT 1;
INSERT INTO beds (room_number, bed_number, department_id, capacity, status, created_at) 
SELECT 'ROOM-303', 'A', d.id, 1, 'reserved'::bed_status_enum, NOW() FROM departments d WHERE d.name = 'Orthopedic' LIMIT 1;
INSERT INTO beds (room_number, bed_number, department_id, capacity, status, created_at) 
SELECT 'ROOM-304', 'A', d.id, 1, 'occupied'::bed_status_enum, NOW() FROM departments d WHERE d.name = 'Orthopedic' LIMIT 1;
INSERT INTO beds (room_number, bed_number, department_id, capacity, status, created_at) 
SELECT 'ROOM-305', 'A', d.id, 1, 'available'::bed_status_enum, NOW() FROM departments d WHERE d.name = 'Orthopedic' LIMIT 1;

-- Insert beds for Cardiology Ward (Room 401-404)
INSERT INTO beds (room_number, bed_number, department_id, capacity, status, created_at) 
SELECT 'ROOM-401', 'A', d.id, 1, 'available'::bed_status_enum, NOW() FROM departments d WHERE d.name = 'Cardiology' LIMIT 1;
INSERT INTO beds (room_number, bed_number, department_id, capacity, status, created_at) 
SELECT 'ROOM-402', 'A', d.id, 1, 'available'::bed_status_enum, NOW() FROM departments d WHERE d.name = 'Cardiology' LIMIT 1;
INSERT INTO beds (room_number, bed_number, department_id, capacity, status, created_at) 
SELECT 'ROOM-403', 'A', d.id, 1, 'available'::bed_status_enum, NOW() FROM departments d WHERE d.name = 'Cardiology' LIMIT 1;
INSERT INTO beds (room_number, bed_number, department_id, capacity, status, created_at) 
SELECT 'ROOM-404', 'A', d.id, 1, 'occupied'::bed_status_enum, NOW() FROM departments d WHERE d.name = 'Cardiology' LIMIT 1;

-- Insert beds for Emergency Ward (Room 501-504)
INSERT INTO beds (room_number, bed_number, department_id, capacity, status, created_at) 
SELECT 'ROOM-501', 'A', d.id, 1, 'available'::bed_status_enum, NOW() FROM departments d WHERE d.name = 'Emergency' LIMIT 1;
INSERT INTO beds (room_number, bed_number, department_id, capacity, status, created_at) 
SELECT 'ROOM-502', 'A', d.id, 1, 'reserved'::bed_status_enum, NOW() FROM departments d WHERE d.name = 'Emergency' LIMIT 1;
INSERT INTO beds (room_number, bed_number, department_id, capacity, status, created_at) 
SELECT 'ROOM-503', 'A', d.id, 1, 'occupied'::bed_status_enum, NOW() FROM departments d WHERE d.name = 'Emergency' LIMIT 1;
INSERT INTO beds (room_number, bed_number, department_id, capacity, status, created_at) 
SELECT 'ROOM-504', 'A', d.id, 1, 'available'::bed_status_enum, NOW() FROM departments d WHERE d.name = 'Emergency' LIMIT 1;

COMMIT;

-- Verify inserted beds
SELECT COUNT(*) as total_beds, 
       SUM(CASE WHEN status = 'available' THEN 1 ELSE 0 END) as available,
       SUM(CASE WHEN status = 'occupied' THEN 1 ELSE 0 END) as occupied,
       SUM(CASE WHEN status = 'reserved' THEN 1 ELSE 0 END) as reserved,
       SUM(CASE WHEN status = 'maintenance' THEN 1 ELSE 0 END) as maintenance
FROM beds;

