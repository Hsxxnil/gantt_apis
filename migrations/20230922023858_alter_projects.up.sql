alter table projects
    add type uuid references project_types (id);

alter table projects
    add code varchar;

alter table projects
    add manager uuid references resources (resource_uuid);

alter table projects
    add start_date timestamp;

alter table projects
    add end_date timestamp;

alter table projects
    add client varchar;

alter table projects
    add status varchar;