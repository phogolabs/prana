package sqlmigr

import (
	"io"
	"time"

	"github.com/apex/log"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
)

// Flog prints the migrations as fields
func Flog(logger log.Interface, migrations []Migration) {
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
			"CreatedAt":   timestamp,
		}

		logger.WithFields(fields).Info("Migration")
	}
}

// Ftable prints the migrations as table
func Ftable(w io.Writer, migrations []Migration) {
	table := tablewriter.NewWriter(w)
	table.SetHeader([]string{"Id", "Description", "Status", "Created At"})

	for _, m := range migrations {
		status := color.YellowString("pending")
		timestamp := ""

		if !m.CreatedAt.IsZero() {
			status = color.GreenString("executed")
			timestamp = m.CreatedAt.Format(time.UnixDate)
		}

		row := []string{m.ID, m.Description, status, timestamp}
		table.Append(row)
	}

	table.Render()
}
