create index idx_tasks_task_uuid
    on tasks using hash (task_uuid);

create index idx_tasks_task_name
    on tasks (task_name);

create index idx_tasks_start_date
    on tasks (start_date);

create index idx_tasks_end_date
    on tasks (end_date);

create index idx_tasks_baseline_start_date
    on tasks (baseline_start_date);

create index idx_tasks_baseline_end_date
    on tasks (baseline_end_date);

create index idx_tasks_coordinator
    on tasks using hash (coordinator);

create index idx_tasks_outline_number
    on tasks (outline_number);

create index idx_tasks_assignments
    on tasks (assignments);

create index idx_tasks_task_color
    on tasks (task_color);

create index idx_tasks_web_link
    on tasks (web_link);

create index idx_tasks_created_at
    on tasks (created_at desc);

create index idx_tasks_created_by
    on tasks using hash (created_by);

create index idx_tasks_updated_at
    on tasks (updated_at desc);

create index idx_tasks_updated_by
    on tasks using hash (updated_by);

create index idx_tasks_project_uuid
    on tasks using hash (project_uuid);

create index idx_tasks_segment
    on tasks (segment);

create index idx_tasks_indicator
    on tasks (indicator);

create index idx_tasks_remark
    on tasks using gin(remark gin_trgm_ops);