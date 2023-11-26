create index idx_event_marks_id
    on event_marks using hash (id);

create index idx_event_marks_name
    on event_marks (name);

create index idx_event_marks_day
    on event_marks (day);

create index idx_event_marks_created_at
    on event_marks (created_at desc);

create index idx_event_marks_created_by
    on event_marks using hash (created_by);

create index idx_event_marks_updated_at
    on event_marks (updated_at desc);

create index idx_event_marks_updated_by
    on event_marks using hash (updated_by);

create index idx_event_marks_project_uuid
    on event_marks using hash (project_uuid);