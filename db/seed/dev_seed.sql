-- Synthetic development data. Every name, address and number here is invented.
--
-- Safe to run any number of times: it only inserts, and rows that already exist
-- are skipped. It never updates or deletes anything.

BEGIN;

INSERT INTO specialities (name, slug, description) VALUES
    ('Cardiology',       'cardiology',       'Heart and blood vessels'),
    ('Dermatology',      'dermatology',      'Skin, hair and nails'),
    ('Orthopaedics',     'orthopaedics',     'Bones, joints and muscles'),
    ('Paediatrics',      'paediatrics',      'Health care for children'),
    ('General Medicine', 'general-medicine', 'First point of contact for adults'),
    ('Gynaecology',      'gynaecology',      'Women''s reproductive health')
ON CONFLICT DO NOTHING;

INSERT INTO hospitals (name, slug, address, city, phone) VALUES
    ('Sunrise Multispeciality Hospital', 'sunrise-multispeciality', '12 Park Road, Shivaji Nagar',   'Pune',   '+91 20 5550 0101'),
    ('Lakeview Medical Centre',          'lakeview-medical-centre',  '48 Lake Road, Powai',          'Mumbai', '+91 22 5550 0202'),
    ('Riverside Care Hospital',          'riverside-care',           '7 River Lane, Kothrud',        'Pune',   '+91 20 5550 0303')
ON CONFLICT DO NOTHING;

-- Fixed IDs make this script repeatable: a second run hits the same primary
-- keys and skips the rows.
INSERT INTO doctors (id, full_name, biography, qualifications, experience_years) VALUES
    ('00000000-0000-7000-8000-000000000001', 'Dr. Anita Deshmukh', 'Interventional cardiologist.',            'MBBS, MD (Medicine), DM (Cardiology)',  16),
    ('00000000-0000-7000-8000-000000000002', 'Dr. Rahul Verma',    'Skin and hair specialist.',               'MBBS, MD (Dermatology)',                9),
    ('00000000-0000-7000-8000-000000000003', 'Dr. Meera Iyer',     'Paediatrician with a focus on newborns.', 'MBBS, DCH, MD (Paediatrics)',           12),
    ('00000000-0000-7000-8000-000000000004', 'Dr. Sanjay Kulkarni','Joint replacement surgeon.',              'MBBS, MS (Orthopaedics)',               20),
    ('00000000-0000-7000-8000-000000000005', 'Dr. Pooja Nair',     'General physician.',                      'MBBS, MD (Medicine)',                   7),
    ('00000000-0000-7000-8000-000000000006', 'Dr. Imran Shaikh',   'Cardiologist, heart failure clinic.',     'MBBS, MD, DM (Cardiology)',             11),
    ('00000000-0000-7000-8000-000000000007', 'Dr. Kavita Joshi',   'Gynaecologist and obstetrician.',         'MBBS, MS (Obstetrics and Gynaecology)', 14),
    ('00000000-0000-7000-8000-000000000008', 'Dr. Vikram Rao',     'Sports injury specialist.',               'MBBS, MS (Orthopaedics)',               6)
ON CONFLICT DO NOTHING;

-- Where each doctor practises, and the fee in paise (50000 = INR 500.00).
INSERT INTO doctor_hospitals (doctor_id, hospital_id, consultation_fee_minor)
SELECT v.doctor_id::uuid, h.id, v.fee
FROM (VALUES
    ('00000000-0000-7000-8000-000000000001', 'sunrise-multispeciality', 80000),
    ('00000000-0000-7000-8000-000000000001', 'lakeview-medical-centre', 90000),
    ('00000000-0000-7000-8000-000000000002', 'sunrise-multispeciality', 60000),
    ('00000000-0000-7000-8000-000000000003', 'riverside-care',          50000),
    ('00000000-0000-7000-8000-000000000004', 'lakeview-medical-centre', 70000),
    ('00000000-0000-7000-8000-000000000005', 'riverside-care',          40000),
    ('00000000-0000-7000-8000-000000000005', 'sunrise-multispeciality', 45000),
    ('00000000-0000-7000-8000-000000000006', 'riverside-care',          75000),
    ('00000000-0000-7000-8000-000000000007', 'sunrise-multispeciality', 65000),
    ('00000000-0000-7000-8000-000000000008', 'riverside-care',          55000)
) AS v(doctor_id, hospital_slug, fee)
JOIN hospitals h ON h.slug = v.hospital_slug
ON CONFLICT DO NOTHING;

INSERT INTO doctor_specialities (doctor_id, speciality_id)
SELECT v.doctor_id::uuid, s.id
FROM (VALUES
    ('00000000-0000-7000-8000-000000000001', 'cardiology'),
    ('00000000-0000-7000-8000-000000000002', 'dermatology'),
    ('00000000-0000-7000-8000-000000000003', 'paediatrics'),
    ('00000000-0000-7000-8000-000000000004', 'orthopaedics'),
    ('00000000-0000-7000-8000-000000000005', 'general-medicine'),
    ('00000000-0000-7000-8000-000000000006', 'cardiology'),
    ('00000000-0000-7000-8000-000000000006', 'general-medicine'),
    ('00000000-0000-7000-8000-000000000007', 'gynaecology'),
    ('00000000-0000-7000-8000-000000000008', 'orthopaedics')
) AS v(doctor_id, speciality_slug)
JOIN specialities s ON s.slug = v.speciality_slug
ON CONFLICT DO NOTHING;

COMMIT;