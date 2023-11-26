alter table tasks
alter column duration type numeric using duration::numeric;