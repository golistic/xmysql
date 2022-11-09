// Copyright (c) 2022, Geert JM Vanderkelen

package xmysql

import (
	"database/sql"
	"testing"

	"github.com/geertjanvdk/xkit/xt"
)

func TestTableExists(t *testing.T) {
	dns, err := ReplaceDSNDatabase(testDSN, "information_schema")
	xt.OK(t, err)

	db, err := sql.Open("mysql", dns)
	xt.OK(t, err)
	defer func() { _ = db.Close() }()

	t.Run("table exists", func(t *testing.T) {
		have, err := TableExists(db, "TABLES")
		xt.OK(t, err)
		xt.Assert(t, have)
	})

	t.Run("table does not exist", func(t *testing.T) {
		have, err := TableExists(db, "mysqlmysqlmysql")
		xt.OK(t, err)
		xt.Assert(t, !have)
	})

	t.Run("error", func(t *testing.T) {
		db, err := sql.Open("mysql", "root:mysql@tcp(127.0.0.1:12345)/?parseTime=true")
		xt.OK(t, err)

		have, err := TableExists(db, "mysqlmysqlmysql")
		xt.KO(t, err)
		xt.Assert(t, !have)
	})
}
