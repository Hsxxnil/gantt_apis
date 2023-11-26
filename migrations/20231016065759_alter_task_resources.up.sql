create index idx_task_resources_id
    on task_resources using hash (id);

create index idx_task_resources_task_uuid
    on task_resources using hash (task_uuid);

create index idx_task_resources_project_resource_uuid
    on task_resources using hash (project_resource_uuid);

create index idx_task_resources_created_at
    on task_resources (created_at desc);

create index idx_task_resources_created_by
    on task_resources using hash (created_by);

create index idx_task_resources_updated_at
    on task_resources (updated_at desc);

create index idx_task_resources_updated_by
    on task_resources using hash (updated_by);