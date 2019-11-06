// Package prana facilitates the work with applications that use database for
// their store
package prana

import (
	"errors"
	"regexp"
	"strings"

	"github.com/go-sql-driver/mysql"
	"github.com/phogolabs/log"
)

//go:generate counterfeiter -fake-name Logger -o ./fake/logger.go . Logger

// Logger used to log any output
type Logger = log.Logger

const (
	MYSQL_DRIVER    = "mysql"
	SQLITE_DRIVER   = "sqlite3"
	POSTGRES_DRIVER = "postgres"
)

var (
	errNoDriverName = errors.New("No driver name")
	errEmptyConnURL = errors.New("URL cannot be empty")
	errInvalidDSN   = errors.New("Invalid DSN")
)

// ParseURL parses a URL and returns the database driver and connection string to the database
func ParseURL(conn string) (string, string, error) {
	driver, source, err := parseRawURL(conn)
	if err != nil {
		return "", "", err
	}

	switch driver {
	case MYSQL_DRIVER:
		mysqlSource, err := parseMySQL(driver, source)
		if err != nil {
			return "", "", err
		}
		return driver, mysqlSource, nil
	case SQLITE_DRIVER:
		return driver, source, nil
	case POSTGRES_DRIVER:
		return driver, conn, nil
	default:
		return driver, conn, nil
	}
}

// parseRawURL returns the db driver name from a URL string
func parseRawURL(url string) (driverName string, path string, err error) {
	if url == "" {
		return "", "", errEmptyConnURL
	}

	// scheme must match
	prog := regexp.MustCompile(`^([a-zA-Z][a-zA-Z0-9+-.]*)://(.*)$`)
	matches := prog.FindStringSubmatch(url)

	if len(matches) > 2 {
		return strings.ToLower(matches[1]), matches[2], nil
	}
	return "", "", errInvalidDSN
}

func parseMySQL(driver, source string) (string, error) {
	cfg, err := mysql.ParseDSN(source)
	if err != nil {
		return "", err
	}
	cfg.ParseTime = true

	return cfg.FormatDSN(), nil
}
