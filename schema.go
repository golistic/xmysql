// Copyright (c) 2022, Geert JM Vanderkelen

package xmysql

import (
	"context"
	"database/sql"
	"time"
)

// CreateSchema creates schema (or database) using the already open connection db.
//
// When error is returned, it is of type xmysql.Error.
func CreateSchema(db *sql.DB, name string) error {
	if _, err := db.ExecContext(context.Background(), "CREATE SCHEMA "+name); err != nil {
		return NewError(err)
	}

	return nil
}

// DropSchema drops schema (or database) using the already open connection db.
//
// When error is returned, it is of type xmysql.Error.
func DropSchema(db *sql.DB, name string) error {
	if _, err := db.ExecContext(context.Background(), "DROP SCHEMA "+name); err != nil {
		return NewError(err)
	}

	return nil
}

// SchemaExists returns whether schema with given name exists.
func SchemaExists(db *sql.DB, name string) (bool, error) {
	q := "SELECT 1 FROM information_schema.SCHEMATA WHERE SCHEMA_NAME = ?"

	var cancel context.CancelFunc
	ctx := context.Background()
	ctx, cancel = context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var n int
	if err := db.QueryRowContext(ctx, q, name).Scan(&n); err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, NewError(err)
	}
	return true, nil
}

// CurrentSchema returns the name of the current schema of db.
func CurrentSchema(db *sql.DB) (string, error) {
	q := "SELECT SCHEMA()"

	var cancel context.CancelFunc
	ctx := context.Background()
	ctx, cancel = context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var name *string
	if err := db.QueryRowContext(ctx, q).Scan(&name); err != nil {
		return "", NewError(err)
	}

	if name == nil {
		return "", nil
	}

	return *name, nil
}
