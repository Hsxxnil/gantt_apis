alter table tasks
    add column coordinator uuid references resources (resource_uuid);