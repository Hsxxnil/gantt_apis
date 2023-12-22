create table departments
(
    id            UUID NOT NULL PRIMARY KEY,
    name          text not null,
    supervisor_id UUID references users (id),
    fax           text,
    tel           text,
    created_at    TIMESTAMP default now(),
    created_by    UUID,
    updated_at    TIMESTAMP,
    updated_by    UUID,
    deleted_at    TIMESTAMP
);

create index idx_departments_id
    on departments using hash (id);

create index idx_departments_name
    on departments using gin (name gin_trgm_ops);

create index idx_departments_supervisor_id
    on departments using hash (supervisor_id);

create index idx_departments_fax
    on departments using gin (fax gin_trgm_ops);

create index idx_departments_tel
    on departments using gin (tel gin_trgm_ops);

create index idx_departments_created_at
    on departments (created_at desc);

create index idx_departments_created_by
    on departments using hash (created_by);

create index idx_departments_updated_at
    on departments (updated_at desc);

create index idx_departments_updated_by
    on departments using hash (updated_by);
