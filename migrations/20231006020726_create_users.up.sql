create table users
(
    id            uuid      default uuid_generate_v4() not null
        primary key,
    resource_uuid uuid references resources (resource_uuid),
    password      text      default ''::text           not null,
    user_name     text      default ''::text           not null,
    name          text      default ''::text           not null,
    phone_number  text,
    email         text,
    role_id       uuid                                 not null references roles (id),
    created_at    timestamp default now(),
    created_by    uuid,
    updated_at    timestamp,
    updated_by    uuid,
    deleted_at    timestamp
);

create index idx_users_id
    on users using hash (id);

create index idx_users_user_name
    on users using gin (user_name gin_trgm_ops);

create index idx_users_name
    on users using gin (name gin_trgm_ops);

create index idx_users_resource_uuid
    on users using hash (resource_uuid);

create index idx_users_phone_number
    on users using gin (phone_number gin_trgm_ops);

create index idx_users_email
    on users using gin (email gin_trgm_ops);

create index idx_users_role_id
    on users using hash (role_id);

create index idx_users_created_at
    on users (created_at desc);

create index idx_users_created_by
    on users using hash (created_by);

create index idx_users_updated_at
    on users (updated_at desc);

create index idx_users_updated_by
    on users using hash (updated_by);

insert into users(id, user_name, name, password, role_id)
values ('a1bb0141-68e3-420c-8a92-9332fc21bd25', 'admin', '管理員',
        '9HXSglPqDWrOyA29croTTu8O8ahmj2EMHhxrsfzrEpJBVykaIkDJ211tJ03aq25Q2iHvkECACPDI/yJXiDsRQDojG1iLqTMQp3nUSmfV/9Yhc3i+ovXLuiRoapCluqw4oxkiuLtqlQMivNTnphmOF+iHnu6sz8N6aouA3mOS89aSoPpHwbWbo4ilh3sPIyEnwLT9npq3ICQwP7FxXPFxaw==',
        'd56fc184-9441-4396-be6c-d48580650171')