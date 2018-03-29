package migration

import (
	"os"
	"path/filepath"
	"time"

	"github.com/svett/gom"
	"github.com/svett/gom/script"
)

type Runner struct {
	Dir     string
	Gateway *gom.Gateway
}

func (r *Runner) Run(m *Item) error {
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

	return nil
}

func (r *Runner) Revert(m *Item) error {
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

	return err
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
