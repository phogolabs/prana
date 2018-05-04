// Package prana facilitates the work with applications that use database for
// their store
package prana

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/go-sql-driver/mysql"
)

// ParseURL parses a URL and returns the database driver and connection string to the database
func ParseURL(conn string) (string, string, error) {
	uri, err := url.Parse(conn)
	if err != nil {
		return "", "", err
	}

	driver := strings.ToLower(uri.Scheme)

	switch driver {
	case "mysql":
		source, err := parseMySQL(driver, conn)
		if err != nil {
			return "", "", nil
		}
		return driver, source, nil
	case "sqlite3":
		source := strings.Replace(conn, fmt.Sprintf("%s://", driver), "", -1)
		return driver, source, nil
	default:
		return driver, conn, nil
	}
}

func parseMySQL(driver, conn string) (string, error) {
	source := strings.Replace(conn, fmt.Sprintf("%s://", driver), "", -1)

	cfg, err := mysql.ParseDSN(source)
	if err != nil {
		return "", err
	}
	cfg.ParseTime = true

	return cfg.FormatDSN(), nil
}
