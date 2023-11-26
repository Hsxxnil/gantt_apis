alter table tasks
alter column duration type integer using duration::integer;