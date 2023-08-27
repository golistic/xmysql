// Copyright (c) 2023, Geert JM Vanderkelen

package xmysql

import (
	"testing"

	"github.com/golistic/xgo/xt"
)

func TestSQLw(t *testing.T) {
	t.Run("common cases", func(t *testing.T) {
		var cases = []struct {
			stmt      string
			keyValues []string
			exp       string
		}{
			{
				exp:       "SELECT 1 FROM t1",
				stmt:      "SELECT 1 FROM $(tblName)",
				keyValues: []string{"tblName", "t1"},
			},
			{
				exp:       "SELECT 1 AS t1_value FROM t1",
				stmt:      "SELECT 1 AS $(tblName)_value FROM $(tblName)",
				keyValues: []string{"tblName", "t1"},
			},
		}

		for _, c := range cases {
			t.Run(c.stmt, func(t *testing.T) {
				xt.Eq(t, c.exp, MustSQLw(c.stmt, c.keyValues...))
			})
		}
	})

	t.Run("placeholders within quotes", func(t *testing.T) {
		var cases = map[string]struct {
			stmt      string
			keyValues []string
			exp       string
		}{
			"single quotes": {
				exp:       "SELECT 1 FROM t1 WHERE c1 = '$(value)'",
				stmt:      "SELECT 1 FROM $(tblName) WHERE c1 = '$(value)'",
				keyValues: []string{"tblName", "t1"},
			},
			"double quotes": {
				exp:       `SELECT 1 FROM t1 WHERE c1 = "$(value)"`,
				stmt:      `SELECT 1 FROM $(tblName) WHERE c1 = "$(value)"`,
				keyValues: []string{"tblName", "t1"},
			},
			"backticks": {
				exp:       "SELECT 1 FROM t1 WHERE c1 = `$(value)` OR c2 = 5",
				stmt:      "SELECT 1 FROM $(tblName) WHERE c1 = `$(value)` OR c2 = $(value)",
				keyValues: []string{"tblName", "t1", "value", "5"},
			},
		}

		for _, c := range cases {
			t.Run(c.stmt, func(t *testing.T) {
				xt.Eq(t, c.exp, MustSQLw(c.stmt, c.keyValues...))
			})
		}
	})

	t.Run("key without value is ignored (dangling)", func(t *testing.T) {
		exp := "SELECT 1 FROM t1"
		stmt := "SELECT 1 FROM $(tblName)"
		keyValues := []string{"tblName", "t1", "lonelyKey"}

		xt.Eq(t, exp, MustSQLw(stmt, keyValues...))
	})

	t.Run("missing key", func(t *testing.T) {
		_, err := SQLw("SELECT 1 FROM $(tblName) JOIN $(otherTableName)")
		xt.KO(t, err)
		xt.Eq(t, "xmysql: key/value missing for otherTableName,tblName", err.Error())
	})

	t.Run("missing placeholders", func(t *testing.T) {
		_, err := SQLw("SELECT 1 FROM $(tblName)", "tblName", "t1", "column", "c1")
		xt.KO(t, err)
		xt.Eq(t, "xmysql: placeholder missing for column", err.Error())
	})
}
