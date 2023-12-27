alter table project_resources
    add constraint project_resources_resource_uuid_fkey
        foreign key (resource_uuid) references resources (resource_uuid);

alter table project_resources
    add constraint project_resources_project_uuid_fkey
        foreign key (project_uuid) references projects (project_uuid);