// Package app provides a TUI for managing AWS AppConfig feature flags.
//
// AppConfig Data Hierarchy:
//   - Application (e.g., "Wordle")
//     └── Configuration Profile (e.g., "WebFeatureFlags", "APIFeatureFlags", "iOSAppFeatureFlags")
//     └── Environment (e.g., "development", "production")
//     └── Feature Flags (key-value pairs with enabled/disabled state)
//
// UI Flow:
//  1. Select Application
//  2. Select Configuration Profile
//  3. View flag matrix (flags × environments)
package app

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/simonschwartz/app-config-lazy-flags/internal/appconfig"
	"github.com/simonschwartz/app-config-lazy-flags/internal/filecache"
)

type view int

const (
	// initial screen displaying list of app config apps user has access to
	appList view = iota

	// displays list of feature flag configs for selected app
	configList

	// displays all feature flags and their state per environment
	flagsTable

	// detail view for modifying a single flag's state across environments
	flagDetail
)

type Model struct {
	appconfigClient appconfig.Client
	filecache       filecache.Cache

	activeView view

	appsPanel      *ListPanel
	appsPanelError string

	configsPanel      *ListPanel
	configsCache      map[string]configsLoader
	configsPanelError string

	flagsTable      *FlagsTable
	flagsTableError string

	flagDetail *FlagDetail
	// Flag detail view state
	selectedFlagIdx int // which flag in flagsData.Flags is selected (-1 = none)
	selectedEnvIdx  int // which environment is highlighted in detail view
}

func NewModel(appconfigClient *appconfig.Client, filecache *filecache.Cache) Model {
	appsPanel := NewAppsPanel(20, 50, []appconfig.App{})
	configsPanel := NewConfigsPanel(20, 50, []appconfig.AppFlagConfig{})
	flagsTable := NewFlagsTable(20, 50, []appconfig.Result{})
	flagDetail := NewFlagDetail(flagsTable.Render)

	return Model{
		appconfigClient: *appconfigClient,
		configsCache:    make(map[string]configsLoader),
		filecache:       *filecache,
		activeView:      appList,
		selectedFlagIdx: -1,
		selectedEnvIdx:  0,
		appsPanel:       appsPanel,
		configsPanel:    configsPanel,
		flagsTable:      flagsTable,
		flagDetail:      flagDetail,
	}
}

func (m Model) Init() tea.Cmd {
	return m.loadAppsCmd()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case appsLoader:
		if msg.err != nil {
			m.appsPanelError = fmt.Sprintf("Error: %v", msg.err)
			return m, nil
		}

		m.appsPanelError = ""
		var appItems []list.Item
		for _, app := range msg.apps {
			appItems = append(appItems, AppItem(app))
		}
		cmd := m.appsPanel.SetItems(appItems)
		return m, cmd
	case configsLoader:
		m.activeView = configList
		if msg.err != nil {
			m.configsPanelError = fmt.Sprintf("Error: %v", msg.err)
			return m, nil
		}
		m.configsPanelError = ""
		var configItems []list.Item
		for _, config := range msg.configs {
			configItems = append(configItems, ConfigItem(config))
		}
		cmd := m.configsPanel.SetItems(configItems)
		m.configsCache[msg.appId] = msg
		return m, cmd
	case flagsLoader:
		m.activeView = flagsTable
		if msg.err != nil {
			m.flagsTableError = fmt.Sprintf("Error: %v", msg.err)
			return m, nil
		}
		cmd := m.flagsTable.SetData(msg.flags)
		return m, cmd

	case tea.KeyMsg:
		switch msg.String() {
		// allow user to go back to previous view
		case "esc":
			switch m.activeView {
			case configList:
				m.activeView = appList
			case flagsTable:
				m.activeView = configList
			case flagDetail:
				// If confirming, cancel and stay in detail view
				if m.flagDetail.IsConfirming() {
					m.flagDetail.CancelConfirm()
					return m, nil
				}
				m.activeView = flagsTable
				m.selectedFlagIdx = -1
			}
			return m, nil
		case "ctrl+c", "q":
			return m, tea.Quit
		case "enter":
			if m.activeView == appList {
				if item, ok := m.appsPanel.SelectedItem(); ok {
					if app, ok := item.(AppItem); ok {
						return m, m.loadConfigsCmd(*app.Id)
					}
				}
			}

			if m.activeView == configList {
				if item, ok := m.configsPanel.SelectedItem(); ok {
					if config, ok := item.(ConfigItem); ok {
						if appItem, ok := m.appsPanel.SelectedItem(); ok {
							if app, ok := appItem.(AppItem); ok {
								return m, m.loadFlagsCmd(*app.Id, *config.Id)
							}
						}
					}
				}
			}

			if m.activeView == flagsTable {
				selectedFlag := m.flagsTable.GetActiveRow()
				cmd := m.flagDetail.SetData(selectedFlag, m.flagsTable.EnvOrder())
				m.activeView = flagDetail
				return m, cmd
			}

			if m.activeView == flagDetail {
				if m.flagDetail.IsConfirming() {
					// Let FlagDetail handle enter for confirm/cancel buttons
					cmd := m.flagDetail.HandleMsg(msg)
					return m, cmd
				}
				m.flagDetail.ShowConfirm()
				return m, nil
			}
		}
	}

	var cmd tea.Cmd
	switch m.activeView {
	case appList:
		cmd = m.appsPanel.HandleMsg(msg)
	case configList:
		// In split view, only update config list (right pane)
		// App list (left pane) is read-only in this view
		if m.configsPanelError == "" && m.configsPanel != nil {
			cmd = m.configsPanel.HandleMsg(msg)
		}
	case flagsTable:
		cmd = m.flagsTable.HandleMsg(msg)
	case flagDetail:
		cmd = m.flagDetail.HandleMsg(msg)
	}
	return m, cmd
}

