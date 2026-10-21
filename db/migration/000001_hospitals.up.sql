CREATE TABLE hospitals (
    id         uuid        PRIMARY KEY DEFAULT uuidv7(),
    name       text        NOT NULL,
    slug       text        NOT NULL,
    address    text        NOT NULL,
    city       text        NOT NULL,
    timezone   text        NOT NULL DEFAULT 'Asia/Kolkata', -- IANA name, validated in Go
    phone      text,
    status     text        NOT NULL DEFAULT 'active',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT hospitals_slug_key       UNIQUE (slug),
    CONSTRAINT hospitals_name_not_blank CHECK (btrim(name) <> ''),
    CONSTRAINT hospitals_slug_format    CHECK (slug ~ '^[a-z0-9]+(-[a-z0-9]+)*$'),
    CONSTRAINT hospitals_status_valid   CHECK (status IN ('active', 'inactive'))
);

CREATE INDEX hospitals_city_idx ON hospitals (city);