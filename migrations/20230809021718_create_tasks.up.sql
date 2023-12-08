CREATE TABLE tasks
(
    task_uuid      UUID PRIMARY KEY,                          --pk
    task_id        SERIAL,
    task_name      VARCHAR NOT NULL,
    start_date     TIMESTAMP,                                 --起始日期
    end_date       TIMESTAMP,                                 --結束日期
    duration       INTEGER,                                   --期間
    progress       INTEGER,                                   --完成百分比
    cost           INTEGER,                                   --花費時間
    coordinator    UUID REFERENCES resources (resource_uuid), --fk with resources.resource_uuid 協調員
    predecessor    VARCHAR,                                   --前任
    outline_number VARCHAR,                                   -- 1.1.2、1.2、1.2.1
    assignments    VARCHAR,                                   -- 未知
    task_color     VARCHAR,                                   --紀錄標的顏色
    web_link       VARCHAR,                                   --預留：外部連結
    is_subtask     BOOL      DEFAULT FALSE,                   --是否為任務
    info           TEXT,
    created_at     TIMESTAMP default now(),
    created_by     UUID,
    updated_at     TIMESTAMP,
    updated_by     UUID,
    deleted_at     TIMESTAMP
);