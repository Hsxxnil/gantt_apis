alter table project_resources
    add column is_editable boolean not null default true;

create index idx_project_resources_is_editable on project_resources (is_editable);