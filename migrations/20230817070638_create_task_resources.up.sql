CREATE TABLE task_resources
(
    id                    UUID    NOT NULL PRIMARY KEY,
    project_resource_uuid UUID    NOT NULL,
    task_uuid             UUID    NOT NULL,
    unit                  numeric not null,
    created_at            TIMESTAMP,
    created_by            UUID,
    updated_at            TIMESTAMP,
    updated_by            UUID,
    deleted_at            TIMESTAMP
);



