package gom

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/gchaincl/dotsql"
)

var _ Preparer = &Cmd{}

type Cmd struct {
	Query  string
	Params []Param
}

func (cmd *Cmd) Prepare() (string, map[string]interface{}) {
	query := cmd.Query
	params := make(map[string]interface{})
	buffer := &bytes.Buffer{}

	var i, j int

	for i = strings.Index(query, "?"); i != -1; i = strings.Index(query, "?") {
		name := fmt.Sprintf("arg%d", j)
		part := fmt.Sprintf("%s:%s", query[:i], name)
		params[name] = cmd.Params[j]

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

type CmdProvider struct {
	Repository map[string]string
}

func (p *CmdProvider) Load(r io.Reader) error {
	scanner := &dotsql.Scanner{}
	stmts := scanner.Run(bufio.NewScanner(r))

	for name, stmt := range stmts {
		if _, ok := p.Repository[name]; ok {
			return fmt.Errorf("Command '%s' already exists", name)
		}

		p.Repository[name] = stmt
	}

	return nil
}

func (p *CmdProvider) Command(name string, params ...Param) (*Cmd, error) {
	if query, ok := p.Repository[name]; ok {
		return &Cmd{
			Query:  query,
			Params: params,
		}, nil
	}

	return nil, fmt.Errorf("Command '%s' not found", name)
}
