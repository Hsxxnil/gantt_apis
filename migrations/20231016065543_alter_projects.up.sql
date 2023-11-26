create index idx_projects_project_uuid
    on projects using hash (project_uuid);

create index idx_projects_project_name
    on projects (project_name);

create index idx_projects_start_date
    on projects (start_date);

create index idx_projects_end_date
    on projects (end_date);

create index idx_projects_created_at
    on projects (created_at desc);

create index idx_projects_created_by
    on projects using hash (created_by);

create index idx_projects_updated_at
    on projects (updated_at desc);

create index idx_projects_updated_by
    on projects using hash (updated_by);

create index idx_projects_type
    on projects using hash (type);

create index idx_projects_code
    on projects (code);

create index idx_projects_manager
    on projects using hash (manager);

create index idx_projects_client
    on projects (client);

create index idx_projects_status
    on projects (status);