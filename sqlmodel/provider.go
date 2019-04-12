package sqlmodel

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-openapi/inflect"
)

var (
	_ SchemaProvider = &ModelProvider{}
	_ SchemaProvider = &PostgreSQLProvider{}
	_ SchemaProvider = &MySQLProvider{}
	_ SchemaProvider = &SQLiteProvider{}
)

// ModelProviderConfig is the ModelProvider's config
type ModelProviderConfig struct {
	// Package name
	Package string
	// UseNamedParams determines whether to use named params
	UseNamedParams bool
	// InlcudeDoc determines whether to include documentation
	InlcudeDoc bool
}

// ModelProvider represents the model provider
type ModelProvider struct {
	// Config for this provider
	Config *ModelProviderConfig
	// Provider represents the actual provider
	Provider SchemaProvider
	// TagBuilder builds struct tags from column type
	TagBuilder TagBuilder
}

// Tables returns all tables for this schema
func (m *ModelProvider) Tables(schema string) ([]string, error) {
	return m.Provider.Tables(schema)
}

// Schema returns the schema definition
func (m *ModelProvider) Schema(name string, tables ...string) (*Schema, error) {
	schema, err := m.Provider.Schema(name, tables...)
	if err != nil {
		return nil, err
	}

	schema.Model = SchemaModel{
		Package:          m.Config.Package,
		HasDocumentation: m.Config.InlcudeDoc,
	}

	for index, table := range schema.Tables {
		table.Name = m.tableName(schema, &table)
		table.Model = TableModel{
			HasDocumentation: m.Config.InlcudeDoc,
			Type:             m.typeName(&table),
			Package:          schema.Model.Package,
			PrimaryKey:       make(map[string]string),
		}

		for index, column := range table.Columns {
			column.Model = ColumnModel{
				HasDocumentation: m.Config.InlcudeDoc,
				Name:             m.fieldName(&column),
				Type:             m.fieldType(&column),
				Tag:              m.TagBuilder.Build(&column),
			}

			table.Columns[index] = column
		}

		m.setPrimaryKey(schema, &table)
		m.setSelectAllRoutine(schema, &table)
		m.setSelectByPKRoutine(schema, &table)
		m.setInsertRoutine(schema, &table)
		m.setDeleteByPKRoutine(schema, &table)
		m.setUpdateByPKRoutine(schema, &table)

		schema.Tables[index] = table
	}

	return schema, nil
}

// Close closes connection to the db
func (m *ModelProvider) Close() error {
	return m.Provider.Close()
}

func (m *ModelProvider) setPrimaryKey(schema *Schema, table *Table) {
	var (
		conditions []string
		cond       string
		arguments  []string
		arg        string
		parameters []string
		param      string
	)

	for _, column := range table.Columns {
		if !column.Type.IsPrimaryKey {
			continue
		}

		parameters = append(parameters, fmt.Sprintf("entity.%s", column.Model.Name))

		if m.Config.UseNamedParams {
			cond = fmt.Sprintf("%s = :%s", column.Name, column.Name)
		} else {
			cond = fmt.Sprintf("%s = ?", column.Name)
		}

		param = inflect.CamelizeDownFirst(column.Name)
		table.Model.PrimaryKey[column.Name] = param

		arg = fmt.Sprintf("%s %s", param, column.ScanType)

		arguments = append(arguments, arg)
		conditions = append(conditions, cond)
	}

	table.Model.PrimaryKeyCondition = strings.Join(conditions, " AND ")
	table.Model.PrimaryKeyArgs = strings.Join(arguments, ", ")
	table.Model.PrimaryKeyParams = strings.Join(parameters, ", ")
}

func (m *ModelProvider) setSelectAllRoutine(schema *Schema, table *Table) {
	table.Model.SelectAllRoutine = fmt.Sprintf("select-all-%s", m.commandName(table.Name, false))
}

func (m *ModelProvider) setSelectByPKRoutine(schema *Schema, table *Table) {
	table.Model.SelectByPKRoutine = fmt.Sprintf("select-%s-by-pk", m.commandName(table.Name, true))
}

func (m *ModelProvider) setInsertRoutine(schema *Schema, table *Table) {
	columns, values := m.insertParam(table)

	if len(columns) == 0 || len(values) == 0 {
		return
	}

	table.Model.InsertRoutine = fmt.Sprintf("insert-%s", m.commandName(table.Name, true))
	table.Model.InsertColumns = columns
	table.Model.InsertValues = values
}

