drop index idx_project_resources_is_editable;

alter table project_resources
    drop column is_editable;

