alter table task_resources
    drop column resource_uuid;

alter table task_resources
    add project_resource_uuid uuid references project_resources (id);