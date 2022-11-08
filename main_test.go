// Copyright (c) 2022, Geert JM Vanderkelen

package xmysql

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
)

var (
	testExitCode int
	testErr      error
	testDSN      string
	testDB       *sql.DB
)

func testTearDown() {
	if testErr != nil {
		fmt.Println(testErr)
		os.Exit(1)
	}
}

func TestMain(m *testing.M) {
	defer func() { os.Exit(testExitCode) }()
	defer testTearDown()

	testDSN = os.Getenv("XMYSQL_DSN")
	if testDSN == "" {
		testDSN = "root:mysql@tcp(127.0.0.1:13399)/?parseTime=true"
	}

	testDB, testErr = sql.Open("mysql", testDSN)
	if testErr != nil {
		return
	}

	testExitCode = m.Run()
}
