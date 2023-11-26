create index idx_project_resources_id
    on project_resources using hash (id);

create index idx_project_resources_resource_uuid
    on project_resources using hash (resource_uuid);

create index idx_project_resources_project_uuid
    on project_resources using hash (project_uuid);

create index idx_project_resources_role
    on project_resources (role);

create index idx_project_resources_created_at
    on project_resources (created_at desc);

create index idx_project_resources_created_by
    on project_resources using hash (created_by);

create index idx_project_resources_updated_at
    on project_resources (updated_at desc);

create index idx_project_resources_updated_by
    on project_resources using hash (updated_by);