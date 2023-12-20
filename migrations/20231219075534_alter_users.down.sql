alter table users
    add column phone_number text;

create index idx_users_phone_number
    on users using gin (phone_number gin_trgm_ops);

alter table users
    drop is_enabled;

drop index idx_users_is_enabled;

alter table users
    drop is_authenticator;

drop index idx_users_is_authenticator;
