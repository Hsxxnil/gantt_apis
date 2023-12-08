drop index idx_tasks_baseline_duration;

alter table tasks
    drop column baseline_duration;