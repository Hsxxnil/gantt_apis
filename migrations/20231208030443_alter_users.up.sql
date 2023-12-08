alter table users
    add column company_id uuid references companies (id);

create index idx_users_company_id
    on users using hash (company_id);