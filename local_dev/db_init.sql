-- SCHEMA: public
CREATE SCHEMA IF NOT EXISTS public
    AUTHORIZATION postgres;

GRANT ALL ON SCHEMA public TO PUBLIC;

GRANT ALL ON SCHEMA public TO postgres;

-- Table: public.block
CREATE TABLE IF NOT EXISTS public.blocks
(
    number BIGINT PRIMARY KEY,
    hash   VARCHAR(255) UNIQUE NOT NULL,
    time   BIGINT,
    parent VARCHAR(255),
    
    stable BOOL,
    created_at BIGINT DEFAULT (date_part('epoch'::text, now()) * (1000)::double precision),
    updated_at BIGINT DEFAULT (date_part('epoch'::text, now()) * (1000)::double precision)
);

-- Table: public.transactions
CREATE TABLE IF NOT EXISTS public.transactions
(
    id         INT          PRIMARY KEY,
    block_hash VARCHAR(255) REFERENCES blocks (hash) ON DELETE CASCADE,
    tx_hash    VARCHAR(255) UNIQUE NOT NULL,
    tx_from    VARCHAR(255),
    tx_to      VARCHAR(255),
    nounce     BIGINT,
    data       bytea,
    value      VARCHAR(255),
    
    created_at BIGINT DEFAULT (date_part('epoch'::text, now()) * (1000)::double precision),
    updated_at BIGINT DEFAULT (date_part('epoch'::text, now()) * (1000)::double precision)
);

-- Table: public.transaction_logs
CREATE TABLE IF NOT EXISTS public.transaction_logs
(
    id        INT          PRIMARY KEY,
    tx_hash   VARCHAR(255) REFERENCES transactions (tx_hash) ON DELETE CASCADE,
    log_index INT,
    data      bytea,
    
    created_at BIGINT DEFAULT (date_part('epoch'::text, now()) * (1000)::double precision),
    updated_at BIGINT DEFAULT (date_part('epoch'::text, now()) * (1000)::double precision)
);

-- TABLESPACE pg_default;
ALTER TABLE public.blocks OWNER to postgres;
ALTER TABLE public.transactions OWNER to postgres;
ALTER TABLE public.transaction_logs OWNER to postgres;