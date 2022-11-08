// Copyright (c) 2022, Geert JM Vanderkelen

package xmysql

import (
	"fmt"
	"regexp"
	"runtime"
	"strings"

	"github.com/geertjanvdk/xkit/xlog"
	"github.com/go-sql-driver/mysql"
)

var (
	reQuoteStrings = regexp.MustCompile(`'(.*?)'`)
	reGoPkg        = regexp.MustCompile(`.*?/go/(src|pkg)/`)
	reCnxRefused   = regexp.MustCompile(`dial (\w+) (.*?): connect: connection refused`)
)

// MySQL Server errors which are handled by this package.
const (
	ErrDBCreateExists uint16 = 1007
	ErrDBDropExists          = 1008
	ErrDubEntry              = 1062
)

// Error wraps mysql.MySQLError with additional information such
// as file information where the error occurred, query with values
// when available, and normalized and nicer message.
type Error struct {
	*mysql.MySQLError
	Err      error
	Query    string
	Values   []any
	Filename string
	Line     int
}

// NewError returns a new xmysql.Error, storing err.
// The filename and line number where the error occurred is saved.
// Panics when the app.Package is not part of the caller's filename.
func NewError(err error) Error {
	return newError(err)
}

// NewErrorSprintf returns a new xmysql.Error, storing err, but sets also the
// message to format and its optional arguments.
func NewErrorSprintf(err error, format string, a ...any) Error {
	e := newError(err)
	e.Message = fmt.Sprintf(format, a...)
	return e
}

// NewErrorQuery returns a new xmysql.Error, storing err with query
// and its values that were interpolated.
// Note: use this with care and only in debugging/development situations.
func NewErrorQuery(err error, query string, values []any) Error {
	e := newError(err)
	e.Query = query
	e.Values = values
	return e
}

func newError(err error) Error {
	myErr, ok := err.(*mysql.MySQLError)

	e := Error{Err: err}
	if ok {
		e.MySQLError = myErr
	}

	var fn string
	var line int
	_, fn, line, ok = runtime.Caller(3)
	if strings.Contains(fn, "golistic/xmysql") {
		_, fn, line, _ = runtime.Caller(2)
	}

	e.Filename = reGoPkg.ReplaceAllString(fn, "")
	e.Line = line

	return e
}

// Error returns the string representation of the e.
func (e Error) Error() string {
	var logMsg string

	if e.MySQLError != nil && e.Message != "" {
		logMsg = e.Message
	}

	logEntry := xlog.WithFields(xlog.Fields{
		"mysqlError":       e.Err.Error(),
		xlog.FieldFileLine: fmt.Sprintf("%s:%d", e.Filename, e.Line),
	})

	if e.MySQLError != nil && e.Number > 0 {
		m := e.Message
		m = strings.Replace(m, "an't", "annot", -1)
		m = strings.Replace(m, "oesn't", "does not", -1)
		switch e.Number {
		case ErrDBCreateExists:
			parts := reQuoteStrings.FindAllStringSubmatch(m, 1)
			logMsg = fmt.Sprintf("schema '%s' not available", parts[0][1])
		case ErrDBDropExists:
			parts := reQuoteStrings.FindAllStringSubmatch(m, 1)
			logMsg = fmt.Sprintf("schema '%s' does not exists", parts[0][1])
		case ErrDubEntry:
			parts := reQuoteStrings.FindAllStringSubmatch(m, 2)
			logMsg = fmt.Sprintf("'%s' not available", parts[0][1])
		default:
			logMsg = e.Err.Error()
		}
	} else {
		logMsg = e.Err.Error()
		parts := reCnxRefused.FindStringSubmatch(logMsg)
		if parts != nil {
			logMsg = fmt.Sprintf("MySQL connection refused (tried %s)", parts[2])
		}
	}

	if e.Query != "" {
		logEntry.WithField("query", e.Query)
		if e.Values != nil {
			strValues := make([]string, len(e.Values))
			for i, value := range e.Values {
				switch v := value.(type) {
				case []byte:
					strValues[i] = string(v)
				default:
					strValues[i] = fmt.Sprintf("%#v", v)
				}
			}
			logEntry.WithField("queryValues", e.Values)
		} else {
		}
	}

	logEntry.Error(logMsg)

	return logMsg
}

// IsDBCreateExists returns whether err is Error and ErrDBCreateExists.
func IsDBCreateExists(err error) bool {
	return ErrorIs(err, ErrDBCreateExists)
}

// ErrorIs returns whether err matches the MySQL Server Error number.
func ErrorIs(err error, number uint16) bool {
	e, ok := err.(Error)
	return ok && e.MySQLError != nil && e.Number == number
}

// ErrorTxBegin returns a xmysql.Error, storing err, and setting a fixed
// message 'failed starting transaction'.
func ErrorTxBegin(err error) Error {
	return NewErrorSprintf(err, "failed starting transaction")
}

// ErrorTxCommit returns a xmysql.Error, storing err, and setting a fixed
// message 'failed committing transaction'.
func ErrorTxCommit(err error) Error {
	return NewErrorSprintf(err, "failed committing transaction")
}
