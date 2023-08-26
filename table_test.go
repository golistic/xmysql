// Copyright (c) 2022, Geert JM Vanderkelen

package xmysql

import (
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/golistic/xgo/xsql"
	"github.com/golistic/xt"
)

func TestTableExists(t *testing.T) {
	dns, err := xsql.ReplaceDSNDatabase(testDSN, "information_schema")
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

func TestTableComment(t *testing.T) {
	schemaName := "xmysql_test_table_comments"
	defer func() { _ = DropSchema(testDB, schemaName) }()

	xt.OK(t, CreateSchema(testDB, schemaName))

	dns, err := xsql.ReplaceDSNDatabase(testDSN, schemaName)
	xt.OK(t, err)

	db, err := sql.Open("mysql", dns)
	xt.OK(t, err)
	defer func() { _ = db.Close() }()

	t.Run("comment from existing table", func(t *testing.T) {
		exp := "I am comment"
		ddl := "CREATE TABLE t1 (id INT) COMMENT='" + exp + "'"
		_, err := db.Exec(ddl)
		xt.OK(t, err)

		have, err := TableComment(testDB, "t1", schemaName)
		xt.OK(t, err)
		xt.Eq(t, exp, have)
	})

	t.Run("structured comment from table", func(t *testing.T) {
		ddl := `CREATE TABLE t10 (id INT) COMMENT='{"comment":"I am a string in JSON","hits":32}'`
		_, err := db.Exec(ddl)
		xt.OK(t, err)

		var have = struct {
			Comment string `json:"comment"`
			Hits    int    `json:"hits"`
		}{}

		xt.OK(t, TableCommentJSON(db, "t10", &have))
		xt.OK(t, err)
		xt.Eq(t, "I am a string in JSON", have.Comment)
		xt.Eq(t, 32, have.Hits)
	})

	t.Run("incorrect data in structured comment from table", func(t *testing.T) {
		ddl := `CREATE TABLE t11 (id INT) COMMENT='{"comment":"I am a string in JSON","hits":"oops"}'`
		_, err := db.Exec(ddl)
		xt.OK(t, err)

		var have = struct {
			Comment string `json:"comment"`
			Hits    int    `json:"hits"`
		}{}

		err = TableCommentJSON(db, "t11", &have)
		xt.KO(t, err)
		xt.Eq(t, "json: cannot unmarshal string into Go struct field .hits of type int", err.Error())
	})

	t.Run("empty comment from existing table", func(t *testing.T) {
		exp := ""
		ddl := "CREATE TABLE t2 (id INT)"
		_, err := db.Exec(ddl)
		xt.OK(t, err)

		have, err := TableComment(db, "t2")
		xt.OK(t, err)
		xt.Eq(t, exp, have)
	})

	t.Run("comment from non-existing table", func(t *testing.T) {
		_, err := TableComment(db, schemaName, "something_not_there")
		xt.KO(t, err)
		xt.Assert(t, errors.Is(err, sql.ErrNoRows))
	})

	t.Run("set string comment using current schema", func(t *testing.T) {
		exp := "string comment"
		ddl := "CREATE TABLE t3 (id INT)"
		_, err := db.Exec(ddl)
		xt.OK(t, err)

		xt.OK(t, SetTableComment(db, "t3", exp))
		have, err := TableComment(db, "t3")
		xt.OK(t, err)
		xt.Eq(t, exp, have)
	})

	t.Run("set string comment providing schema", func(t *testing.T) {
		exp := "string comment"
		ddl := "CREATE TABLE t4 (id INT)"
		_, err := db.Exec(ddl)
		xt.OK(t, err)

		xt.OK(t, SetTableComment(testDB, "t4", exp, schemaName))
		have, err := TableComment(testDB, "t4", schemaName)
		xt.OK(t, err)
		xt.Eq(t, exp, have)
	})

	t.Run("set comment as structured data (JSON)", func(t *testing.T) {
		ts := time.Date(2023, 2, 10, 11, 27, 23, 0, time.UTC)
		exp := fmt.Sprintf(`{"comment":"I am a string in JSON","added":"%s"}`, ts.Format(time.RFC3339))
		ddl := "CREATE TABLE t5 (id INT)"
		_, err := db.Exec(ddl)
		xt.OK(t, err)

		xt.OK(t, SetTableCommentJSON(db, "t5", struct {
			Comment string    `json:"comment"`
			Added   time.Time `json:"added"`
		}{
			Comment: "I am a string in JSON",
			Added:   ts,
		}))
		have, err := TableComment(db, "t5")
		xt.OK(t, err)
		xt.Eq(t, exp, have)
	})
}
