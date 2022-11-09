// Copyright (c) 2022, Geert JM Vanderkelen

package xmysql

import (
	"context"
	"database/sql"
	"time"
)

// TableExists returns whether table with given name exists in schema db.
func TableExists(db *sql.DB, name string) (bool, error) {
	schema, err := CurrentSchema(db)
	if err != nil {
		return false, err
	}

	q := "SELECT 1 FROM information_schema.TABLES WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?"

	var cancel context.CancelFunc
	ctx := context.Background()
	ctx, cancel = context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var n int
	if err := db.QueryRowContext(ctx, q, schema, name).Scan(&n); err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, NewError(err)
	}
	return true, nil
}
