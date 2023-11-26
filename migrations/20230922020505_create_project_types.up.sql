CREATE TABLE project_types
(
    id           UUID NOT NULL PRIMARY KEY,
    name         VARCHAR,
    created_at   TIMESTAMP default now(),
    created_by   UUID,
    updated_at   TIMESTAMP,
    updated_by   UUID,
    deleted_at   TIMESTAMP
);