-- Insert doctor details for Dr. Sarah Wilson (doctor@shalby.com)
-- Only run this if the user exists (doctor@shalby.com)
INSERT INTO doctors (user_id, specialization, department, license_number, consultation_fee, joining_date, is_active, created_at)
SELECT u.id, 'Cardiology', 'Cardiology', 'DOC12345', 500.00, '2023-01-15', true, NOW()
FROM users u WHERE u.email = 'doctor@shalby.com'
ON CONFLICT (user_id) DO NOTHING;

-- Insert doctor schedule for Dr. Sarah Wilson - Monday to Friday, 9AM to 5PM
INSERT INTO doctor_schedules (doctor_id, day_of_week, start_time, end_time, slot_duration_minutes, max_patients_per_slot)
SELECT d.id, 1, '09:00:00'::TIME, '17:00:00'::TIME, 30, 1
FROM doctors d JOIN users u ON d.user_id = u.id WHERE u.email = 'doctor@shalby.com'
ON CONFLICT DO NOTHING;

INSERT INTO doctor_schedules (doctor_id, day_of_week, start_time, end_time, slot_duration_minutes, max_patients_per_slot)
SELECT d.id, 2, '09:00:00'::TIME, '17:00:00'::TIME, 30, 1
FROM doctors d JOIN users u ON d.user_id = u.id WHERE u.email = 'doctor@shalby.com'
ON CONFLICT DO NOTHING;

INSERT INTO doctor_schedules (doctor_id, day_of_week, start_time, end_time, slot_duration_minutes, max_patients_per_slot)
SELECT d.id, 3, '09:00:00'::TIME, '17:00:00'::TIME, 30, 1
FROM doctors d JOIN users u ON d.user_id = u.id WHERE u.email = 'doctor@shalby.com'
ON CONFLICT DO NOTHING;

INSERT INTO doctor_schedules (doctor_id, day_of_week, start_time, end_time, slot_duration_minutes, max_patients_per_slot)
SELECT d.id, 4, '09:00:00'::TIME, '17:00:00'::TIME, 30, 1
FROM doctors d JOIN users u ON d.user_id = u.id WHERE u.email = 'doctor@shalby.com'
ON CONFLICT DO NOTHING;

INSERT INTO doctor_schedules (doctor_id, day_of_week, start_time, end_time, slot_duration_minutes, max_patients_per_slot)
SELECT d.id, 5, '09:00:00'::TIME, '17:00:00'::TIME, 30, 1
FROM doctors d JOIN users u ON d.user_id = u.id WHERE u.email = 'doctor@shalby.com'
ON CONFLICT DO NOTHING;
