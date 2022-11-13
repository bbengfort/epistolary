package users

import (
	"context"
	"database/sql"
	"time"

	"github.com/bbengfort/epistolary/pkg/server/db"
	"github.com/bbengfort/epistolary/pkg/server/passwd"
)

// Database model for a user object.
type User struct {
	ID              int64
	FullName        sql.NullString
	Email           string
	Username        string
	Password        string
	RoleID          int64
	LastSeen        sql.NullTime
	PasswordChanged sql.NullTime
	Created         time.Time
	Modified        time.Time
	role            *Role
	permissions     []string
}

const (
	getUserUnameSQL = "SELECT id, full_name, email, password, role_id, last_seen, pwchanged, created, modified FROM users WHERE username=$1"
	getUserEmailSQL = "SELECT id, full_name, username, password, role_id, last_seen, pwchanged, created, modified FROM users WHERE email=$1"
	getUserIDSQL    = "SELECT full_name, email, username, role_id, last_seen, pwchanged, created, modified FROM users WHERE id=$1"
)

// UserFromUsername gets a user and populates the role and permissions if claims is true.
// This function returns the password for login purposes unless claims is false.
func UserFromUsername(ctx context.Context, username string, claims bool) (user *User, err error) {
	user = &User{Username: username}

	var tx *sql.Tx
	if tx, err = db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true}); err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if err = tx.QueryRow(getUserUnameSQL, username).Scan(&user.ID, &user.FullName, &user.Email, &user.Password, &user.RoleID, &user.LastSeen, &user.PasswordChanged, &user.Created, &user.Modified); err != nil {
		return nil, err
	}

	if claims {
		if err = user.fetchRole(tx); err != nil {
			return nil, err
		}

		if err = user.fetchPermissions(tx); err != nil {
			return nil, err
		}
	} else {
		user.Password = ""
	}

	tx.Commit()
	return user, nil
}

// UserFromEmail gets a user and populates the role and permissions if claims is true.
// This function returns the password for login purposes unless claims is false.
func UserFromEmail(ctx context.Context, email string, claims bool) (user *User, err error) {
	user = &User{Email: email}

	var tx *sql.Tx
	if tx, err = db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true}); err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if err = tx.QueryRow(getUserEmailSQL, email).Scan(&user.ID, &user.FullName, &user.Username, &user.Password, &user.RoleID, &user.LastSeen, &user.PasswordChanged, &user.Created, &user.Modified); err != nil {
		return nil, err
	}

	if claims {
		if err = user.fetchRole(tx); err != nil {
			return nil, err
		}

		if err = user.fetchPermissions(tx); err != nil {
			return nil, err
		}
	} else {
		user.Password = ""
	}

	tx.Commit()
	return user, nil
}

// UserFromID gets a user without password or claims and should be used for non-login purposes.
func UserFromID(ctx context.Context, id int64) (user *User, err error) {
	user = &User{ID: id}

	var tx *sql.Tx
	if tx, err = db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true}); err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if err = tx.QueryRow(getUserIDSQL, id).Scan(user.FullName, user.Email, user.Username, user.RoleID, user.LastSeen, user.PasswordChanged, user.Created, user.Modified); err != nil {
		return nil, err
	}

	tx.Commit()
	return user, nil
}

const (
	createUserSQL = "INSERT INTO users (full_name, email, username, password, role_id, pwchanged) VALUES ($1, $2, $3, $4, $5, $6) RETURNING ID;"
	getUserTSSQL  = "SELECT created, modified FROM users WHERE id=$1"
)

// Create a user from the model. The ID, Created, and Modified timestamps will be
// populated on the model after creation. Note that the LastSeen timestamp is ignored
// and the PasswordChange timestamp is set to now (no matter what it was set to before).
// TODO: handle specific database errors like uniqueness constraints or other schema errors.
func (u *User) Create(ctx context.Context) (err error) {
	// Sanity checks: password must be a derived key.
	if !passwd.IsDerivedKey(u.Password) {
		return ErrNotDerivedKey
	}

	// Password changed is now since the password is being created.
	u.PasswordChanged = sql.NullTime{
		Time:  time.Now(),
		Valid: true,
	}

	// Set the default role if required
	if u.RoleID < 1 {
		u.RoleID = DefaultRoleID
	}

	var tx *sql.Tx
	if tx, err = db.BeginTx(ctx, &sql.TxOptions{ReadOnly: false}); err != nil {
		return err
	}
	defer tx.Rollback()

	if err = tx.QueryRow(createUserSQL, u.FullName, u.Email, u.Username, u.Password, u.RoleID, u.PasswordChanged).Scan(&u.ID); err != nil {
		return err
	}

	if err = tx.QueryRow(getUserTSSQL, u.ID).Scan(&u.Created, &u.Modified); err != nil {
		return err
	}

	tx.Commit()
	return nil
}

