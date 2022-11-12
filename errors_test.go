// Copyright (c) 2022, Geert JM Vanderkelen

package xmysql

import (
	"testing"

	"github.com/geertjanvdk/xkit/xt"
	"github.com/go-sql-driver/mysql"
)

func TestNewError(t *testing.T) {
	t.Run("not specially handled error", func(t *testing.T) {
		myErr := &mysql.MySQLError{
			Number:  1046,
			Message: "No database selected",
		}

		err := NewError(myErr)
		xt.Eq(t, "no database selected", err.Error())
	})
}
