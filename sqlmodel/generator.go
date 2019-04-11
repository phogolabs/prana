package sqlmodel

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"time"

	"github.com/aymerick/raymond"
	"github.com/phogolabs/parcello"
	"golang.org/x/tools/imports"
)

func init() {
	raymond.RegisterHelper("now", func() raymond.SafeString {
		return raymond.SafeString(time.Now().Format(time.RFC1123))
	})
}

var (
	_ Generator = &Codegen{}
)

// Codegen generates Golang structs from database schema
type Codegen struct {
	// Meta information
	Meta map[string]interface{}
	// Format the code
	Format bool
}

// Generate generates the golang structs from database schema
func (g *Codegen) Generate(ctx *GeneratorContext) error {
	buffer := &bytes.Buffer{}

	if len(ctx.Schema.Tables) == 0 {
		return nil
	}

	template, err := g.template(ctx.Template)
	if err != nil {
		return err
	}

	param := map[string]interface{}{
		"Schema": ctx.Schema,
		"Meta":   g.Meta,
	}

	result, err := raymond.Render(template, param)
	if err != nil {
		return err
	}

	_, err = buffer.WriteString(result)
	if err != nil {
		return err
	}

	if g.Format {
		if err := g.format(ctx.Template, buffer); err != nil {
			return err
		}
	}

	_, err = io.Copy(ctx.Writer, buffer)
	return err
}

func (g *Codegen) template(name string) (string, error) {
	template, err := parcello.Open(fmt.Sprintf("template/%s.mustache", name))
	if err != nil {
		return "", err
	}
	defer template.Close()

	data, err := ioutil.ReadAll(template)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (g *Codegen) format(template string, buffer *bytes.Buffer) error {
	data, err := imports.Process(template, buffer.Bytes(), nil)
	if err != nil {
		return err
	}

	buffer.Reset()

	_, err = buffer.Write(data)
	return err
}
