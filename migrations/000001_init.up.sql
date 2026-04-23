
CREATE EXTENSION IF NOT EXISTS "pgcrypto";   -- for gen_random_uuid()

CREATE TYPE user_role AS ENUM ('admin', 'doctor', 'patient');

CREATE TABLE IF NOT EXISTS users (
    id            UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    name          VARCHAR(100) NOT NULL,
    email         VARCHAR(255) NOT NULL UNIQUE,
    password_hash TEXT         NOT NULL,
    role          user_role    NOT NULL DEFAULT 'patient',
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_role  ON users(role);

CREATE TYPE slot_status AS ENUM ('available', 'booked', 'cancelled');

CREATE TABLE IF NOT EXISTS time_slots (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    doctor_id   UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    start_time  TIMESTAMPTZ NOT NULL,
    end_time    TIMESTAMPTZ NOT NULL,
    status      slot_status NOT NULL DEFAULT 'available',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_slot_times CHECK (end_time > start_time)
);

CREATE INDEX IF NOT EXISTS idx_slots_doctor_id   ON time_slots(doctor_id);
CREATE INDEX IF NOT EXISTS idx_slots_status      ON time_slots(status);
CREATE INDEX IF NOT EXISTS idx_slots_start_time  ON time_slots(start_time);

CREATE EXTENSION IF NOT EXISTS btree_gist;

ALTER TABLE time_slots
    ADD CONSTRAINT no_overlap_per_doctor
    EXCLUDE USING gist (
        doctor_id WITH =,
        tstzrange(start_time, end_time) WITH &&
    )
    WHERE (status != 'cancelled');

CREATE TYPE appointment_status AS ENUM ('pending', 'confirmed', 'cancelled', 'expired');

CREATE TABLE IF NOT EXISTS appointments (
    id          UUID               PRIMARY KEY DEFAULT gen_random_uuid(),
    patient_id  UUID               NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    slot_id     UUID               NOT NULL REFERENCES time_slots(id) ON DELETE CASCADE,
    status      appointment_status NOT NULL DEFAULT 'pending',
    notes       TEXT               NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ        NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ        NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_appt_patient_id ON appointments(patient_id);
CREATE INDEX IF NOT EXISTS idx_appt_slot_id    ON appointments(slot_id);
CREATE INDEX IF NOT EXISTS idx_appt_status     ON appointments(status);
CREATE INDEX IF NOT EXISTS idx_appt_created_at ON appointments(created_at);
