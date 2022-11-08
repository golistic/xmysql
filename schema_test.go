// Copyright (c) 2022, Geert JM Vanderkelen

package xmysql

import (
	"database/sql"
	"testing"

	"github.com/geertjanvdk/xkit/xt"
)

func TestCreateSchema(t *testing.T) {
	t.Run("create schema", func(t *testing.T) {
		schemaName := "xmysql_test_ifk29389wldow"
		defer func() { _ = DropSchema(testDB, schemaName) }()

		xt.OK(t, CreateSchema(testDB, schemaName))

		// next should fail
		err := CreateSchema(testDB, schemaName)
		xt.KO(t, err)
		xt.Assert(t, IsDBCreateExists(err))
	})
}

func TestDropSchema(t *testing.T) {
	t.Run("create schema", func(t *testing.T) {
		schemaName := "xmysql_test_fiei2o3if"
		defer func() { _ = DropSchema(testDB, schemaName) }()

		xt.OK(t, CreateSchema(testDB, schemaName))
		xt.OK(t, DropSchema(testDB, schemaName))

		// next should fail
		err := DropSchema(testDB, schemaName)
		xt.KO(t, err)
		xt.Assert(t, ErrorIs(err, ErrDBDropExists))
	})
}

func TestSchemaExists(t *testing.T) {
	t.Run("schema exists", func(t *testing.T) {
		have, err := SchemaExists(testDB, "mysql")
		xt.OK(t, err)
		xt.Assert(t, have)
	})

	t.Run("schema does not exists", func(t *testing.T) {
		have, err := SchemaExists(testDB, "mysqlmysqlmysql")
		xt.OK(t, err)
		xt.Assert(t, !have)
	})

	t.Run("error", func(t *testing.T) {
		db, err := sql.Open("mysql", "root:mysql@tcp(127.0.0.1:12345)/?parseTime=true")
		xt.OK(t, err)

		have, err := SchemaExists(db, "mysqlmysqlmysql")
		xt.KO(t, err)
		xt.Assert(t, !have)
	})
}
