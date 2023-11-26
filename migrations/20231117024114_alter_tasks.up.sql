alter table tasks
    rename column task_id to task_old_id;

alter table tasks
    add column task_id text;