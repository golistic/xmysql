// Copyright (c) 2022, Geert JM Vanderkelen

package xmysql

import (
	"regexp"
	"strings"

	"github.com/go-sql-driver/mysql"
)

var reDSNPassword = regexp.MustCompile(`^(.*):[^/]*?(@.*)$`)
var dsnPasswordMask = "********"

// ReplaceDSNDatabase takes dsn and replaces the database name. It
// returns the new Data Source Name.
func ReplaceDSNDatabase(dsn string, name string) (string, error) {
	config, err := mysql.ParseDSN(dsn)
	if err != nil {
		return "", NewError(err)
	}

	config.DBName = name

	newDSN := config.FormatDSN()
	if _, err := mysql.ParseDSN(dsn); err != nil {
		return "", NewError(err)
	}

	return newDSN, nil
}

// SetDSNParams sets parameters for the given dsn.
//
// Some parameters are always set, if not provided through parameters or
// when none were given:
// * parseTime = true
func SetDSNParams(dsn string, parameters map[string]string) (string, error) {
	config, err := mysql.ParseDSN(dsn)
	if err != nil {
		return "", NewError(err)
	}

	if config.Params == nil {
		config.Params = map[string]string{}
	}

	for k, v := range parameters {
		config.Params[k] = v
	}

	if _, ok := config.Params["parseTime"]; !ok {
		config.Params["parseTime"] = "true"
	}

	return config.FormatDSN(), nil
}

// MaskPasswordInDSN masks the password within the MySQL data source name dsn. This
// function is usually used when displaying or logging the DSN.
//
// When password is empty (not provided) the mask is added anyway.
// When the DSN is something that was not a DSN, the mask itself is returned to
// prevent possible mistakes.
func MaskPasswordInDSN(dsn string) string {
	var res = dsn

	if reDSNPassword.MatchString(dsn) {
		res = reDSNPassword.ReplaceAllString(dsn, `$1:`+dsnPasswordMask+`$2`)
	} else {
		if strings.Contains(dsn, ":@") {
			strings.Replace(dsn, ":@", ":"+dsnPasswordMask+"@", 1)
		} else {
			strings.Replace(dsn, "@", ":"+dsnPasswordMask+"@", 1)
		}
	}

	if res == dsn {
		return dsnPasswordMask
	}

	return res
}
