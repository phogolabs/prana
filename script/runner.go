package script

import (
	"github.com/svett/gom"
)

type Runner struct {
	Dir     string
	Gateway *gom.Gateway
}

func (r *Runner) Run(name string, args ...gom.Param) (*gom.Rows, error) {
	provider := &gom.CmdProvider{
		Repository: make(map[string]string),
	}

	if err := provider.LoadDir(r.Dir); err != nil {
		return nil, err
	}

	cmd, err := provider.Command(name, args...)
	if err != nil {
		return nil, err
	}

	return r.Gateway.Query(cmd)
}
