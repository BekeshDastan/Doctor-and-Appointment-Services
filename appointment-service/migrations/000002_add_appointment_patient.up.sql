ALTER TABLE appointments ADD COLUMN patient_name TEXT NOT NULL DEFAULT '';

CREATE INDEX idx_appointments_doctor_id ON appointments (doctor_id);
