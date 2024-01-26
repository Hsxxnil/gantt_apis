create table s3_files
(
    id             UUID NOT NULL PRIMARY KEY,
    file_url       text not null,
    file_name      text not null,
    file_extension text not null,
    source_uuid    UUID not null,
    created_at     TIMESTAMP default now(),
    created_by     UUID,
    deleted_at     TIMESTAMP
);

create index idx_s3_files_id
    on s3_files using hash (id);

create index idx_s3_files_file_url
    on s3_files (file_url);

create index idx_s3_files_file_name
    on s3_files using gin (file_name gin_trgm_ops);

create index idx_s3_files_file_extension
    on s3_files (file_extension);

create index idx_s3_files_source_uuid
    on s3_files using hash (source_uuid);

create index idx_s3_files_created_at
    on s3_files (created_at desc);

create index idx_s3_files_created_by
    on s3_files using hash (created_by);
