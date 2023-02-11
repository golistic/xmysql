// Copyright (c) 2022, Geert JM Vanderkelen

package xmysql

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
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

// TableComment retrieves the comment of a table. If schema is not provided, current schema
// will be used.
func TableComment(db *sql.DB, table string, schema ...string) (string, error) {
	dml := `SELECT TABLE_COMMENT FROM information_schema.TABLES WHERE TABLE_NAME = ? AND TABLE_SCHEMA = SCHEMA()`
	args := []any{table}
	if len(schema) > 0 && schema[0] != "" {
		dml = `SELECT TABLE_COMMENT FROM information_schema.TABLES WHERE TABLE_NAME = ? AND TABLE_SCHEMA = ?`
		args = append(args, schema[0])
	}

	var comment string
	if err := db.QueryRow(dml, args...).Scan(&comment); err != nil {
		return "", err
	}

	return comment, nil
}

// TableCommentJSON retrieves the comment of a table, unmarshalls it as JSON, and stores it in
// dest. If schema is not provided, current schema will be used.
// Panics for same reasons as Go's json.Unmarshal would.
func TableCommentJSON(db *sql.DB, table string, dest any, schema ...string) error {
	dml := `SELECT TABLE_COMMENT FROM information_schema.TABLES WHERE TABLE_NAME = ? AND TABLE_SCHEMA = SCHEMA()`
	args := []any{table}
	if len(schema) > 0 && schema[0] != "" {
		dml = `SELECT TABLE_COMMENT FROM information_schema.TABLES WHERE TABLE_NAME = ? AND TABLE_SCHEMA = ?`
		args = append(args, schema[0])
	}

	var comment string
	if err := db.QueryRow(dml, args...).Scan(&comment); err != nil {
		return err
	}

	return json.Unmarshal([]byte(comment), dest)
}

// SetTableComment sets the table's comment. If schema is not provided, current schema
// will be used.
func SetTableComment(db *sql.DB, table string, comment string, schema ...string) error {
	ddl := fmt.Sprintf("ALTER TABLE %s COMMENT = '%%s'", table)
	if len(schema) > 0 && schema[0] != "" {
		ddl = fmt.Sprintf("ALTER TABLE %s.%s COMMENT = '%%s'", schema[0], table)
	}

	if _, err := db.Exec(fmt.Sprintf(ddl, comment)); err != nil {
		return err
	}

	return nil
}

// SetTableCommentJSON sets the table's comment marshalled as JSON. If schema is not provided, current schema
// will be used.
func SetTableCommentJSON(db *sql.DB, table string, comment any, schema ...string) error {
	data, err := json.Marshal(comment)
	if err != nil {
		return err
	}

	ddl := fmt.Sprintf("ALTER TABLE %s COMMENT = '%%s'", table)
	if len(schema) > 0 && schema[0] != "" {
		ddl = fmt.Sprintf("ALTER TABLE %s.%s COMMENT = '%%s'", schema[0], table)
	}

	if _, err := db.Exec(fmt.Sprintf(ddl, string(data))); err != nil {
		return err
	}

	return nil
}
