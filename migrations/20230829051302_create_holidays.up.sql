CREATE TABLE holidays
(
    id         UUID    NOT NULL PRIMARY KEY,
    name       text NOT NULL,
    start_date TIMESTAMP,
    end_date   TIMESTAMP,
    css        text,
    created_at TIMESTAMP default now(),
    created_by UUID,
    updated_at TIMESTAMP,
    updated_by UUID,
    deleted_at TIMESTAMP
);