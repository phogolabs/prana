// Package prana facilitates the work with applications that use database for
// their store
package prana

import (
	"fmt"
	"net/url"
	"strings"
)

// ParseURL parses a URL and returns the database driver and connection string to the database
func ParseURL(conn string) (string, string, error) {
	uri, err := url.Parse(conn)
	if err != nil {
		return "", "", err
	}

	driver := strings.ToLower(uri.Scheme)

	switch driver {
	case "mysql", "sqlite3":
		source := strings.Replace(conn, fmt.Sprintf("%s://", driver), "", -1)
		return driver, source, nil
	default:
		return driver, conn, nil
	}
}
