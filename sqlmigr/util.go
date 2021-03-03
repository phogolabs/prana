package sqlmigr

import "github.com/jmoiron/sqlx"

// RunAll runs all sqlmigrs
func RunAll(db *sqlx.DB, storage FileSystem) error {
	executor := &Executor{
		Provider: &Provider{
			FileSystem: storage,
			DB:         db,
		},
		Runner: &Runner{
			FileSystem: storage,
			DB:         db,
		},
	}

	_, err := executor.RunAll()
	return err
}
