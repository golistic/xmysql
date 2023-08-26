// Copyright (c) 2023, Geert JM Vanderkelen

package xmysql

import (
	"github.com/golistic/xgo/xsql"
)

// ReplaceDSNDatabase takes dsn and replaces the database name. It
// returns the new Data Source Name.
// Deprecated: use github.com/golistic/xgo/xsql
func ReplaceDSNDatabase(dsn string, name string) (string, error) {
	return xsql.ReplaceDSNDatabase(dsn, name)
}

// SetDSNParams sets parameters for the given dsn.
//
// Some parameters are always set, if not provided through parameters or
// when none were given:
// * parseTime = true
// Deprecated: use github.com/golistic/xgo/xsql
func SetDSNParams(dsn string, parameters map[string]string) (string, error) {
	return xsql.SetDSNOptions(dsn, parameters)
}

// MaskPasswordInDSN masks the password within the MySQL data source name dsn. This
// function is usually used when displaying or logging the DSN.
//
// When password is empty (not provided) the mask is added anyway.
// When the DSN is something that was not a DSN, the mask itself is returned to
// prevent possible mistakes.
// Deprecated: use github.com/golistic/xgo/xsql
func MaskPasswordInDSN(dsn string) string {
	return xsql.MaskPasswordInDSN(dsn)
}
