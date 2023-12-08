create table companies
(
    id            UUID NOT NULL PRIMARY KEY,
    domain        text not null,
    name          text not null,
    tax_id_number text,
    address       text,
    phone         text,
    remarks       text,
    created_at    TIMESTAMP default now(),
    created_by    UUID,
    updated_at    TIMESTAMP,
    updated_by    UUID,
    deleted_at    TIMESTAMP
);

create index idx_companies_id
    on companies using hash (id);

create index idx_companies_domain
    on companies (domain);

create index idx_companies_name
    on companies using gin (name gin_trgm_ops);

create index idx_companies_tax_id_number
    on companies (tax_id_number);

create index idx_companies_address
    on companies using gin (address gin_trgm_ops);

create index idx_companies_phone
    on companies using gin (phone gin_trgm_ops);

create index idx_companies_remarks
    on companies using gin (remarks gin_trgm_ops);

create index idx_companies_created_at
    on companies (created_at desc);

create index idx_companies_created_by
    on companies using hash (created_by);

create index idx_companies_updated_at
    on companies (updated_at desc);

create index idx_companies_updated_by
    on companies using hash (updated_by);