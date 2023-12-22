alter table departments
    add column supervisor_id UUID references users (id);

create index idx_departments_supervisor_id
    on departments using hash (supervisor_id);