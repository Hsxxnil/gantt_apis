alter table users
    add column org_id uuid references organizations (id);

create index idx_users_org_id
    on users using hash (org_id);