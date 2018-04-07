package script

import (
	"bytes"
	"fmt"
	"strings"
)

// Cme represents a single command from SQL script.
type Cmd struct {
	query  string
	params []Param
}

// Prepare prepares the command for execution.
func (cmd *Cmd) Prepare() (string, map[string]interface{}) {
	query := cmd.query
	params := make(map[string]interface{})
	buffer := &bytes.Buffer{}

	if len(cmd.params) == 0 {
		return query, params
	}

	var i, j int

	for i = strings.Index(query, "?"); i != -1; i = strings.Index(query, "?") {
		if j >= len(cmd.params) {
			return "", nil
		}

		name := fmt.Sprintf("arg%d", j)
		part := fmt.Sprintf("%s:%s", query[:i], name)
		params[name] = cmd.params[j]

		if _, err := buffer.WriteString(part); err != nil {
			return "", nil
		}

		query = query[i+1:]
		j = j + 1
	}

	if _, err := buffer.WriteString(query); err != nil {
		return "", nil
	}

	query = buffer.String()
	return query, params
}

// SQL create a new command from raw query
func SQL(query string, params ...Param) *Cmd {
	return &Cmd{
		query:  query,
		params: params,
	}
}
