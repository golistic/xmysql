// Copyright (c) 2022, Geert JM Vanderkelen

package xmysql

import "database/sql"

// GlobalVariable returns value of a global variable by its name.
func GlobalVariable(db *sql.DB, name string) (string, error) {
	var v string
	q := "SELECT VARIABLE_VALUE FROM performance_schema.global_variables WHERE VARIABLE_NAME = ?"
	if err := db.QueryRow(q, name).Scan(&v); err != nil {
		return "", err
	}

	return v, nil
}