func (m *ModelProvider) setUpdateByPKRoutine(schema *Schema, table *Table) {
	table.Model.UpdateByPKRoutine = fmt.Sprintf("update-%s-by-pk", m.commandName(table.Name, true))
	table.Model.UpdateByPKColumns = m.updateColumns(table)
}

func (m *ModelProvider) setDeleteByPKRoutine(schema *Schema, table *Table) {
	table.Model.DeleteByPKRoutine = fmt.Sprintf("delete-%s-by-pk", m.commandName(table.Name, true))
}

func (m *ModelProvider) commandName(name string, singularize bool) string {
	name = strings.Replace(name, ".", "-", -1)
	name = strings.Replace(name, "_", "-", -1)
	if singularize {
		name = inflect.Singularize(name)
	}
	return name
}

func (m *ModelProvider) tableName(schema *Schema, table *Table) string {
	name := table.Name

	if !schema.IsDefault {
		name = fmt.Sprintf("%s.%s", schema.Name, name)
	}

	return name
}

func (m *ModelProvider) insertParam(table *Table) (string, string) {
	var (
		columns []string
		values  []string
		param   string
	)

	for _, column := range table.Columns {
		columns = append(columns, column.Name)

		if m.Config.UseNamedParams {
			param = fmt.Sprintf(":%s", column.Name)
		} else {
			param = "?"
		}

		values = append(values, param)
	}

	return strings.Join(columns, ", "), strings.Join(values, ", ")
}

func (m *ModelProvider) updateColumns(table *Table) string {
	var (
		values []string
		param  string
	)

	for _, column := range table.Columns {
		if m.Config.UseNamedParams {
			param = fmt.Sprintf("%s = :%s", column.Name, column.Name)
		} else {
			param = fmt.Sprintf("%s = ?", column.Name)
		}

		if !column.Type.IsPrimaryKey {
			values = append(values, param)
		}
	}

	return strings.Join(values, ", ")
}

func (m *ModelProvider) typeName(table *Table) string {
	name := strings.Replace(table.Name, ".", " ", -1)
	name = inflect.Camelize(name)
	name = inflect.Singularize(name)
	return name
}

func (m *ModelProvider) fieldName(column *Column) string {
	name := inflect.Camelize(column.Name)
	name = strings.Replace(name, "Id", "ID", -1)
	return name
}

func (m *ModelProvider) fieldType(column *Column) string {
	return column.ScanType
}

// PostgreSQLProvider represents a metadata provider for PostgreSQL
type PostgreSQLProvider struct {
	// DB is a connection to PostgreSQL database
	DB Querier
}

// Close closes connection to the db
func (m *PostgreSQLProvider) Close() error {
	return m.DB.Close()
}

// Tables returns all tables for this schema
func (m *PostgreSQLProvider) Tables(schema string) ([]string, error) {
	schema = m.nameOf(schema)
	tables := []string{}

	query := &bytes.Buffer{}
	query.WriteString("SELECT table_name FROM information_schema.tables ")
	query.WriteString("WHERE table_schema = $1 ")
	query.WriteString("ORDER BY table_name")

	rows, err := m.DB.Query(query.String(), schema)
	if err != nil {
		return tables, err
	}

	defer func() {
		if ioErr := rows.Close(); err == nil {
			err = ioErr
		}
	}()

	for rows.Next() {
		table := ""

		_ = rows.Scan(&table)
		tables = append(tables, table)
	}

	return tables, nil
}

