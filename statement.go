package gom

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/gchaincl/dotsql"
)

var provider *StmtProvider
var _ Preparer = &EmbeddedStmt{}

func init() {
	provider = &StmtProvider{
		Repository: make(map[string]string),
	}
}

type Param = interface{}

type EmbeddedStmt struct {
	Query  string
	Params []Param
}

func (stmt *EmbeddedStmt) Prepare() (string, map[string]interface{}) {
	query := stmt.Query
	params := make(map[string]interface{})
	buffer := make([]byte, 0, len(query)+10)

	var i, j int

	for i = strings.Index(query, "?"); i != -1; i = strings.Index(query, "?") {
		name := fmt.Sprintf("arg%d", j)
		params[name] = stmt.Params[j]

		buffer = append(buffer, query[:i]...)
		buffer = append(buffer, ':')
		buffer = append(buffer, name...)

		query = query[i+1:]
		j = j + 1
	}

	query = string(append(buffer, query...))
	return query, params
}

type StmtProvider struct {
	Repository map[string]string
}

func (p *StmtProvider) Load(r io.Reader) error {
	scanner := &dotsql.Scanner{}
	stmts := scanner.Run(bufio.NewScanner(r))

	for name, stmt := range stmts {
		if _, ok := p.Repository[name]; ok {
			return fmt.Errorf("Statement '%s' already exists", name)
		}

		p.Repository[name] = stmt
	}

	return nil
}

func (p *StmtProvider) Statement(name string, params ...Param) *EmbeddedStmt {
	if query, ok := p.Repository[name]; ok {
		return &EmbeddedStmt{
			Query:  query,
			Params: params,
		}
	}
	return nil
}

func Load(r io.Reader) error {
	return provider.Load(r)
}

func Statement(name string, params ...Param) *EmbeddedStmt {
	return provider.Statement(name, params...)
}
