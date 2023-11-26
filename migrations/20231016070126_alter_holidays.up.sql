create index idx_holidays_id
    on holidays using hash (id);

create index idx_holidays_name
    on holidays (name);

create index idx_holidays_start_date
    on holidays (start_date);

create index idx_holidays_end_date
    on holidays (end_date);

create index idx_holidays_created_at
    on holidays (created_at desc);

create index idx_holidays_created_by
    on holidays using hash (created_by);

create index idx_holidays_updated_at
    on holidays (updated_at desc);

create index idx_holidays_updated_by
    on holidays using hash (updated_by);