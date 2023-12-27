alter table task_resources
    add constraint task_resources_task_uuid_fkey
        foreign key (task_uuid) references tasks (task_uuid);