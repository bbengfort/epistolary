package users

import (
	"context"
	"database/sql"
	"time"

	"github.com/bbengfort/epistolary/pkg/server/db"
)

const DefaultRoleID int64 = 2

// Database model for a role object.
type Role struct {
	ID          int64
	Title       string
	Description sql.NullString
	Created     time.Time
	Modified    time.Time
	permissions []*Permission
}

// Database model for a permission object.
type Permission struct {
	ID          int64
	Title       string
	Description sql.NullString
	Created     time.Time
	Modified    time.Time
}

const (
	getRoleSQL = "SELECT title, description, created, modified FROM roles WHERE id=$1"
)

func GetRole(ctx context.Context, id int64) (role *Role, err error) {
	role = &Role{
		ID: id,
	}

	var tx *sql.Tx
	if tx, err = db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true}); err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if err = tx.QueryRow(getRoleSQL, role.ID).Scan(&role.Title, &role.Description, &role.Created, &role.Modified); err != nil {
		return nil, err
	}

	tx.Commit()
	return role, nil
}

const (
	rolePermsSQL = "SELECT p.id, p.title, p.description, p.created, p.modified FROM role_permissions rp JOIN permissions p ON rp.permission_id=p.id WHERE rp.role_id=$1"
)

func (r *Role) Permissions(ctx context.Context, reset bool) (_ []*Permission, err error) {
	if reset || len(r.permissions) == 0 {
		var tx *sql.Tx
		if tx, err = db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true}); err != nil {
			return nil, err
		}
		defer tx.Rollback()

		var rows *sql.Rows
		r.permissions = make([]*Permission, 0, 4)
		if rows, err = tx.Query(rolePermsSQL, r.ID); err != nil {
			return nil, err
		}

		for rows.Next() {
			permission := &Permission{}
			if err = rows.Scan(&permission.ID, &permission.Title, &permission.Description, &permission.Created, &permission.Modified); err != nil {
				return nil, err
			}

			r.permissions = append(r.permissions, permission)
		}

		tx.Commit()
	}

	return r.permissions, nil
}
