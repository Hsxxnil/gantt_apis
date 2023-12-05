alter table tasks
    drop constraint tasks_coordinator_fkey;

alter table tasks
    add foreign key (coordinator) references project_resources (id);

