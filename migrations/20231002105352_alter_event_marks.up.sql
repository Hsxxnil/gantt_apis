alter table event_marks
    add project_uuid UUID REFERENCES projects(project_uuid);