// Schema returns the schema definition
func (m *PostgreSQLProvider) Schema(schema string, names ...string) (*Schema, error) {
	schema = m.nameOf(schema)

	query := &bytes.Buffer{}
	query.WriteString("SELECT column_name, data_type, udt_name, is_nullable = 'YES' AS is_nullable, ")
	query.WriteString("CASE WHEN numeric_precision IS NULL THEN 0 ELSE numeric_precision END, ")
	query.WriteString("CASE WHEN numeric_scale IS NULL THEN 0 ELSE numeric_scale END, ")
	query.WriteString("CASE WHEN character_maximum_length IS NULL THEN 0 ELSE character_maximum_length END ")
	query.WriteString("FROM information_schema.columns ")
	query.WriteString("WHERE table_schema = $1 AND table_name = $2 ")
	query.WriteString("ORDER BY table_schema, table_name, ordinal_position")

	tables := []Table{}
	for _, name := range names {
		primaryKey, err := m.primaryKey(schema, name)
		if err != nil {
			return nil, err
		}

		table := Table{
			Name:   name,
			Driver: "postgresql",
		}

		rows, err := m.DB.Query(query.String(), schema, name)
		if err != nil {
			return nil, err
		}

		for rows.Next() {
			column := Column{}

			fields := []interface{}{
				&column.Name,
				&column.Type.Name,
				&column.Type.Underlying,
				&column.Type.IsNullable,
				&column.Type.Precision,
				&column.Type.PrecisionScale,
				&column.Type.CharMaxLength,
			}

			_ = rows.Scan(fields...)

			column.Type.IsPrimaryKey = contains(primaryKey, column.Name)
			column.ScanType = m.translate(&column.Type)
			table.Columns = append(table.Columns, column)
		}

		tables = append(tables, table)
	}

	if len(tables) == 0 {
		return nil, fmt.Errorf("No tables found")
	}

	schemaDef := &Schema{
		Name:      schema,
		Driver:    "postgresql",
		Tables:    tables,
		IsDefault: schema == "public",
	}

	return schemaDef, nil
}

func (m *PostgreSQLProvider) primaryKey(schema, table string) ([]string, error) {
	query := &bytes.Buffer{}
	query.WriteString("SELECT c.column_name ")
	query.WriteString("FROM information_schema.key_column_usage AS c ")
	query.WriteString("LEFT JOIN information_schema.table_constraints AS t ")
	query.WriteString("ON t.constraint_name = c.constraint_name ")
	query.WriteString("WHERE t.table_schema = $1 AND t.table_name = $2 AND t.constraint_type = 'PRIMARY KEY' ")
	query.WriteString("ORDER BY c.column_name")

	rows, err := m.DB.Query(query.String(), schema, table)
	if err != nil {
		return nil, err
	}

	var (
		columns []string
		column  string
	)

	for rows.Next() {
		_ = rows.Scan(&column)
		columns = append(columns, column)
	}

	return columns, nil
}

func (m *PostgreSQLProvider) nameOf(schema string) string {
	if schema == "" {
		schema = "public"
	}
	return schema
}

func (m *PostgreSQLProvider) translate(columnType *ColumnType) string {
	name := strings.Replace(strings.ToLower(columnType.Name), `"`, "", -1)

	switch name {
	case "user-defined":
		return m.userDefType(columnType)
	default:
		return translate(columnType)
	}
}

func (m *PostgreSQLProvider) userDefType(columnType *ColumnType) string {
	nullable := columnType.IsNullable
	name := sanitize(columnType.Underlying)

	switch name {
	case "hstore":
		return hstoreDef.As(nullable)
	default:
		return stringDef.As(nullable)
	}
}

// SQLiteProvider represents a metadata provider for SQLite
type SQLiteProvider struct {
	// DB is a connection to PostgreSQL database
	DB Querier
}

// Close closes connection to the db
func (m *SQLiteProvider) Close() error {
	return m.DB.Close()
}

// Tables returns all tables for this schema
func (m *SQLiteProvider) Tables(schema string) ([]string, error) {
	tables := []string{}

	rows, err := m.DB.Query("SELECT DISTINCT tbl_name FROM sqlite_master ORDER BY tbl_name")
	if err != nil {
		return tables, err
	}
	defer func() {
		if ioErr := rows.Close(); err == nil {
			err = ioErr
		}
	}()

	for rows.Next() {
		table := ""

		_ = rows.Scan(&table)
		tables = append(tables, table)
	}

	return tables, nil
}

