alter table projects
    rename column type to type_id;

alter table projects
    rename column manager to manager_id;

alter index idx_projects_type
    rename to idx_projects_type_id;

alter index idx_projects_manager
    rename to idx_projects_manager_id;