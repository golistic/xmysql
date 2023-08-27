// Copyright (c) 2022, Geert JM Vanderkelen

package xmysql

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/golistic/xgo/xstrings"
)

type LogOutput string

const (
	LogOutputNone  LogOutput = "NONE"
	LogOutputFile  LogOutput = "FILE"
	LogOutputTable LogOutput = "TABLE"
)

type GeneralLogEvent struct {
	Time        time.Time
	UserHost    string
	ThreadID    int
	ServerID    int
	CommandType string
	Argument    string
}

const generalLogResultHardLimit = 1000

// SetLogOutput sets the destination of the general and slow query logs.
// It is possible to specify multiple destination. If LogOutputNone is provided,
// it will take precedence over all others.
// No-op if no output is provided.
func SetLogOutput(db *sql.DB, outputs ...LogOutput) error {
	switch {
	case len(outputs) == 0:
		// no-op
		return nil
	case xstrings.SliceHas(outputs, LogOutputNone):
		outputs = []LogOutput{LogOutputNone}
	}

	if _, err := db.Exec("SET GLOBAL log_output=?", xstrings.Join(outputs, ",")); err != nil {
		return err
	}
	return nil
}

// EnableGeneralLog turns on the General Log.
// Since all queries are logged, it is recommended not doing this on a production instance.
func EnableGeneralLog(db *sql.DB) error {
	if _, err := db.Exec("SET GLOBAL general_log = 'ON'"); err != nil {
		return err
	}
	return nil
}

// GeneralQueryLogEnabled turns whether the General Query Log is enabled.
// Since all queries are logged, it is recommended not doing this on a production instance.
func GeneralQueryLogEnabled(db *sql.DB) (bool, error) {
	var v string
	v, err := GlobalVariable(db, "general_log")
	if err != nil {
		return false, nil
	}

	return v == "ON", nil
}

// DisableGeneralLog turns off the General Log.
func DisableGeneralLog(db *sql.DB) error {
	if _, err := db.Exec("SET GLOBAL general_log = 'OFF'"); err != nil {
		return err
	}
	return nil
}

// GetGeneralLogEvents retrieves events from the General Query Log using the
// argLike string, if not empty, to filter through the argument field. When limit is 0, or
// over the hard limit of 1000, the hard limit will be used.
// Note that this only works when log output TABLE is active.
func GetGeneralLogEvents(db *sql.DB, argLike string, limit int) ([]*GeneralLogEvent, error) {
	q := "SELECT event_time, user_host, thread_id, server_id, command_type, argument " +
		"FROM mysql.general_log"

	var values []any
	if argLike != "" {
		q += " WHERE argument LIKE ?"
		values = append(values, argLike)
	}

	q += " ORDER BY event_time ASC"

	if limit == 0 || limit > generalLogResultHardLimit {
		limit = generalLogResultHardLimit
	}

	q += fmt.Sprintf(" LIMIT %d", limit)

	rows, err := db.Query(q, values...)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	var events []*GeneralLogEvent
	for rows.Next() {
		ev := &GeneralLogEvent{}
		err := rows.Scan(&ev.Time, &ev.UserHost, &ev.ThreadID, &ev.ServerID, &ev.CommandType, &ev.Argument)
		if err != nil {
			return nil, err
		}
		events = append(events, ev)
	}

	return events, nil
}

// FlushGeneralLog will flush the general log file and truncate the general_log table.
// During these operations, the genera log is disabled, and enabled again after when
// it was enabled before.
func FlushGeneralLog(db *sql.DB) error {
	enabled, err := GeneralQueryLogEnabled(db)
	if err != nil {
		return err
	}

	if enabled {
		if err := DisableGeneralLog(db); err != nil {
			return err
		}
	}

	if _, err := db.Exec("FLUSH GENERAL LOGS"); err != nil {
		return err
	}

	if _, err := db.Exec("TRUNCATE TABLE mysql.general_log"); err != nil {
		return err
	}

	if enabled {
		if err := EnableGeneralLog(db); err != nil {
			return err
		}
	}

	return nil
}
