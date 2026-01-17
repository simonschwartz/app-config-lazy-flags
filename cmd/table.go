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

func pivotResults(results []appconfig.Result, envOrder []string) FlagsTableData {
	// Build lookup map: flagName -> envName -> enabled
	flagStates := make(map[string]map[string]bool)

	for _, result := range results {
		for flagName, flag := range result.Flags {
			if flagStates[flagName] == nil {
				flagStates[flagName] = make(map[string]bool)
			}
			flagStates[flagName][result.EnvName] = flag.Enabled
		}
	}

	// Build structured flag data
	flags := make([]FlagRowData, 0, len(flagStates))
	for flagName, envMap := range flagStates {
		flagRow := FlagRowData{
			FlagName:  flagName,
			EnvStates: make(map[string]string),
		}

		for _, envName := range envOrder {
			enabled, exists := envMap[envName]
			if !exists {
				flagRow.EnvStates[envName] = "-"
			} else if enabled {
				flagRow.EnvStates[envName] = "✓"
			} else {
				flagRow.EnvStates[envName] = "✗"
			}
		}

		flags = append(flags, flagRow)
	}

	return FlagsTableData{
		Flags:    flags,
		EnvOrder: envOrder,
	}
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
