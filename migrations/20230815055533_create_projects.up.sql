CREATE TABLE projects
(
    project_uuid UUID    NOT NULL PRIMARY KEY,
    project_id   SERIAL,
    project_name text NOT NULL,
    created_at   TIMESTAMP default now(),
    created_by   UUID,
    updated_at   TIMESTAMP,
    updated_by   UUID,
    deleted_at   TIMESTAMP
);