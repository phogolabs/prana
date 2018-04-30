package sqlexec

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"

	"github.com/jmoiron/sqlx/reflectx"
)

var _ Query = &Stmt{}
var _ Query = &NamedStmt{}

// Stmt represents a single command from SQL sqlexec.
type Stmt struct {
	query  string
	params []Param
}

// Prepare prepares the command for execution.
func (cmd *Stmt) Prepare() (string, map[string]interface{}) {
	query := cmd.query
	params := make(map[string]interface{})
	buffer := &bytes.Buffer{}

	if len(cmd.params) == 0 {
		return query, params
	}

	var i, j int

	for i = strings.Index(query, "?"); i != -1; i = strings.Index(query, "?") {
		name := fmt.Sprintf("arg%d", j)
		part := fmt.Sprintf("%s:%s", query[:i], name)

		if j < len(cmd.params) {
			params[name] = cmd.params[j]
		}

		buffer.WriteString(part)
		query = query[i+1:]
		j = j + 1
	}

	buffer.WriteString(query)
	query = buffer.String()
	return query, params
}

// NamedStmt is command that can use named parameters
type NamedStmt struct {
	query string
	param Param
}

// Prepare prepares the command for execution.
func (cmd *NamedStmt) Prepare() (string, map[string]interface{}) {
	params := make(map[string]interface{})

	args, ok := cmd.param.(map[string]interface{})
	if !ok {
		args = cmd.bindArgs(cmd.param)
	}

	for k, v := range args {
		params[k] = v
	}

	return cmd.query, params
}

func (cmd *NamedStmt) bindArgs(param Param) map[string]interface{} {
	params := make(map[string]interface{})
	mapper := reflectx.NewMapper("db")

	v := reflect.ValueOf(param)

	for v = reflect.ValueOf(param); v.Kind() == reflect.Ptr; {
		v = v.Elem()
	}

	for key, value := range mapper.FieldMap(v) {
		key = strings.ToLower(key)
		params[key] = value.Interface()
	}

	return params
}

// SQL create a new command from raw query
func SQL(query string, params ...Param) Query {
	return &Stmt{
		query:  query,
		params: params,
	}
}

// NamedSQL create a new named command from raw query
func NamedSQL(query string, param Param) Query {
	return &NamedStmt{
		query: query,
		param: param,
	}
}
