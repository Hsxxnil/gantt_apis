drop index idx_users_company_id;

alter table users
    drop column company_id;