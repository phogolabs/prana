package sqlmigr

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/apex/log"
	"github.com/fatih/color"
	"github.com/gosuri/uitable"
)

// Flog prints the migrations as fields
func Flog(logger log.Interface, migrations []*Migration) {
	for _, m := range migrations {
		status := "pending"
		timestamp := ""

		if !m.CreatedAt.IsZero() {
			status = "executed"
			timestamp = m.CreatedAt.Format(time.UnixDate)
		}

		fields := log.Fields{
			"Id":          m.ID,
			"Description": m.Description,
			"Status":      status,
			"Drivers":     strings.Join(m.Drivers, ", "),
			"CreatedAt":   timestamp,
		}

		logger.WithFields(fields).Info("Migration")
	}
}

// Ftable prints the migrations as table
func Ftable(w io.Writer, migrations []*Migration) {
	table := uitable.New()
	table.MaxColWidth = 50

	for _, m := range migrations {
		status := color.YellowString("pending")
		timestamp := "--"

		if !m.CreatedAt.IsZero() {
			status = color.GreenString("executed")
			timestamp = m.CreatedAt.Format(time.UnixDate)
		}

		table.AddRow("Id", m.ID)
		table.AddRow("Description", m.Description)
		table.AddRow("Status", status)
		table.AddRow("Drivers", strings.Join(m.Drivers, ", "))
		table.AddRow("Created At", timestamp)
		table.AddRow("")
	}

	fmt.Fprintln(w, table)
}
