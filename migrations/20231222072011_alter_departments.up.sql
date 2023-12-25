drop index idx_departments_supervisor_id;

alter table departments
    drop column supervisor_id;