// Schema returns the schema definition
func (m *SQLiteProvider) Schema(schema string, names ...string) (*Schema, error) {
	tables := []Table{}

	for _, name := range names {
		table := Table{
			Name:   name,
			Driver: "sqlite",
		}

		query := fmt.Sprintf("pragma table_info(%s)", name)
		rows, err := m.DB.Query(query)
		if err != nil {
			return nil, err
		}

		for rows.Next() {
			column := Column{}
			info := sqliteInf{}

			fields := []interface{}{
				&info.CID,
				&column.Name,
				&info.Type,
				&info.NotNullable,
				&info.DefaultValue,
				&info.PK,
			}

			_ = rows.Scan(fields...)

			column.Type = m.create(&info)
			column.ScanType = translate(&column.Type)

			table.Columns = append(table.Columns, column)
		}

		tables = append(tables, table)
	}

	if len(tables) == 0 {
		return nil, fmt.Errorf("No tables found")
	}

	schemaDef := &Schema{
		Name:      "default",
		Driver:    "sqlite",
		Tables:    tables,
		IsDefault: true,
	}

	return schemaDef, nil
}

func (m *SQLiteProvider) create(info *sqliteInf) ColumnType {
	pattern := regexp.MustCompile("([a-z\\s]*)\\(([0-9]*),?([0-9]*)\\)")

	var (
		max            int
		precision      int
		precisionScale int
	)

	if matches := pattern.FindStringSubmatch(info.Type); len(matches) >= 4 {
		info.Type = matches[1]

		switch {
		case info.Type == "bit":
			precision, _ = strconv.Atoi(matches[2])
		case strings.TrimSpace(matches[3]) == "":
			max, _ = strconv.Atoi(matches[2])
		default:
			precision, _ = strconv.Atoi(matches[2])
			precisionScale, _ = strconv.Atoi(matches[3])
		}
	}

	columnType := ColumnType{
		Name:           info.Type,
		Underlying:     info.Type,
		IsPrimaryKey:   info.PK == 1,
		IsNullable:     info.NotNullable == 0,
		CharMaxLength:  max,
		Precision:      precision,
		PrecisionScale: precisionScale,
	}

	return columnType
}

// MySQLProvider represents a metadata provider for MySQL
type MySQLProvider struct {
	// DB is a connection to MySQL database
	DB Querier
}

// Close closes connection to the db
func (m *MySQLProvider) Close() error {
	return m.DB.Close()
}

// Tables returns all tables for this schema
func (m *MySQLProvider) Tables(schema string) ([]string, error) {
	var (
		tables []string
		err    error
	)

	if schema == "" {
		if schema, err = m.database(); err != nil {
			return tables, err
		}
	}

	query := &bytes.Buffer{}
	query.WriteString("SELECT table_name FROM information_schema.tables ")
	query.WriteString("WHERE table_schema = ? and table_type = ? ")
	query.WriteString("ORDER BY table_name")

	rows, err := m.DB.Query(query.String(), schema, "BASE TABLE")
	if err != nil {
		return tables, err
	}
	defer func() {
		if ioErr := rows.Close(); err == nil {
			err = ioErr
		}
	}()

	for rows.Next() {
		table := ""

		_ = rows.Scan(&table)
		tables = append(tables, table)
	}

	return tables, nil
}

// Schema returns the schema definition
func (m *MySQLProvider) Schema(schema string, names ...string) (*Schema, error) {
	var (
		err      error
		database string
	)

	if database, err = m.database(); err != nil {
		return nil, err
	}

	if schema == "" {
		schema = database
	}

	query := &bytes.Buffer{}
	query.WriteString("SELECT column_name, data_type, REPLACE(column_type, ' unsigned', ''), is_nullable = 'YES' AS is_nullable, ")
	query.WriteString("INSTR(column_type, 'unsigned') > 0 AS is_unsigned, ")
	query.WriteString("CASE WHEN numeric_precision IS NULL THEN 0 ELSE numeric_precision END, ")
	query.WriteString("CASE WHEN numeric_scale IS NULL THEN 0 ELSE numeric_scale END, ")
	query.WriteString("CASE WHEN character_maximum_length IS NULL THEN 0 ELSE character_maximum_length END ")
	query.WriteString("FROM information_schema.columns ")
	query.WriteString("WHERE table_schema = ? AND table_name = ? ")
	query.WriteString("ORDER BY table_schema, table_name, ordinal_position")

	tables := []Table{}
	for _, name := range names {
		table := Table{
			Name:   name,
			Driver: "mysql",
		}

		primaryKey, err := m.primaryKey(schema, name)
		if err != nil {
			return nil, err
		}

		rows, err := m.DB.Query(query.String(), schema, name)
		if err != nil {
			return nil, err
		}

		for rows.Next() {
			column := Column{}

			fields := []interface{}{
				&column.Name,
				&column.Type.Name,
				&column.Type.Underlying,
				&column.Type.IsNullable,
				&column.Type.IsUnsigned,
				&column.Type.Precision,
				&column.Type.PrecisionScale,
				&column.Type.CharMaxLength,
			}

			_ = rows.Scan(fields...)

			column.Type.IsPrimaryKey = contains(primaryKey, column.Name)
			column.ScanType = translate(&column.Type)
			table.Columns = append(table.Columns, column)
		}

		tables = append(tables, table)
	}

	if len(tables) == 0 {
		return nil, fmt.Errorf("No tables found")
	}

	schemaDef := &Schema{
		Name:      schema,
		Driver:    "mysql",
		Tables:    tables,
		IsDefault: schema == database,
	}

	return schemaDef, nil
}

