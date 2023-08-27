// Copyright (c) 2022, Geert JM Vanderkelen

package xmysql

import (
	"fmt"
	"testing"

	"github.com/golistic/xgo/xt"
)

func TestEnableGeneralQueryLog(t *testing.T) {
	xt.OK(t, DisableGeneralLog(testDB))
	defer func() { _ = DisableGeneralLog(testDB) }()

	have, err := GeneralQueryLogEnabled(testDB)
	xt.OK(t, err)
	xt.Assert(t, !have)

	xt.OK(t, EnableGeneralLog(testDB))

	have, err = GeneralQueryLogEnabled(testDB)
	xt.OK(t, err)
	xt.Assert(t, have)
}

func TestSetLogOutput(t *testing.T) {
	t.Run("setting as NONE ignores other outputs", func(t *testing.T) {
		xt.OK(t, SetLogOutput(testDB, LogOutputFile, LogOutputNone, LogOutputFile))

		have, err := GlobalVariable(testDB, "log_output")
		xt.OK(t, err)
		xt.Eq(t, LogOutputNone, LogOutput(have))
	})

	t.Run("set multiple outputs", func(t *testing.T) {
		xt.OK(t, SetLogOutput(testDB, LogOutputFile, LogOutputTable))

		exp := "FILE,TABLE"
		have, err := GlobalVariable(testDB, "log_output")
		xt.OK(t, err)
		xt.Eq(t, exp, have)
	})

	t.Run("no-op when no output given", func(t *testing.T) {
		xt.OK(t, SetLogOutput(testDB, LogOutputFile))
		xt.OK(t, SetLogOutput(testDB))

		exp := "FILE"
		have, err := GlobalVariable(testDB, "log_output")
		xt.OK(t, err)
		xt.Eq(t, exp, have)
	})
}

func TestGetGeneralLogEvents(t *testing.T) {
	xt.OK(t, SetLogOutput(testDB, LogOutputTable))
	xt.OK(t, EnableGeneralLog(testDB))
	xt.OK(t, FlushGeneralLog(testDB))

	defer func() { _ = FlushGeneralLog(testDB) }()

	queryFormat := "/* %03d */ SELECT NOW()"
	for i := 0; i < 10; i++ {
		_, err := testDB.Query(fmt.Sprintf(queryFormat, i))
		xt.OK(t, err)
	}

	t.Run("get particular log event", func(t *testing.T) {
		events, err := GetGeneralLogEvents(testDB, "%* 004 *%", 1)
		xt.OK(t, err)
		xt.Eq(t, 1, len(events))

		xt.Eq(t, "/* 004 */ SELECT NOW()", events[0].Argument)
	})

	t.Run("get 3 log events", func(t *testing.T) {
		events, err := GetGeneralLogEvents(testDB, "%* 00_ *%", 3)
		xt.OK(t, err)
		xt.Eq(t, 3, len(events))

		for i := 0; i < 3; i++ {
			xt.Eq(t, fmt.Sprintf(queryFormat, i), events[i].Argument)
		}
	})

	t.Run("get everything up till hard limit", func(t *testing.T) {
		events, err := GetGeneralLogEvents(testDB, "", 0)
		xt.OK(t, err)
		xt.Assert(t, len(events) > 20)
		xt.Assert(t, len(events) < generalLogResultHardLimit)
	})
}
