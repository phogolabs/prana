package migration

import "github.com/jmoiron/sqlx"

// RunAll runs all migrations
func RunAll(db *sqlx.DB, fileSystem FileSystem, dir string) error {
	executor := &Executor{
		Provider: &Provider{
			Dir:        dir,
			FileSystem: fileSystem,
			DB:         db,
		},
		Runner: &Runner{
			Dir:        dir,
			FileSystem: fileSystem,
			DB:         db,
		},
		Generator: &Generator{
			Dir:        dir,
			FileSystem: fileSystem,
		},
	}

	if err := executor.Setup(); err != nil {
		return err
	}

	_, err := executor.RunAll()
	return err
}
