package app

import (
	"github.com/charmbracelet/bubbles/table"
	"github.com/simonschwartz/app-config-lazy-flags/internal/appconfig"
)

func pivotResults(results []appconfig.Result, envOrder []string) []table.Row {
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

	// Build rows with flag name + state per environment
	rows := make([]table.Row, 0, len(flagStates))
	for flagName, envMap := range flagStates {
		row := table.Row{flagName}
		for _, envName := range envOrder {
			enabled, exists := envMap[envName]
			if !exists {
				row = append(row, "-")
			} else if enabled {
				row = append(row, "✓")
			} else {
				row = append(row, "✗")
			}
		}
		rows = append(rows, row)
	}

	return rows
}

func RenderFlagsTable(flags []appconfig.Result) table.Model {
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

	rows := pivotResults(flags, envOrder)

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
	)

	return t
}
