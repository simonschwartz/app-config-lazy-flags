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

	appsPanel *ListPanel
	appsError string // stores apps loading error message

	configsPanel *ListPanel
	configsCache map[string]configsLoader
	configError  string // stores config loading error message

	flagsTable *FlagsTable

	// Flag detail view state
	selectedFlagIdx int // which flag in flagsData.Flags is selected (-1 = none)
	selectedEnvIdx  int // which environment is highlighted in detail view
}

func NewModel(appconfigClient *appconfig.Client, filecache *filecache.Cache) Model {
	appsPanel := NewAppsPanel(20, 50, []appconfig.App{})
	configsPanel := NewConfigsPanel(20, 50, []appconfig.AppFlagConfig{})
	flagsTable := NewFlagsTable(20, 50, []appconfig.Result{})

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
	}
}

func (m Model) Init() tea.Cmd {
	return m.loadAppsCmd()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case appsLoader:
		if msg.err != nil {
			m.appsError = fmt.Sprintf("Error: %v", msg.err)
			return m, nil
		}
		m.appsError = ""
		var appItems []list.Item
		for _, app := range msg.apps {
			appItems = append(appItems, AppItem(app))
		}
		cmd := m.appsPanel.SetItems(appItems)
		return m, cmd
	case configsLoader:
		if msg.err != nil {
			m.configError = fmt.Sprintf("Error: %v", msg.err)
			m.activeView = configList
			return m, nil
		}
		m.configError = ""
		m.activeView = configList
		m.configsPanel = NewConfigsPanel(20, 50, msg.configs)
		m.configsCache[msg.appId] = msg
		return m, nil
	case flagsLoader:
		if msg.err != nil {
			return m, nil
		}
		m.activeView = flagsTable
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
				if m.configsPanel != nil {
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
			}

			// if m.activeView == flagsTable {
			// 	cursor := m.flagsTable.Cursor()
			// 	if cursor >= 0 && cursor < len(m.flagsData.Flags) {
			// 		m.selectedFlagIdx = cursor
			// 		m.selectedEnvIdx = 0
			// 		m.activeView = flagDetail
			// 		return m, nil
			// 	}
			// }
		}
	}

	var cmd tea.Cmd
	switch m.activeView {
	case appList:
		cmd = m.appsPanel.HandleMsg(msg)
	case configList:
		// In split view, only update config list (right pane)
		// App list (left pane) is read-only in this view
		if m.configError == "" && m.configsPanel != nil {
			cmd = m.configsPanel.HandleMsg(msg)
		}
	case flagsTable:
		cmd = m.flagsTable.HandleMsg(msg)
	case flagDetail:
		// Don't pass key events to underlying components in detail view
		return m, nil
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
		if m.appsError != "" {
			view += m.appsPanel.RenderError(m.appsError)
		} else {
			view += m.appsPanel.Render()
		}
	case configList:
		if m.configsPanel != nil {
			if m.configError != "" {
				view += m.configsPanel.RenderError(m.configError)
			} else {
				view += m.configsPanel.Render()
			}
		}
	case flagsTable:
		view += m.flagsTable.Render()
		// case flagDetail:
		// 	view += m.renderFlagDetailView()
	}

	return view
}

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
