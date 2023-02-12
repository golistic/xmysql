// Copyright (c) 2023, Geert JM Vanderkelen

package xmysql

import (
	"fmt"
	"sort"
	"strings"
)

// SQLw takes the SQL statement and substitutes the key/value pairs using
// the placeholder format `$(key)`.
// For example, executing `SQLw("SELECT c1 FROM $(tblName)", tblName, "t1")` results
// in `SELECT c1 FROM $(tblName)`.
// If a key is found within a quoted part of the statement, it is not substituted.
// Dangling keys (key without value) are ignored.
//
// Important! You must use as much as possible parameter placeholders such as '?' or
// '$1'. This is however not always possible: for example, when using a dynamic schema
// or table based on context.
func SQLw(statement string, keyValuePairs ...string) (string, error) {
	queryLen := len(statement)

	values := map[string]string{}

	for i := 0; i < len(keyValuePairs); {
		if i+1 == len(keyValuePairs) {
			// dangling; ignore
			break
		}
		key, val := keyValuePairs[i], keyValuePairs[i+1]
		values[key] = val
		i += 2
	}

	var result strings.Builder
	qb := []byte(statement)
	var quoted byte
	substituted := map[string]struct{}{}
	notProvided := map[string]struct{}{}

	for p := 0; p < queryLen; p++ {
		if qb[p] == quoted {
			// end of quoted part
			quoted = 0
		} else if qb[p] == '\'' || qb[p] == '"' || qb[p] == '`' {
			// start of quoted part
			quoted = qb[p]
		} else if quoted == 0 && qb[p] == '$' && (p+1 < queryLen && qb[p+1] == '(') {
			// handle substitution
			pos := p
			p += 2
			var key strings.Builder
			for ; statement[p] != ')'; p++ {
				if p+1 >= queryLen {
					return "", fmt.Errorf("xmysql: unclosed substitution at position %d", pos)
				}
				key.WriteByte(statement[p])
			}
			if v, ok := values[key.String()]; ok {
				result.WriteString(v)
				substituted[key.String()] = struct{}{}
			} else {
				notProvided[key.String()] = struct{}{}
			}
			continue
		}

		result.WriteByte(qb[p])
	}

	switch {
	case len(values) != len(substituted):
		var missing []string
		for k := range values {
			if _, ok := substituted[k]; !ok {
				missing = append(missing, k)
			}
		}
		sort.Strings(missing)
		return "", fmt.Errorf("xmysql: placeholder missing for %s", strings.Join(missing, ","))
	case len(notProvided) > 0:
		var missing []string
		for k := range notProvided {
			missing = append(missing, k)
		}
		sort.Strings(missing)
		return "", fmt.Errorf("xmysql: key/value missing for %s", strings.Join(missing, ","))
	}

	return result.String(), nil
}

// MustSQLw calls SQLw but instead of returning errors, it panics.
func MustSQLw(statement string, keyValuePairs ...string) string {
	s, err := SQLw(statement, keyValuePairs...)
	if err != nil {
		panic(err)
	}
	return s
}
