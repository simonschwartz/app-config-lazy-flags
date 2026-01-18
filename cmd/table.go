package app

import (
	"github.com/charmbracelet/bubbles/table"
	"github.com/simonschwartz/app-config-lazy-flags/internal/appconfig"
)

// FlagRowData represents a single flag and its state across all environments
type FlagRowData struct {
	FlagName  string
	EnvStates map[string]string // envName -> state (✓/✗/-)
}

// ToTableRow converts the structured data to a table.Row for rendering
func (f FlagRowData) ToTableRow(envOrder []string) table.Row {
	row := table.Row{f.FlagName}
	for _, envName := range envOrder {
		state, exists := f.EnvStates[envName]
		if !exists {
			state = "-"
		}
		row = append(row, state)
	}
	return row
}

// GetEnvState returns the state for a specific environment
func (f FlagRowData) GetEnvState(envName string) string {
	if state, ok := f.EnvStates[envName]; ok {
		return state
	}
	return "-"
}

// FlagsTableData holds both the structured data and rendering info
type FlagsTableData struct {
	Flags    []FlagRowData
	EnvOrder []string
}

// ToTableRows converts all flag data to table rows
func (d FlagsTableData) ToTableRows() []table.Row {
	rows := make([]table.Row, 0, len(d.Flags))
	for _, flag := range d.Flags {
		rows = append(rows, flag.ToTableRow(d.EnvOrder))
	}
	return rows
}


func RenderFlagsTable(flags []appconfig.Result) (table.Model, FlagsTableData) {
	columns := []table.Column{
		{Title: "Flag Name", Width: 20},
	}

	// Build columns and track environment order
	envOrder := make([]string, 0, len(flags))
	for _, flag := range flags {
		envOrder = append(envOrder, flag.EnvName)
		columns = append(columns, table.Column{
			Title: flag.EnvName,
			Width: 15,
		})
	}

	tableData := pivotResults(flags, envOrder)

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(tableData.ToTableRows()),
		table.WithFocused(true),
	)

	return t, tableData
}
