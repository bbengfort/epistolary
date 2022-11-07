/*
 * Initial database schema for the Epistolary API server.
 */
BEGIN;

/*
 * Enumerations and custom data types.
 */

CREATE TYPE reading_status AS ENUM (
    'queued',
    'started',
    'finished',
    'archived'
);

/*
 * Table definitions including PRIMARY KEY and UNIQUE indices
 */

-- Epistles is the collection of letters that we're saving for further reading
CREATE TABLE IF NOT EXISTS epistles (
    id          SERIAL PRIMARY KEY,
    link        VARCHAR(2000) UNIQUE NOT NULL,
    title       VARCHAR(512),
    description VARCHAR(4096),
    favicon     VARCHAR(2000),
    created     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    modified    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Maps the epistles to the user that created them and information about reading
CREATE TABLE IF NOT EXISTS reading (
    epistle_id    INTEGER,
    user_id       INTEGER,
    status        reading_status DEFAULT 'queued',
    started       TIMESTAMPTZ,
    finished      TIMESTAMPTZ,
    archived      TIMESTAMPTZ,
    created       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    modified      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (epistle_id, user_id)
);

-- Primary authentication table that holds usernames and hashed passwords
CREATE TABLE IF NOT EXISTS users (
    id          SERIAL PRIMARY KEY,
    email       VARCHAR(255) UNIQUE NOT NULL,
    username    VARCHAR(255) UNIQUE NOT NULL,
    password    VARCHAR(255) NOT NULL,
    role_id     INTEGER NOT NULL,
    last_seen   TIMESTAMPTZ,
    created     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    modified    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Roles are collections of permissions that can be quickly assigned to a user
CREATE TABLE IF NOT EXISTS roles (
    id          SERIAL PRIMARY KEY,
    title       VARCHAR(255) UNIQUE NOT NULL,
    description VARCHAR(512),
    created     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    modified    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Permissions (or scopes) authorize the user to perform actions on the API
CREATE TABLE IF NOT EXISTS permissions (
    id          SERIAL PRIMARY KEY,
    title       VARCHAR(255) UNIQUE NOT NULL,
    description VARCHAR(512),
    created     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    modified    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Maps the default permissions to a role
CREATE TABLE IF NOT EXISTS role_permissions (
    role_id       INTEGER,
    permission_id INTEGER,
    created       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    modified      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (role_id, permission_id)
);

/*
 * Foreign Key Relationships
 */

ALTER TABLE reading ADD CONSTRAINT fk_reading_epistle
    FOREIGN KEY (epistle_id) REFERENCES epistles (id)
    ON DELETE CASCADE;

ALTER TABLE reading ADD CONSTRAINT fk_reading_user
    FOREIGN KEY (user_id) REFERENCES users (id)
    ON DELETE CASCADE;

ALTER TABLE users ADD CONSTRAINT fk_users_role
    FOREIGN KEY (role_id) REFERENCES roles (id)
    ON DELETE RESTRICT;

ALTER TABLE role_permissions ADD CONSTRAINT fk_role_permissions_role
    FOREIGN KEY (role_id) REFERENCES roles (id)
    ON DELETE CASCADE;

ALTER TABLE role_permissions ADD CONSTRAINT fk_role_permissions_permission
    FOREIGN KEY (permission_id) REFERENCES permissions (id)
    ON DELETE CASCADE;

/*
 * Views
 */

-- Allows the easy selection of all permissions for a user based on their role
CREATE OR REPLACE VIEW user_permissions AS
    SELECT u.id as user_id, p.title as permission
        FROM users u
        JOIN role_permissions rp ON rp.role_id = u.role_id
        JOIN permissions p ON p.id = rp.permission_id
;

/*
 * Automatically update modified timestamps
 */

CREATE OR REPLACE FUNCTION trigger_set_modified_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.modified = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Epistles modified timestamp
CREATE TRIGGER set_epistles_modified
BEFORE UPDATE ON epistles
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_modified_timestamp();

-- Epistles User modified timestamp
CREATE TRIGGER set_reading_modified
BEFORE UPDATE ON reading
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_modified_timestamp();

-- Users modified timestamp
CREATE TRIGGER set_users_modified
BEFORE UPDATE ON users
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_modified_timestamp();

-- Roles modified timestamp
CREATE TRIGGER set_roles_modified
BEFORE UPDATE ON roles
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_modified_timestamp();

-- Permissions modified timestamp
CREATE TRIGGER set_permissions_modified
BEFORE UPDATE ON permissions
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_modified_timestamp();

-- Role Permissions modified timestamp
CREATE TRIGGER set_role_permissions_modified
BEFORE UPDATE ON role_permissions
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_modified_timestamp();

COMMIT;