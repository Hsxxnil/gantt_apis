drop index idx_users_org_id;

alter table users
    drop column org_id;