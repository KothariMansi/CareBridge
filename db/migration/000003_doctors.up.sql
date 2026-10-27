CREATE TABLE doctors (
    id               uuid        PRIMARY KEY DEFAULT uuidv7(),
    full_name        text        NOT NULL,
    biography        text,
    qualifications   text,
    experience_years integer     NOT NULL DEFAULT 0,
    status           text        NOT NULL DEFAULT 'active',
    created_at       timestamptz NOT NULL DEFAULT now(),
    updated_at       timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT doctors_full_name_not_blank  CHECK (btrim(full_name) <> ''),
    CONSTRAINT doctors_experience_range     CHECK (experience_years BETWEEN 0 AND 80),
    CONSTRAINT doctors_status_valid         CHECK (status IN ('active', 'inactive'))
);

-- A doctor can practise at several hospitals, with a different fee at each.
-- Slots and appointments will hang off this table, not off doctors.
CREATE TABLE doctor_hospitals (
    id                     uuid        PRIMARY KEY DEFAULT uuidv7(),
    doctor_id              uuid        NOT NULL,
    hospital_id            uuid        NOT NULL,
    consultation_fee_minor integer     NOT NULL,           -- paise, never floats for money
    currency               text        NOT NULL DEFAULT 'INR',
    status                 text        NOT NULL DEFAULT 'active',
    created_at             timestamptz NOT NULL DEFAULT now(),
    updated_at             timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT doctor_hospitals_doctor_id_fkey   FOREIGN KEY (doctor_id)   REFERENCES doctors (id),
    CONSTRAINT doctor_hospitals_hospital_id_fkey FOREIGN KEY (hospital_id) REFERENCES hospitals (id),
    CONSTRAINT doctor_hospitals_pair_key         UNIQUE (doctor_id, hospital_id),
    CONSTRAINT doctor_hospitals_fee_not_negative CHECK (consultation_fee_minor >= 0),
    CONSTRAINT doctor_hospitals_currency_format  CHECK (currency ~ '^[A-Z]{3}$'),
    CONSTRAINT doctor_hospitals_status_valid     CHECK (status IN ('active', 'inactive'))
);

-- "Which doctors work at this hospital?" The (doctor_id, hospital_id) unique
-- index already serves lookups by doctor.
CREATE INDEX doctor_hospitals_hospital_id_idx ON doctor_hospitals (hospital_id);

CREATE TABLE doctor_specialities (
    doctor_id     uuid NOT NULL,
    speciality_id uuid NOT NULL,

    CONSTRAINT doctor_specialities_pkey               PRIMARY KEY (doctor_id, speciality_id),
    CONSTRAINT doctor_specialities_doctor_id_fkey     FOREIGN KEY (doctor_id)     REFERENCES doctors (id),
    CONSTRAINT doctor_specialities_speciality_id_fkey FOREIGN KEY (speciality_id) REFERENCES specialities (id)
);

-- "Which doctors have this speciality?" The primary key serves the other direction.
CREATE INDEX doctor_specialities_speciality_id_idx ON doctor_specialities (speciality_id);