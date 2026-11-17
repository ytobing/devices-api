-- +goose Up
CREATE TABLE devices (
    id         TEXT PRIMARY KEY,
    name       TEXT NOT NULL,
    brand      TEXT NOT NULL,
    state      TEXT NOT NULL DEFAULT 'AVAILABLE'
               CHECK (state IN ('AVAILABLE', 'IN_USE', 'INACTIVE')),
    created_at TEXT NOT NULL
);

-- +goose Down
DROP TABLE devices;
