drop index idx_users_phone_number;

alter table users
    drop column phone_number;

alter table users
    add is_enabled boolean not null default false;

create index idx_users_is_enabled on users (is_enabled);

alter table users
    add is_authenticator boolean not null default false;

create index idx_users_is_authenticator on users (is_authenticator);
