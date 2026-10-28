-- name: ListDoctors :many
SELECT d.id,
       d.full_name,
       d.biography,
       d.qualifications,
       d.experience_years
FROM doctors d
WHERE d.status = 'active'
  AND (sqlc.narg('q')::text IS NULL OR d.full_name ILIKE '%' || sqlc.narg('q')::text || '%')
  AND (sqlc.narg('speciality_id')::uuid IS NULL OR EXISTS (
        SELECT 1
        FROM doctor_specialities ds
        WHERE ds.doctor_id = d.id
          AND ds.speciality_id = sqlc.narg('speciality_id')::uuid))
  AND EXISTS (
        SELECT 1
        FROM doctor_hospitals dh
        JOIN hospitals h ON h.id = dh.hospital_id
        WHERE dh.doctor_id = d.id
          AND dh.status = 'active'
          AND h.status = 'active'
          AND (sqlc.narg('hospital_id')::uuid IS NULL OR h.id = sqlc.narg('hospital_id')::uuid)
          AND (sqlc.narg('city')::text IS NULL OR lower(h.city) = lower(sqlc.narg('city')::text)))
ORDER BY d.full_name, d.id
LIMIT sqlc.arg('page_limit')::int
OFFSET sqlc.arg('page_offset')::int;

-- name: CountDoctors :one
SELECT count(*)
FROM doctors d
WHERE d.status = 'active'
  AND (sqlc.narg('q')::text IS NULL OR d.full_name ILIKE '%' || sqlc.narg('q')::text || '%')
  AND (sqlc.narg('speciality_id')::uuid IS NULL OR EXISTS (
        SELECT 1
        FROM doctor_specialities ds
        WHERE ds.doctor_id = d.id
          AND ds.speciality_id = sqlc.narg('speciality_id')::uuid))
  AND EXISTS (
        SELECT 1
        FROM doctor_hospitals dh
        JOIN hospitals h ON h.id = dh.hospital_id
        WHERE dh.doctor_id = d.id
          AND dh.status = 'active'
          AND h.status = 'active'
          AND (sqlc.narg('hospital_id')::uuid IS NULL OR h.id = sqlc.narg('hospital_id')::uuid)
          AND (sqlc.narg('city')::text IS NULL OR lower(h.city) = lower(sqlc.narg('city')::text)));

-- name: ListDoctorSpecialities :many
SELECT ds.doctor_id,
    s.id AS speciality_id,
    s.name,
    s.slug
FROM doctor_specialities ds
JOIN specialities s ON s.id = ds.speciality_id
WHERE ds.doctor_id = ANY(sqlc.arg('doctor_ids')::uuid[])
ORDER BY s.name;

-- name: ListDoctorHospitals :many
SELECT dh.doctor_id,
       h.id AS hospital_id,
       h.name,
       h.city,
       dh.consultation_fee_minor,
       dh.currency
FROM doctor_hospitals dh
JOIN hospitals h ON h.id = dh.hospital_id
WHERE dh.doctor_id = ANY(sqlc.arg('doctor_ids')::uuid[])
  AND dh.status = 'active'
  AND h.status = 'active'
ORDER BY h.name;