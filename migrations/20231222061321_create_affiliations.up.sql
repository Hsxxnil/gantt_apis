create table affiliations
(
    id            UUID    NOT NULL PRIMARY KEY,
    user_id       UUID    not null references users (id),
    dept_id       UUID    not null references departments (id),
    job_title     text,
    is_supervisor boolean not null default false,
    created_at    TIMESTAMP        default now(),
    created_by    UUID,
    updated_at    TIMESTAMP,
    updated_by    UUID,
    deleted_at    TIMESTAMP
);

create index idx_affiliations_id
    on affiliations using hash (id);

create index idx_affiliations_user_id
    on affiliations using hash (user_id);

create index idx_affiliations_dept_id
    on affiliations using hash (dept_id);

create index idx_affiliations_job_title
    on affiliations using gin (job_title gin_trgm_ops);

create index idx_affiliations_is_supervisor
    on affiliations (is_supervisor);

create index idx_affiliations_created_at
    on affiliations (created_at desc);

create index idx_affiliations_created_by
    on affiliations using hash (created_by);

create index idx_affiliations_updated_at
    on affiliations (updated_at desc);

create index idx_affiliations_updated_by
    on affiliations using hash (updated_by);
