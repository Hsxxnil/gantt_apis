alter table projects
    add column manager_id uuid references resources (resource_uuid);