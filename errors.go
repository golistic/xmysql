// Copyright (c) 2022, Geert JM Vanderkelen

package xmysql

import (
	"errors"
	"fmt"
	"net"
	"reflect"
	"regexp"
	"runtime"
	"strings"

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

	rv := reflect.ValueOf(err)
	if rv.Kind() == reflect.Pointer {
		rv = rv.Elem()
	}

	if rv.IsValid() && rv.Kind() == reflect.Struct {
		for _, n := range []string{"Code", "Number"} {
			f := rv.FieldByName(n)
			if f.IsValid() && !f.IsZero() {
				switch f.Kind() {
				case reflect.Uint16:
					// github.com/go-sql-driver/mysql uses uint16
					e.Number = int(f.Uint())
				case reflect.Int:
					// github.com/golistic/pxmysql uses int
					e.Number = int(f.Int())
				}
			}
		}
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

	// if we have an error number, we try to correct the message
	if e.Number > 0 {
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

	var v *net.OpError
	switch {
	case errors.As(e.DriverError, &v):
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

	return msg
}

// IsDBCreateExists returns whether err is Error and ErrDBCreateExists.
func IsDBCreateExists(err error) bool {
	return ErrorIs(err, ErrDBCreateExists)
}

// ErrorIs returns whether err matches the MySQL Server Error number.
func ErrorIs(err error, number int) bool {
	var e Error
	ok := errors.As(err, &e)
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
