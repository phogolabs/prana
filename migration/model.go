package migration

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"
)

//go:generate counterfeiter -fake-name MigrationRunner -o ../fake/MigrationRunner.go . ItemRunner
//go:generate counterfeiter -fake-name MigrationProvider -o ../fake/MigrationProvider.go . ItemProvider
//go:generate counterfeiter -fake-name MigrationGenerator -o ../fake/MigrationGenerator.go . FileGenerator

var (
	format = "20060102150405"
	min    = time.Date(1, time.January, 1970, 0, 0, 0, 0, time.UTC)
)

type ItemRunner interface {
	Run(item *Item) error
	Revert(item *Item) error
}

type ItemProvider interface {
	Migrations() ([]Item, error)
}

type FileGenerator interface {
	Create(m *Item) (string, error)
	Write(m *Item, content *Content) error
}

type Content struct {
	UpCommand   io.Reader
	DownCommand io.Reader
}

type Item struct {
	Id          string    `db:"id"`
	Description string    `db:"description"`
	CreatedAt   time.Time `db:"created_at"`
}

func (m Item) Filename() string {
	return fmt.Sprintf("%s_%s.sql", m.Id, m.Description)
}

func Parse(path string) (*Item, error) {
	name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	parts := strings.Split(name, "_")
	parseErr := fmt.Errorf("Migration '%s' has an invalid file name", path)

	if len(parts) != 2 {
		return nil, parseErr
	}

	if _, err := time.Parse(format, parts[0]); err != nil {
		return nil, parseErr
	}

	return &Item{
		Id:          parts[0],
		Description: parts[1],
	}, nil
}
