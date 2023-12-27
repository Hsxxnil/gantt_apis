alter table project_resources
    drop constraint project_resources_resource_uuid_fkey;

alter table project_resources
    drop constraint project_resources_project_uuid_fkey;