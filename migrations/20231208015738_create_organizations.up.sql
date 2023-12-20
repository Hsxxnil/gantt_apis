create table organizations
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

create index idx_organizations_id
    on organizations using hash (id);

create index idx_organizations_domain
    on organizations (domain);

create index idx_organizations_name
    on organizations using gin (name gin_trgm_ops);

create index idx_organizations_tax_id_number
    on organizations (tax_id_number);

create index idx_organizations_address
    on organizations using gin (address gin_trgm_ops);

create index idx_organizations_phone
    on organizations using gin (phone gin_trgm_ops);

create index idx_organizations_remarks
    on organizations using gin (remarks gin_trgm_ops);

create index idx_organizations_created_at
    on organizations (created_at desc);

create index idx_organizations_created_by
    on organizations using hash (created_by);

create index idx_organizations_updated_at
    on organizations (updated_at desc);

create index idx_organizations_updated_by
    on organizations using hash (updated_by);