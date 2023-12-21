alter table projects
    rename column type_id to type;

alter table projects
    rename column manager_id to manager;

alter index idx_projects_type_id
    rename to idx_projects_type;

alter index idx_projects_manager_id
    rename to idx_projects_manager;