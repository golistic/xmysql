// Copyright (c) 2022, Geert JM Vanderkelen

package xmysql

import (
	"fmt"
	"net"
	"regexp"
	"runtime"
	"strings"

	"github.com/geertjanvdk/xkit/xlog"
	"github.com/go-sql-driver/mysql"
	"github.com/golistic/pxmysql/mysqlerrors"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var (
	reQuoteStrings       = regexp.MustCompile(`'(.*?)'`)
	reGoPkg              = regexp.MustCompile(`.*?/go/(src|pkg)/`)
	reCnxRefused         = regexp.MustCompile(`dial (\w+) (.*?): connect: connection refused`)
	reGoMySQLDriverError = regexp.MustCompile(`Error (\w+): (.*)`)
)

// Error numbers of MySQL handled by this package.
const (
	ErrDBCreateExists    int = 1007
	ErrDBDropExists      int = 1008
	ErrDubEntry          int = 1062
	ErrClientConnRefused int = 2005
)

// Error wraps mysql.MySQLError with additional information such
// as file information where the error occurred, query with values
// when available, and normalized and nicer message.
type Error struct {
	Message     string
	DriverError error
	Query       string
	Values      []any
	Filename    string
	Line        int
	Number      int
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
	e := Error{DriverError: err}

	switch v := err.(type) {
	case *mysql.MySQLError:
		// from go-sql-driver/mysql
		e.Number = int(v.Number)
	case *mysqlerrors.Error:
		// from golistic/pxmysql
		e.Number = v.Code
	}

	_, fn, line, ok := runtime.Caller(3)
	if ok {
		if strings.Contains(fn, "golistic/xmysql") {
			_, fn, line, _ = runtime.Caller(2)
		}
	}

	e.Filename = reGoPkg.ReplaceAllString(fn, "")
	e.Line = line

	return e
}

// Error returns the string representation of the e.
func (e Error) Error() string {
	var msg string

	switch {
	case e.Message != "":
		msg = e.Message
	case e.DriverError != nil:
		msg = e.DriverError.Error()
	default:
		msg = "unknown MySQL error"
	}

	logEntry := xlog.WithFields(xlog.Fields{
		"mysqlError":       e.DriverError.Error(),
		xlog.FieldFileLine: fmt.Sprintf("%s:%d", e.Filename, e.Line),
	})

	// if we have an error number, we try to correct the message
	if e.Number > 0 {
		logEntry.WithField("errorNumber", e.Number)
		m := e.Message
		m = strings.Replace(m, "an't", "annot", -1)
		m = strings.Replace(m, "oesn't", "does not", -1)
		switch e.Number {
		case ErrDBCreateExists:
			parts := reQuoteStrings.FindAllStringSubmatch(m, 1)
			msg = fmt.Sprintf("schema '%s' not available", parts[0][1])
		case ErrDBDropExists:
			parts := reQuoteStrings.FindAllStringSubmatch(m, 1)
			msg = fmt.Sprintf("schema '%s' does not exists", parts[0][1])
		case ErrDubEntry:
			parts := reQuoteStrings.FindAllStringSubmatch(m, 2)
			msg = fmt.Sprintf("'%s' not available", parts[0][1])
		default:
			parts := reGoMySQLDriverError.FindStringSubmatch(msg)
			if len(parts) == 3 {
				msg = cases.Lower(language.English, cases.Compact).String(parts[2])
			}
		}
	}

	// Go's 'connection refused' message
	//switch v := e.DriverError.(type) {
	//case *mysql.MySQLError:
	//	parts := reCnxRefused.FindStringSubmatch(msg)
	//	if len(parts) == 3 {
	//		msg = fmt.Sprintf("unknown MySQL server host '%s' (%w)", parts[1])
	//	}
	//}

	switch v := e.DriverError.(type) {
	case *net.OpError:
		const f = "unknown MySQL server host '%s' (%s) [2005:HY000]"

		unwrapped := strings.TrimPrefix(v.Unwrap().Error(), "connect: ")
		parts := reCnxRefused.FindStringSubmatch(msg)
		e.Number = ErrClientConnRefused
		if len(parts) == 3 {
			msg = fmt.Sprintf(f, parts[2], unwrapped)
		} else {
			msg = fmt.Sprintf(f, "<unknown>", unwrapped)
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
		}
	}

	logEntry.Error(msg)

	return msg
}

// IsDBCreateExists returns whether err is Error and ErrDBCreateExists.
func IsDBCreateExists(err error) bool {
	return ErrorIs(err, ErrDBCreateExists)
}

// ErrorIs returns whether err matches the MySQL Server Error number.
func ErrorIs(err error, number int) bool {
	e, ok := err.(Error)
	return ok && e.DriverError != nil && e.Number == number
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
