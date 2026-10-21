CREATE TABLE specialities (
    id          uuid        PRIMARY KEY DEFAULT uuidv7(),
    name        text        NOT NULL,
    slug        text        NOT NULL,
    description text,
    created_at  timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT specialities_slug_key       UNIQUE (slug),
    CONSTRAINT specialities_name_not_blank CHECK (btrim(name) <> ''),
    CONSTRAINT specialities_slug_format    CHECK (slug ~ '^[a-z0-9]+(-[a-z0-9]+)*$')
);

-- "Cardiology" and "cardiology" must not both exist.
CREATE UNIQUE INDEX specialities_name_lower_key ON specialities (lower(name));