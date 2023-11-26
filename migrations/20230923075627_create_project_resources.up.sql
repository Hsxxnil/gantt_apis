CREATE TABLE project_resources
(
    id            UUID    NOT NULL PRIMARY KEY,
    resource_uuid UUID    NOT NULL,
    project_uuid  UUID    NOT NULL,
    role          VARCHAR not null,
    created_at    TIMESTAMP,
    created_by    UUID,
    updated_at    TIMESTAMP,
    updated_by    UUID,
    deleted_at    TIMESTAMP
);