func (m Model) View() string {
	// Clear screen and add consistent top padding
	// The \033[H moves cursor to home position (0,0)
	// The \033[2J clears the screen
	view := "\033[H\033[2J"

	// Add spacing from top
	view += "\n"

	switch m.activeView {
	case appList:
		if m.appsPanelError != "" {
			view += m.appsPanel.RenderError(m.appsPanelError)
		} else {
			view += m.appsPanel.Render()
		}
	case configList:
		if m.configsPanelError != "" {
			view += m.configsPanel.RenderError(m.configsPanelError)
		} else {
			view += m.configsPanel.Render()
		}
	case flagsTable:
		if m.flagsTableError != "" {
			view += m.flagsTable.RenderError(m.flagsTableError)
		} else {
			view += m.flagsTable.Render()
		}
	case flagDetail:
		view += m.flagDetail.Render()
	}

	return view
}

// func (m Model) renderFlagDetailModal() string {
// 	// Render the background (flags table)
// 	background := m.flagsTable.Render()
//
// 	// Get the selected flag
// 	data := m.flagsTable.Data()
// 	if m.selectedFlagIdx < 0 || m.selectedFlagIdx >= len(data.Flags) {
// 		return background
// 	}
// 	flag := data.Flags[m.selectedFlagIdx]
//
// 	// Create the modal style
// 	modalStyle := lipgloss.NewStyle().
// 		Border(lipgloss.RoundedBorder()).
// 		BorderForeground(lipgloss.Color("#8cafd2")).
// 		Padding(1, 2).
// 		Width(40)
//
// 	// Modal content
// 	content := fmt.Sprintf("Flag: %s\n\nPress ESC to close", flag.FlagName)
// 	modal := modalStyle.Render(content)
//
// 	// Overlay the modal centered on the background
// 	return overlay.Composite(modal, background, overlay.Center, overlay.Center, 0, 0)
// }

// func (m Model) renderFlagDetailView() string {
// 	if m.selectedFlagIdx < 0 || m.selectedFlagIdx >= len(m.flagsData.Flags) {
// 		return "No flag selected"
// 	}
//
// 	flag := m.flagsData.Flags[m.selectedFlagIdx]
//
// 	var content strings.Builder
//
// 	// Flag name
// 	content.WriteString(fmt.Sprintf("Flag: %s\n", flag.FlagName))
// 	content.WriteString("\n")
//
// 	// Instructions
// 	content.WriteString("Use ← → (or h/l) to navigate\n")
// 	content.WriteString("Press SPACE or 't' to toggle state\n")
// 	content.WriteString("Press ESC to go back\n")
// 	content.WriteString("\n")
//
// 	// Environment states with selection indicator
// 	for i, envName := range m.flagsData.EnvOrder {
// 		state := flag.GetEnvState(envName)
//
// 		// Highlight selected environment
// 		if i == m.selectedEnvIdx {
// 			content.WriteString(fmt.Sprintf("→ %-20s │ %-10s\n", envName, state))
// 		} else {
// 			content.WriteString(fmt.Sprintf("  %-20s │ %-10s\n", envName, state))
// 		}
// 	}
//
// 	return RenderPanel(content.String(), "Edit Flag State", 50)
// }

type appsLoader struct {
	apps []appconfig.App
	err  error
}

func (m Model) loadAppsCmd() tea.Cmd {
	return func() tea.Msg {
		apps, err := m.appconfigClient.ListApps(context.Background())
		return appsLoader{apps: apps, err: err}
	}
}

type configsLoader struct {
	appId   string
	configs []appconfig.AppFlagConfig
	err     error
}

func (m Model) loadConfigsCmd(appId string) tea.Cmd {
	return func() tea.Msg {
		cachedConfig, ok := m.configsCache[appId]
		if ok {
			return cachedConfig
		}

		configs, err := m.appconfigClient.ListAppFlagConfigs(context.Background(), appId)
		result := configsLoader{appId: appId, configs: configs, err: err}

		return result
	}
}

type flagsLoader struct {
	flags []appconfig.Result
	err   error
}

// fetch all feature flags for all environments for a given app + config
// this request includes a request that costs $$$ so results are cached to file
func (m Model) loadFlagsCmd(appId string, configId string) tea.Cmd {
	return func() tea.Msg {
		cacheKey := fmt.Sprintf("%s:%s", appId, configId)
		cached, ok := m.filecache.Get(cacheKey)
		if ok {
			return flagsLoader{
				flags: cached,
				err:   nil,
			}
		}

		flags, err := m.appconfigClient.GetFlags(context.Background(), appId, configId)
		result := flagsLoader{
			flags: flags,
			err:   err,
		}
		m.filecache.Add(cacheKey, flags)
		return result
	}
}
