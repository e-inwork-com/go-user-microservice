CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    created_at_dt timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    email_t text UNIQUE NOT NULL,
    password_hash bytea NOT NULL,
    first_name_t char varying(100) NOT NULL,
    last_name_t char varying(100) NOT NULL,
    activated_b bool NOT NULL DEFAULT false,
    version integer NOT NULL DEFAULT 1
);