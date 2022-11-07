-- Allows a postgres server to be migrated with SQL only (bypassing go-migrate).
BEGIN;

CREATE TABLE IF NOT EXISTS schema_migrations (
    version BIGINT NOT NULL
        CONSTRAINT schema_migrations_pkey
            PRIMARY KEY,
    dirty BOOLEAN NOT NULL
);

INSERT INTO schema_migrations(version, dirty) VALUES (2, false);

COMMIT;