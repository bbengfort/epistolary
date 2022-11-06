BEGIN;

-- drop views first to ensure tables can be dropped
DROP VIEW IF EXISTS user_permissions;

-- drop tables in order of foreign key dependencies
DROP TABLE IF EXISTS role_permissions;
DROP TABLE IF EXISTS epistles_user;
DROP TABLE IF EXISTS permissions;
DROP TABLE IF EXISTS roles;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS epistles;

-- drop triggers and functions last
DROP FUNCTION trigger_set_modified_timestamp;

COMMIT;