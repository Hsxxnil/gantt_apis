alter table task_resources
    drop column project_resource_uuid;

alter table task_resources
    add resource_uuid uuid references resources (resource_uuid);