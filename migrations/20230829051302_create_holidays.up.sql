CREATE TABLE holidays
(
    id         UUID    NOT NULL PRIMARY KEY,
    name       VARCHAR NOT NULL,
    start_date TIMESTAMP,
    end_date   TIMESTAMP,
    css        VARCHAR,
    created_at TIMESTAMP default now(),
    created_by UUID,
    updated_at TIMESTAMP,
    updated_by UUID,
    deleted_at TIMESTAMP
);