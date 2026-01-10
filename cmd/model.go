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
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/simonschwartz/app-config-lazy-flags/internal/appconfig"
	"github.com/simonschwartz/app-config-lazy-flags/internal/filecache"
)

// pagination
// filter search
// add loading state
// add test mode

type view int

const (
	// initial screen displaying list of app config apps user has access to
	appList view = iota

	// displays list of feature flag configs for selected app
	configList

	// displays all feature flags and their state per environment
	flagsTable
)

type Model struct {
	appconfigClient appconfig.Client
	filecache       filecache.Cache

	activeView view

	appList list.Model

	configList   list.Model
	configsCache map[string]configsLoader

	flagsTable table.Model
}

func NewModel(appconfigClient *appconfig.Client, filecache *filecache.Cache) Model {
	return Model{
		appconfigClient: *appconfigClient,
		configsCache:    make(map[string]configsLoader),
		filecache:       *filecache,
		appList:         RenderAppList([]appconfig.App{}),
		activeView:      appList,
	}
}

func (m Model) Init() tea.Cmd {
	return m.loadAppsCmd()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case appsLoader:
		if msg.err != nil {
			return m, m.appList.NewStatusMessage(msg.err.Error())
		}
		m.appList = RenderAppList(msg.apps)
		return m, nil
	case configsLoader:
		if msg.err != nil {
			return m, m.appList.NewStatusMessage("Could not access configs")
		}
		m.activeView = configList
		m.configList = RenderConfigList(msg.configs)
		m.configsCache[msg.appId] = msg
		return m, nil
	case flagsLoader:
		if msg.err != nil {
		}
		m.activeView = flagsTable
		m.flagsTable = RenderFlagsTable(msg.flags)
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		// allow user to go back to previous view
		case "esc":
			switch m.activeView {
			case configList:
				m.activeView = appList
			case flagsTable:
				m.activeView = configList
			}
			return m, nil
		case "ctrl+c", "q":
			return m, tea.Quit
		case "enter":
			if m.activeView == appList {
				i, ok := m.appList.SelectedItem().(AppItem)
				if ok {
					return m, m.loadConfigsCmd(*i.Id)
				}
			}

			if m.activeView == configList {
				if config, ok := m.configList.SelectedItem().(ConfigItem); ok {
					if app, ok := m.appList.SelectedItem().(AppItem); ok {
						return m, m.loadFlagsCmd(*app.Id, *config.Id)
					}
				}
			}
		}
	}

	var cmd tea.Cmd
	switch m.activeView {
	case appList:
		m.appList, cmd = m.appList.Update(msg)
	case configList:
		m.configList, cmd = m.configList.Update(msg)
	case flagsTable:
		m.flagsTable, cmd = m.flagsTable.Update(msg)
	}
	return m, cmd
}

func (m Model) View() string {
	switch m.activeView {
	case appList:
		return m.appList.View()
	case configList:
		return m.configList.View()
	case flagsTable:
		return m.flagsTable.View()
	}
	return ""
}

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