func (m *MySQLProvider) database() (string, error) {
	schema := ""
	row := m.DB.QueryRow("SELECT database()")

	if err := row.Scan(&schema); err != nil {
		return "", err
	}

	return schema, nil
}

func (m *MySQLProvider) primaryKey(schema, table string) ([]string, error) {
	query := &bytes.Buffer{}
	query.WriteString("SELECT c.column_name ")
	query.WriteString("FROM information_schema.key_column_usage AS c ")
	query.WriteString("INNER JOIN information_schema.table_constraints AS t ")
	query.WriteString("ON t.constraint_name = c.constraint_name ")
	query.WriteString("WHERE c.table_schema = ? AND c.table_name = ? AND ")
	query.WriteString("t.table_schema = ? AND t.table_name = ? AND t.constraint_type = 'PRIMARY KEY' ")
	query.WriteString("ORDER BY c.column_name")

	rows, err := m.DB.Query(query.String(), schema, table, schema, table)
	if err != nil {
		return nil, err
	}

	columns := []string{}

	for rows.Next() {
		column := ""

		_ = rows.Scan(&column)
		columns = append(columns, column)
	}

	return columns, nil
}

func sanitize(name string) string {
	return strings.Replace(strings.ToLower(name), `"`, "", -1)
}

// nolint: gocyclo
func translate(columnType *ColumnType) string {
	nullable := columnType.IsNullable
	name := strings.Replace(strings.ToLower(columnType.Name), `"`, "", -1)

	if columnType.IsUnsigned {
		switch name {
		case "tinyint":
			switch columnType.Underlying {
			case "tinyint(1)":
				return boolDef.As(nullable)
			default:
				return uint8Def.As(nullable)
			}
		case "smallint":
			return uint16Def.As(nullable)
		case "mediumint":
			return uint32Def.As(nullable)
		case "int", "integer":
			return uintDef.As(nullable)
		case "bigint":
			return uint64Def.As(nullable)
		}
	}

	switch name {
	case "tinyint":
		switch columnType.Underlying {
		case "tinyint(1)":
			return boolDef.As(nullable)
		default:
			return int8Def.As(nullable)
		}
	case "mediumint":
		return int32Def.As(nullable)
	case "binary", "varbinary", "tinyblob", "blob", "mediumblob", "longblob":
		return byteDef.As(nullable)
	case "bigint", "bigserial":
		return int64Def.As(nullable)
	case "int", "integer", "serial":
		return intDef.As(nullable)
	case "smallint", "smallserial":
		return int16Def.As(nullable)
	case "decimal", "numeric", "double precision":
		return float64Def.As(nullable)
	case "real":
		return float32Def.As(nullable)
	case "bit", "interval", "uuint", "bit varying", "character", "money", "character varying", "cidr", "inet", "macaddr", "text", "xml":
		return stringDef.As(nullable)
	case "char":
		return byteDef.As(nullable)
	case "json", "jsonb":
		return jsonDef.As(nullable)
	case "bytea":
		return byteSliceDef.As(nullable)
	case "boolean":
		return boolDef.As(nullable)
	case "abstime", "date", "time", "datetime", "timestamp", "timestamp without time zone", "timestamp with time zone", "time without time zone", "time with time zone":
		return timeDef.As(nullable)
	case "uuid":
		return uuidDef.As(nullable)
	default:
		return stringDef.As(nullable)
	}
}