const (
	updateUserSQL = "UPDATE users SET full_name=$2, email=$3, username=$4, role_id=$5, last_seen=$6 WHERE id=$1"
)

// Update a user's full name, email, username, role id, and/or last seen timestamp. The
// user must have an ID field populated or ErrNoUserID will be returned.
func (u *User) Update(ctx context.Context) (err error) {
	if u.ID < 1 {
		return ErrNoUserID
	}

	var tx *sql.Tx
	if tx, err = db.BeginTx(ctx, &sql.TxOptions{ReadOnly: false}); err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err = tx.Exec(updateUserSQL, u.ID, u.FullName, u.Email, u.Username, u.RoleID, u.LastSeen); err != nil {
		return err
	}

	tx.Commit()
	return nil
}

const (
	updatePasswordSQL = "UPDATE users SET password=$2, pwchanged=$3 WHERE id=$1"
)

// Update a user's password. The user must have an ID field populated and the Password
// must be a derived key or ErrNoUserID or ErrNotDerivedKey will be returned.
func (u *User) UpdatePassword(ctx context.Context) (err error) {
	if u.ID < 1 {
		return ErrNoUserID
	}

	// Sanity checks: password must be a derived key.
	if !passwd.IsDerivedKey(u.Password) {
		return ErrNotDerivedKey
	}

	// Password changed is now since the password is being created.
	u.PasswordChanged = sql.NullTime{
		Time:  time.Now(),
		Valid: true,
	}

	var tx *sql.Tx
	if tx, err = db.BeginTx(ctx, &sql.TxOptions{ReadOnly: false}); err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err = tx.Exec(updatePasswordSQL, u.ID, u.Password, u.PasswordChanged); err != nil {
		return err
	}

	tx.Commit()
	return nil
}

const (
	updateLastSeenSQL = "UPDATE users SET last_seen=$2 WHERE id=$1"
)

// Update a user's last seen timestamp to now. The user must have an ID field populated
// or ErrNotDerivedKey will be returned.
func (u *User) UpdateLastSeen(ctx context.Context) (err error) {
	if u.ID < 1 {
		return ErrNoUserID
	}

	// Update the last seen timestamp
	u.LastSeen = sql.NullTime{
		Time:  time.Now(),
		Valid: true,
	}

	var tx *sql.Tx
	if tx, err = db.BeginTx(ctx, &sql.TxOptions{ReadOnly: false}); err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err = tx.Exec(updateLastSeenSQL, u.ID, u.LastSeen); err != nil {
		return err
	}

	tx.Commit()
	return nil
}

const (
	userRoleSQL = "SELECT id, title, description, created, modified FROM roles WHERE id=$1"
)

// Role returns the role associated with the user. If the role is not cached on the
// struct then a database query is performed and an error may be returned. Use the reset
// bool to force a database query even if the role is cached on the struct.
func (u *User) Role(ctx context.Context, reset bool) (_ *Role, err error) {
	if reset || u.role == nil {
		var tx *sql.Tx
		if tx, err = db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true}); err != nil {
			return nil, err
		}
		defer tx.Rollback()

		if err = u.fetchRole(tx); err != nil {
			return nil, err
		}

		tx.Commit()
	}

	return u.role, nil
}

func (u *User) fetchRole(tx *sql.Tx) (err error) {
	u.role = &Role{}
	if err = tx.QueryRow(userRoleSQL, u.RoleID).Scan(&u.role.ID, &u.role.Title, &u.role.Description, &u.role.Created, &u.role.Modified); err != nil {
		return err
	}
	return nil
}

const (
	userPermsSQL = "SELECT permission FROM user_permissions WHERE user_id=$1"
)

// Permissions returns the claims associated with the user. If the permissions are not
// cached on the struct, then a database query to the user_permissions view is executed.
// Use the reset bool to force a database query even if the permissions are cached.
func (u *User) Permissions(ctx context.Context, reset bool) (_ []string, err error) {
	if reset || len(u.permissions) == 0 {
		var tx *sql.Tx
		if tx, err = db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true}); err != nil {
			return nil, err
		}
		defer tx.Rollback()

		if err = u.fetchPermissions(tx); err != nil {
			return nil, err
		}

		tx.Commit()
	}

	return u.permissions, nil
}

func (u *User) fetchPermissions(tx *sql.Tx) (err error) {
	var rows *sql.Rows
	u.permissions = make([]string, 0, 4)
	if rows, err = tx.Query(userPermsSQL, u.ID); err != nil {
		return err
	}

	for rows.Next() {
		var permission string
		if err = rows.Scan(&permission); err != nil {
			return err
		}
		u.permissions = append(u.permissions, permission)
	}
	return nil
}
