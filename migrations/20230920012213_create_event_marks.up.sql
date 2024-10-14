CREATE TABLE event_marks
(
    id         UUID    NOT NULL PRIMARY KEY,
    name       text NOT NULL,
    day        TIMESTAMP,
    created_at TIMESTAMP default now(),
    created_by UUID,
    updated_at TIMESTAMP,
    updated_by UUID,
    deleted_at TIMESTAMP
);