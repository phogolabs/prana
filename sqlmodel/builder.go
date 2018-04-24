package sqlmodel

import (
	"fmt"
	"strings"
)

var _ TagBuilder = &CompositeTagBuilder{}
var _ TagBuilder = &SQLXTagBuilder{}
var _ TagBuilder = &GORMTagBuilder{}
var _ TagBuilder = &JSONTagBuilder{}
var _ TagBuilder = &XMLTagBuilder{}
var _ TagBuilder = &ValidateTagBuilder{}

// CompositeTagBuilder composes multiple builders
type CompositeTagBuilder []TagBuilder

// Build builds tags for given column
func (composition CompositeTagBuilder) Build(column *Column) string {
	tags := []string{}

	for _, builder := range composition {
		tag := strings.TrimSpace(builder.Build(column))
		if tag == "" {
			continue
		}
		tags = append(tags, tag)
	}

	return fmt.Sprintf("`%s`", strings.Join(tags, " "))
}

// SQLXTagBuilder builds tags for SQLX mapper
type SQLXTagBuilder struct{}

// Build builds tags for given column
func (builder SQLXTagBuilder) Build(column *Column) string {
	options := []string{}
	options = append(options, column.Name)

	if column.Type.IsPrimaryKey {
		options = append(options, "primary_key")
	}

	if column.Type.IsNullable {
		options = append(options, "null")
	} else {
		options = append(options, "not_null")
	}

	if size := column.Type.CharMaxLength; size > 0 {
		options = append(options, fmt.Sprintf("size=%d", size))
	}

	return fmt.Sprintf("db:\"%s\"", strings.Join(options, ","))
}

// GORMTagBuilder builds tags for GORM mapper
type GORMTagBuilder struct{}

// Build builds tags for given column
func (builder GORMTagBuilder) Build(column *Column) string {
	options := []string{}
	options = append(options, fmt.Sprintf("column:%s", column.Name))
	options = append(options, fmt.Sprintf("type:%s", column.Type.DBType()))

	if column.Type.IsPrimaryKey {
		options = append(options, "primary_key")
	}

	if column.Type.IsNullable {
		options = append(options, "null")
	} else {
		options = append(options, "not null")
	}

	if size := column.Type.CharMaxLength; size > 0 {
		options = append(options, fmt.Sprintf("size:%d", size))
	}

	if precision := column.Type.Precision; precision > 0 {
		options = append(options, fmt.Sprintf("precision:%d", precision))
	}

	return fmt.Sprintf("gorm:\"%s\"", strings.Join(options, ";"))
}

// JSONTagBuilder builds JSON tags
type JSONTagBuilder struct{}

// Build builds tags for given column
func (builder JSONTagBuilder) Build(column *Column) string {
	return fmt.Sprintf("json:\"%s\"", column.Name)
}

// XMLTagBuilder builds XML tags
type XMLTagBuilder struct{}

// Build builds tags for given column
func (builder XMLTagBuilder) Build(column *Column) string {
	return fmt.Sprintf("xml:\"%s\"", column.Name)
}

// ValidateTagBuilder builds JSON tags
type ValidateTagBuilder struct{}

// Build builds tags for given column
func (builder ValidateTagBuilder) Build(column *Column) string {
	options := []string{}

	if !column.Type.IsNullable {
		options = append(options, "required")
	}

	if size := column.Type.CharMaxLength; size > 0 {
		options = append(options, fmt.Sprintf("max=%d", size))
	}

	if len(options) == 0 {
		options = append(options, "-")
	}

	return fmt.Sprintf("validate:\"%s\"", strings.Join(options, ","))
}
