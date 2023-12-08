alter table tasks
    add column baseline_duration numeric;

create index idx_tasks_baseline_duration
    on tasks (baseline_duration);