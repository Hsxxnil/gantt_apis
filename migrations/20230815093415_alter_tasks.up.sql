alter table tasks
    add project_uuid UUID REFERENCES projects(project_uuid);