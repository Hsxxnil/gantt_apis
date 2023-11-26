create index idx_resources_resource_uuid
    on resources using hash (resource_uuid);

create index idx_resources_resource_name
    on resources (resource_name);

create index idx_resources_resource_group
    on resources (resource_group);

create index idx_resources_email
    on resources (email);

create index idx_resources_phone
    on resources (phone);

create index idx_resources_created_at
    on resources (created_at desc);

create index idx_resources_created_by
    on resources using hash (created_by);

create index idx_resources_updated_at
    on resources (updated_at desc);

create index idx_resources_updated_by
    on resources using hash (updated_by);