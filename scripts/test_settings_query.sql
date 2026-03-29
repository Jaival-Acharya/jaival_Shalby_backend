-- Test query to check if system_settings data exists
-- Run this to verify data is in the database

SELECT 
  COUNT(*) as total_settings,
  MIN(setting_key) as first_setting_key,
  MAX(setting_key) as last_setting_key
FROM system_settings;

-- List all settings
SELECT setting_key, setting_value FROM system_settings ORDER BY setting_key;

--Check specific settings we care about
SELECT setting_key, setting_value 
FROM system_settings 
WHERE setting_key IN ('hospital_name', 'hospital_email', 'hospital_phone', 'currency', 'timezone')
ORDER BY setting_key;

-- If table doesn't exist, it will show error
-- If no data, it will show 0 count
-- If existing data found, you'll see all settings listed
