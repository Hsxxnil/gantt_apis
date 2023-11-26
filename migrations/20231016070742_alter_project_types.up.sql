create index idx_project_types_id
    on project_types using hash (id);

create index idx_project_types_name
    on project_types (name);

create index idx_project_types_created_at
    on project_types (created_at desc);

create index idx_project_types_created_by
    on project_types using hash (created_by);

create index idx_project_types_updated_at
    on project_types (updated_at desc);

create index idx_project_types_updated_by
    on project_types using hash (updated_by);