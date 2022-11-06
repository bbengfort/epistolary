-- Populate the database with initial roles and permissions data
BEGIN;

INSERT INTO roles (id, title, description) VALUES
    (1, 'Admin', 'Can create and manage any epistles in the database'),
    (2, 'Member', 'Can create and manage their own epistles'),
    (3, 'Observer', 'Only has read only access to epistles they created')
;

INSERT INTO permissions (id, title, description) VALUES
    (1, 'admin:cms', 'Can access the CMS site and manage the epistles database'),
    (2, 'epistles:read', 'Can view the epistles for the logged in user'),
    (3, 'epistles:create', 'Can create new epistles for the logged in user'),
    (4, 'epistles:delete', 'Can delete an epistle for th elogged in user')
;

INSERT INTO role_permissions (role_id, permission_id) VALUES
    (1, 1),
    (1, 2),
    (1, 3),
    (1, 4),
    (2, 2),
    (2, 3),
    (2, 4),
    (3, 2)
;

COMMIT;