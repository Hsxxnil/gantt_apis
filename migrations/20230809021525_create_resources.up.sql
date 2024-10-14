CREATE EXTENSION IF NOT EXISTS "pg_trgm";
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE resources
(
    resource_uuid UUID PRIMARY KEY,
    resource_id   SERIAL,
    resource_name text,
    role          text,
    email         text,
    phone         text,
    standard_cost NUMERIC,
    total_cost    NUMERIC,
    total_load    NUMERIC,
    unit          INTEGER,
    created_at    TIMESTAMP default now(),
    created_by    UUID,
    updated_at    TIMESTAMP,
    updated_by    UUID,
    deleted_at    TIMESTAMP
);

--csv
--id = resource_id
--name = resource_name