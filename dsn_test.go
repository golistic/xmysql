// Copyright (c) 2022, Geert JM Vanderkelen

package xmysql

import (
	"fmt"
	"strings"
	"testing"

	"github.com/golistic/xt"
)

var baseDSN = "u:pwd@tcp(127.0.0.1:3306)/"

func TestReplaceDSNDatabase(t *testing.T) {
	t.Run("using DSN with no database name", func(t *testing.T) {
		have := baseDSN
		exp := have + "foo"

		dsn, err := ReplaceDSNDatabase(have, "foo")

		xt.OK(t, err)
		xt.Eq(t, exp, dsn)
	})

	t.Run("using DSN with database name", func(t *testing.T) {
		have := baseDSN + "bar"
		exp := baseDSN + "foo"

		dsn, err := ReplaceDSNDatabase(have, "foo")

		xt.OK(t, err)
		xt.Eq(t, exp, dsn)
	})
}

func TestSetDSNParams(t *testing.T) {
	t.Run("set some parameter", func(t *testing.T) {
		dsn, err := SetDSNParams(baseDSN, map[string]string{
			"fooBar": "1234",
		})
		xt.OK(t, err)
		xt.Eq(t, "u:pwd@tcp(127.0.0.1:3306)/?fooBar=1234&parseTime=true", dsn)
	})
}

func TestMaskPasswordInDSN(t *testing.T) {
	t.Run("mask password", func(t *testing.T) {
		have := baseDSN
		exp := strings.Replace(have, ":pwd", ":"+dsnPasswordMask, 1)

		xt.Eq(t, exp, MaskPasswordInDSN(have))
	})

	t.Run("dsn have no password", func(t *testing.T) {
		have := strings.Replace(baseDSN, ":pwd", ":", 1)
		exp := strings.Replace(have, ":@", ":"+dsnPasswordMask+"@", 1)
		fmt.Println("### have", have)
		fmt.Println("### exp", exp)

		xt.Eq(t, exp, MaskPasswordInDSN(have))
	})

	t.Run("something not DSN", func(t *testing.T) {
		xt.Eq(t, dsnPasswordMask, MaskPasswordInDSN("foobar"))
	})
}
