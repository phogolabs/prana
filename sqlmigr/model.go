// Package sqlmigr provides primitives and functions to work with SQL
// sqlmigrs.
package sqlmigr

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/phogolabs/parcello"
)

//go:generate counterfeiter -fake-name MigrationRunner -o ../fake/MigrationRunner.go . ItemRunner
//go:generate counterfeiter -fake-name MigrationProvider -o ../fake/MigrationProvider.go . ItemProvider
//go:generate counterfeiter -fake-name MigrationGenerator -o ../fake/MigrationGenerator.go . ItemGenerator

var (
	format = "20060102150405"
	min    = time.Date(1, time.January, 1970, 0, 0, 0, 0, time.UTC)
)

// FileSystem provides with primitives to work with the underlying file system
type FileSystem = parcello.FileSystem

// ItemRunner runs or reverts a given sqlmigr item.
type ItemRunner interface {
	// Run runs a given sqlmigr item.
	Run(item *Item) error
	// Revert reverts a given sqlmigr item.
	Revert(item *Item) error
}

// ItemProvider provides all items.
type ItemProvider interface {
	// Migrations returns all sqlmigr items.
	Migrations() ([]Item, error)
	// Insert inserts executed sqlmigr item in the sqlmigrs table.
	Insert(item *Item) error
	// Delete deletes applied sqlmigr item from sqlmigrs table.
	Delete(item *Item) error
	// Exists returns true if the sqlmigr exists
	Exists(item *Item) bool
}

// ItemGenerator generates a sqlmigr item file.
type ItemGenerator interface {
	// Create creates a new sqlmigr.
	Create(m *Item) error
	// Write creates a new sqlmigr for given content.
	Write(m *Item, content *Content) error
}

// Content represents a sqlmigr content.
type Content struct {
	// UpCommand is the content for upgrade operation.
	UpCommand io.Reader
	// DownCommand is the content for rollback operation.
	DownCommand io.Reader
}

// Item represents a single sqlmigr.
type Item struct {
	// Id is the primary key for this sqlmigr
	ID string `db:"id"`
	// Description is the short description of this sqlmigr.
	Description string `db:"description"`
	// CreatedAt returns the time of sqlmigr execution.
	CreatedAt time.Time `db:"created_at"`
}

// Filename returns the item filename
func (m Item) Filename() string {
	return fmt.Sprintf("%s_%s.sql", m.ID, m.Description)
}

// Parse parses a given file path to a sqlmigr item.
func Parse(path string) (*Item, error) {
	name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	parts := strings.SplitN(name, "_", 2)
	parseErr := fmt.Errorf("Migration '%s' has an invalid file name", path)

	if len(parts) != 2 {
		return nil, parseErr
	}

	if _, err := time.Parse(format, parts[0]); err != nil {
		return nil, parseErr
	}

	return &Item{
		ID:          parts[0],
		Description: parts[1],
	}, nil
}
