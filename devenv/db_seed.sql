-- SCHEMA: public
CREATE SCHEMA IF NOT EXISTS eth
    AUTHORIZATION postgres;

GRANT ALL ON SCHEMA eth TO PUBLIC;

GRANT ALL ON SCHEMA eth TO postgres;

-- Table: public.block
CREATE TABLE IF NOT EXISTS eth.blocks
(
    block_num BIGINT PRIMARY KEY,
    block_hash   VARCHAR(255) UNIQUE NOT NULL,
    block_time   BIGINT,
    parent_hash VARCHAR(255),
    
    stable BOOL,
    created_at BIGINT DEFAULT (date_part('epoch'::text, now()) * (1000)::double precision),
    updated_at BIGINT DEFAULT (date_part('epoch'::text, now()) * (1000)::double precision)
);

-- TABLESPACE pg_default;
ALTER TABLE eth.blocks OWNER to postgres;
