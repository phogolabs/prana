package schema

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
)

var (
	_ Provider = &PostgreSQLProvider{}
	_ Provider = &MySQLProvider{}
	_ Provider = &SQLiteProvider{}
)

// PostgreSQLProvider represents a metadata provider for PostgreSQL
type PostgreSQLProvider struct {
	// DB is a connection to PostgreSQL database
	DB *sqlx.DB
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
	defer rows.Close()

	for rows.Next() {
		table := ""

		if err := rows.Scan(&table); err != nil {
			return []string{}, err
		}

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
	query.WriteString("FROM INFORMATION_SCHEMA.columns ")
	query.WriteString("WHERE table_schema = $1 AND table_name = $2 ")
	query.WriteString("ORDER BY table_schema, table_name, ordinal_position")

	tables := []Table{}
	for _, name := range names {
		table := Table{
			Name: name,
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

			if err := rows.Scan(fields...); err != nil {
				return nil, err
			}

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
		Tables:    tables,
		IsDefault: schema == "public",
	}

	return schemaDef, nil
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
	name := sanitize(columnType.Name)

	switch name {
	case "hstore":
		return HStoreDef.As(nullable)
	default:
		return StringDef.As(nullable)
	}
}

type sqliteInf struct {
	CID          int
	Type         string
	NotNullable  int
	DefaultValue interface{}
	PK           int
}

// SQLiteProvider represents a metadata provider for SQLite
type SQLiteProvider struct {
	// DB is a connection to PostgreSQL database
	DB *sqlx.DB
}

// Tables returns all tables for this schema
func (m *SQLiteProvider) Tables(schema string) ([]string, error) {
	tables := []string{}

	rows, err := m.DB.Query("SELECT DISTINCT tbl_name FROM sqlite_master ORDER BY tbl_name")
	if err != nil {
		return tables, err
	}
	defer rows.Close()

	for rows.Next() {
		table := ""

		if err := rows.Scan(&table); err != nil {
			return []string{}, err
		}

		tables = append(tables, table)
	}

	return tables, nil
}

// Schema returns the schema definition
func (m *SQLiteProvider) Schema(schema string, names ...string) (*Schema, error) {
	tables := []Table{}

	for _, name := range names {
		table := Table{
			Name: name,
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

			if err := rows.Scan(fields...); err != nil {
				return nil, err
			}

			if info.Type == "" {
				return nil, fmt.Errorf("SQLite does not provide any information for column '%s' in table '%s'", column.Name, name)
			}

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

	if matches := pattern.FindStringSubmatch(info.Type); len(matches) > 0 {
		info.Type = matches[1]

		switch len(matches) {
		case 3:
			max, _ = strconv.Atoi(matches[2])
		case 4:
			precision, _ = strconv.Atoi(matches[2])
			precisionScale, _ = strconv.Atoi(matches[3])
		}
	}

	columnType := ColumnType{
		Name:           info.Type,
		Underlying:     info.Type,
		IsNullable:     info.NotNullable == 0,
		CharMaxLength:  max,
		Precision:      precision,
		PrecisionScale: precisionScale,
	}

	return columnType
}

// MySQLProvider represents a metadata provider for MySQL
type MySQLProvider struct {
	// DB is a connection to PostgreSQL database
	DB *sqlx.DB
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
	defer rows.Close()

	for rows.Next() {
		table := ""

		if err := rows.Scan(&table); err != nil {
			return []string{}, err
		}

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
	query.WriteString("SELECT column_name, data_type, data_type, is_nullable = 'YES' AS is_nullable, ")
	query.WriteString("INSTR(column_type, 'unsigned') > 0 AS is_unsigned, ")
	query.WriteString("CASE WHEN numeric_precision IS NULL THEN 0 ELSE numeric_precision END, ")
	query.WriteString("CASE WHEN numeric_scale IS NULL THEN 0 ELSE numeric_scale END, ")
	query.WriteString("CASE WHEN character_maximum_length IS NULL THEN 0 ELSE character_maximum_length END ")
	query.WriteString("FROM INFORMATION_SCHEMA.columns ")
	query.WriteString("WHERE table_schema = ? AND table_name = ? ")
	query.WriteString("ORDER BY table_schema, table_name, ordinal_position")

	tables := []Table{}
	for _, name := range names {
		table := Table{
			Name: name,
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

			if err := rows.Scan(fields...); err != nil {
				return nil, err
			}

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

func sanitize(name string) string {
	return strings.Replace(strings.ToLower(name), `"`, "", -1)
}

func translate(columnType *ColumnType) string {
	nullable := columnType.IsNullable
	name := strings.Replace(strings.ToLower(columnType.Name), `"`, "", -1)

	if columnType.IsUnsigned {
		switch name {
		case "tinyint":
			switch columnType.Precision {
			case 1:
				return BoolDef.As(nullable)
			default:
				return UInt8Def.As(nullable)
			}
		case "smallint":
			return UInt16Def.As(nullable)
		case "mediumint":
			return UInt32Def.As(nullable)
		case "int", "integer":
			return UIntDef.As(nullable)
		case "bigint":
			return UInt64Def.As(nullable)
		}
	}

	switch name {
	case "tinyint":
		switch columnType.Precision {
		case 1:
			return BoolDef.As(nullable)
		default:
			return Int8Def.As(nullable)
		}
	case "mediumint":
		return Int32Def.As(nullable)
	case "binary", "varbinary", "tinyblob", "blob", "mediumblob", "longblob":
		return ByteDef.As(nullable)
	case "bigint", "bigserial":
		return Int64Def.As(nullable)
	case "int", "integer", "serial":
		return IntDef.As(nullable)
	case "smallint", "smallserial":
		return Int16Def.As(nullable)
	case "decimal", "numeric", "double precision":
		return Float64Def.As(nullable)
	case "real":
		return Float32Def.As(nullable)
	case "bit", "interval", "uuint", "bit varying", "character", "money", "character varying", "cidr", "inet", "macaddr", "text", "xml":
		return StringDef.As(nullable)
	case "char":
		return ByteDef.As(nullable)
	case "json", "jsonb":
		return JSONDef.As(nullable)
	case "bytea":
		return ByteSliceDef.As(nullable)
	case "boolean":
		return BoolDef.As(nullable)
	case "date", "time", "datetime", "timestamp", "timestamp without time zone", "timestamp with time zone", "time without time zone", "time with time zone":
		return TimeDef.As(nullable)
	case "uuid":
		return UUIDDef.As(nullable)
	default:
		return StringDef.As(nullable)
	}
}
