// Copyright (c) 2022, Geert JM Vanderkelen

package xmysql_test

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/go-sql-driver/mysql"
	"github.com/golistic/pxmysql/mysqlerrors"
	"github.com/golistic/xt"

	"github.com/golistic/xmysql"

	_ "github.com/go-sql-driver/mysql" // activate SQL driver 'mysql`
	_ "github.com/golistic/pxmysql"    // activate SQL driver 'pxmysql`
)

func TestNewError(t *testing.T) {

	t.Run("golistic/pxmysql", func(t *testing.T) {
		t.Run("connection refused", func(t *testing.T) {
			db, err := sql.Open("mysqlpx", ":@tcp(127.0.0.1:33445)/")
			xt.OK(t, err)
			defer func() { _ = db.Close() }()

			haveErr := db.Ping()

			xt.Eq(t, "unknown MySQL server host '127.0.0.1:33445' (connection refused) [2005:HY000]", haveErr.Error())
		})

		t.Run("driver error", func(t *testing.T) {
			myErr := &xmysql.Error{
				DriverError: &mysqlerrors.Error{
					Inner:      fmt.Errorf("connection refused"),
					Message:    "unknown MySQL server host '%s' (%w)",
					Code:       2005,
					SQLState:   "HY000",
					Parameters: []any{"127.0.0.1:3306", fmt.Errorf("connection refused")},
				},
				Number: 2005,
			}

			xt.Eq(t, "unknown MySQL server host '127.0.0.1:3306' (connection refused) [2005:HY000]", myErr.Error())
		})

	})

	t.Run("go-sql-driver/mysql", func(t *testing.T) {
		t.Run("connection refused", func(t *testing.T) {
			db, err := sql.Open("mysql", ":@tcp(127.0.0.1:33445)/")
			xt.OK(t, err)
			defer func() { _ = db.Close() }()

			haveErr := xmysql.NewError(db.Ping())

			xt.Eq(t, "unknown MySQL server host '127.0.0.1:33445' (connection refused) [2005:HY000]", haveErr.Error())
		})

		t.Run("driver error", func(t *testing.T) {
			myErr := &mysql.MySQLError{
				Number:  1046,
				Message: "No database selected",
			}

			err := xmysql.NewError(myErr)
			xt.Eq(t, "no database selected", err.Error())
		})
	})

}
