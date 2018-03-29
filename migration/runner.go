package migration

import (
	"os"
	"path/filepath"
	"time"

	"github.com/apex/log"
	"github.com/phogolabs/gom"
	"github.com/phogolabs/gom/script"
)

// Runner runs or reverts a given migration item.
type Runner struct {
	// Dir represents the project directory.
	Dir string
	// Gateway is a client to underlying database.
	Gateway *gom.Gateway
	// Logger logs the runner exection
	Logger log.Interface
}

// Run runs a given migration item.
func (r *Runner) Run(m *Item) error {
	r.log("Running migration '%s'", m.Filename())

	p, err := r.provider(m)
	if err != nil {
		return err
	}

	cmd, err := p.Command("up")
	if err != nil {
		return err
	}

	if _, err := r.Gateway.Exec(cmd); err != nil {
		return err
	}

	m.CreatedAt = time.Now()

	query := gom.Insert("migrations").
		Set(
			gom.Pair("id", m.Id),
			gom.Pair("description", m.Description),
			gom.Pair("created_at", m.CreatedAt),
		)

	if _, err := r.Gateway.Exec(query); err != nil {
		return err
	}

	r.log("Running migration '%s' completed successfully", m.Filename())
	return nil
}

// Revert reverts a given migration item.
func (r *Runner) Revert(m *Item) error {
	r.log("Reverting migration '%s'", m.Filename())

	p, err := r.provider(m)
	if err != nil {
		return err
	}

	cmd, err := p.Command("down")
	if err != nil {
		return err
	}

	if _, err := r.Gateway.Exec(cmd); err != nil {
		return err
	}

	query := gom.Delete("migrations").Where(gom.Condition("id").Equal(m.Id))

	if _, err := r.Gateway.Exec(query); err != nil {
		return err
	}

	r.log("Reverting migration '%s' completed successfully", m.Filename())
	return nil
}

func (r *Runner) provider(m *Item) (*script.Provider, error) {
	provider := &script.Provider{}

	path := filepath.Join(r.Dir, m.Filename())
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer func() {
		if ioErr := file.Close(); err == nil {
			err = ioErr
		}
	}()

	if err = provider.Load(file); err != nil {
		return nil, err
	}

	return provider, nil
}

func (r *Runner) log(text string, param ...interface{}) {
	if r.Logger != nil {
		r.Logger.Infof(text, param...)
	}
}
