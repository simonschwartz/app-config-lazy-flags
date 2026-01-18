package app

import (
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/simonschwartz/app-config-lazy-flags/internal/appconfig"
)

const nordfoxBlue = lipgloss.Color("#8cafd2")

const (
	Title = "Feature Flags"
)

type FlagsTable struct {
	height     int
	minWidth   int
	tableWidth int
	model      table.Model
	data       FlagsTableData
}

func NewFlagsTable(height int, minWidth int, flags []appconfig.Result) *FlagsTable {
	ft := &FlagsTable{
		height:   height,
		minWidth: minWidth,
	}
	ft.buildTable(flags)
	return ft
}

func (t *FlagsTable) buildTable(flags []appconfig.Result) {
	columns := []table.Column{
		{Title: "Flag Name", Width: 20},
	}

	envOrder := make([]string, 0, len(flags))
	for _, flag := range flags {
		envOrder = append(envOrder, flag.EnvName)
		columns = append(columns, table.Column{
			Title: flag.EnvName,
			Width: 15,
		})
	}

	// Calculate total width from columns (+ padding between columns)
	t.tableWidth = 0
	for _, col := range columns {
		t.tableWidth += col.Width + 2 // +2 for cell padding
	}
	if t.tableWidth < t.minWidth {
		t.tableWidth = t.minWidth
	}

	t.data = pivotResults(flags, envOrder)

	s := table.DefaultStyles()
	s.Header = lipgloss.NewStyle().
		Bold(true).
		Padding(0, 1)
	s.Selected = lipgloss.NewStyle().
		Foreground(nordfoxBlue).
		Bold(true)
	s.Cell = lipgloss.NewStyle().
		Padding(0, 1)

	t.model = table.New(
		table.WithColumns(columns),
		table.WithRows(t.data.ToTableRows()),
		table.WithFocused(true),
		table.WithStyles(s),
	)
}

func (t *FlagsTable) SetData(flags []appconfig.Result) tea.Cmd {
	t.buildTable(flags)
	return nil
}

func (t *FlagsTable) Render() string {
	if len(t.data.Flags) == 0 {
		msg := "You have no flags"
		paddedMsg := msg + strings.Repeat("\n", t.height-1)
		return RenderPanel(paddedMsg, Title, t.minWidth)
	}

	// Shift table content 1 space left to align with panel title
	view := t.model.View()
	lines := strings.Split(view, "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, " ") {
			lines[i] = line[1:]
		} else if strings.HasPrefix(line, "\x1b") {
			// Selected row starts with ANSI escape code - find and remove space after it
			// ANSI codes are like \x1b[...m
			endIdx := strings.Index(line, "m")
			if endIdx != -1 && endIdx+1 < len(line) && line[endIdx+1] == ' ' {
				lines[i] = line[:endIdx+1] + line[endIdx+2:]
			}
		}
	}
	return RenderPanel(strings.Join(lines, "\n"), Title, t.tableWidth+7)
}

func (t *FlagsTable) HandleMsg(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	t.model, cmd = t.model.Update(msg)
	return cmd
}

func (t *FlagsTable) Cursor() int {
	return t.model.Cursor()
}

func (t *FlagsTable) Data() FlagsTableData {
	return t.data
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
				flagRow.EnvStates[envName] = "on"
			} else {
				flagRow.EnvStates[envName] = "off"
			}
		}

		flags = append(flags, flagRow)
	}

	// Sort flags alphabetically for consistent ordering
	sort.Slice(flags, func(i, j int) bool {
		return flags[i].FlagName < flags[j].FlagName
	})

	return FlagsTableData{
		Flags:    flags,
		EnvOrder: envOrder,
	}
}
