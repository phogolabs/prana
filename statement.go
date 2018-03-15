package gom

import (
	"bufio"
	"fmt"
	"io"

	"github.com/gchaincl/dotsql"
)

var provider *StmtProvider
var _ Preparer = &EmbeddedStmt{}

func init() {
	provider = &StmtProvider{
		Repository: make(map[string]string),
	}
}

type Params = map[string]interface{}

type EmbeddedStmt struct {
	Query  string
	Params Params
}

func (stmt *EmbeddedStmt) Prepare() (string, map[string]interface{}) {
	return stmt.Query, stmt.Params
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

func (p *StmtProvider) Statement(name string) *EmbeddedStmt {
	return p.StatementWithParams(name, Params{})
}

func (p *StmtProvider) StatementWithParams(name string, params Params) *EmbeddedStmt {
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

func Statement(name string) *EmbeddedStmt {
	return provider.Statement(name)
}

func StatementWithParams(name string, params Params) *EmbeddedStmt {
	return provider.StatementWithParams(name, params)
}
