alter table tasks
    drop column task_id;

alter table tasks
    rename column task_old_id to task_id;