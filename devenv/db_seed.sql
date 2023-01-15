-- SCHEMA: public
CREATE SCHEMA IF NOT EXISTS eth
    AUTHORIZATION postgres;

GRANT ALL ON SCHEMA eth TO PUBLIC;

GRANT ALL ON SCHEMA eth TO postgres;

-- Table: eth.block
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

-- Table: eth.transactions
CREATE TABLE IF NOT EXISTS eth.transactions
(
    block_hash VARCHAR(255) REFERENCES eth.blocks (block_hash) ON DELETE CASCADE,
    tx_hash    VARCHAR(255) UNIQUE NOT NULL,
    tx_from    VARCHAR(255),
    tx_to      VARCHAR(255),
    nonce     BIGINT,
    tx_data    bytea,
    tx_value   VARCHAR(255),

    created_at BIGINT DEFAULT (date_part('epoch'::text, now()) * (1000)::double precision),
    updated_at BIGINT DEFAULT (date_part('epoch'::text, now()) * (1000)::double precision)
);

-- Table: eth.receipts
CREATE TABLE IF NOT EXISTS eth.receipts
(
    tx_hash   VARCHAR(255)  UNIQUE NOT NULL REFERENCES eth.transactions (tx_hash) ON DELETE CASCADE,

    created_at BIGINT DEFAULT (date_part('epoch'::text, now()) * (1000)::double precision),
    updated_at BIGINT DEFAULT (date_part('epoch'::text, now()) * (1000)::double precision)
);

-- Table: eth.transaction_logs
CREATE TABLE IF NOT EXISTS eth.transaction_logs
(
    tx_hash   VARCHAR(255) NOT NULL REFERENCES eth.receipts (tx_hash) ON DELETE CASCADE,
    log_index BIGINT,
    log_data   bytea
);


ALTER TABLE eth.blocks OWNER to postgres;
ALTER TABLE eth.transactions OWNER to postgres;
ALTER TABLE eth.receipts OWNER to postgres;
ALTER TABLE eth.transaction_logs OWNER to postgres;
