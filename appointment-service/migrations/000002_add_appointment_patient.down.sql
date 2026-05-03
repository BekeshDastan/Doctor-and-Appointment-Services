DROP INDEX IF EXISTS idx_appointments_doctor_id;

ALTER TABLE appointments DROP COLUMN patient_name;
