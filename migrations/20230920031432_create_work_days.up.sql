CREATE TABLE work_days
(
    id           UUID NOT NULL PRIMARY KEY,
    work_week    text,
    working_time text,
    created_at   TIMESTAMP default now(),
    created_by   UUID,
    updated_at   TIMESTAMP,
    updated_by   UUID,
    deleted_at   TIMESTAMP
);