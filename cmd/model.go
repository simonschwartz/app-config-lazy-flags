package app

import (
	"context"
	"fmt"
	"log"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/simonschwartz/app-config-lazy-flags/internal/appconfig"
	"github.com/simonschwartz/app-config-lazy-flags/internal/filecache"
)

// generalise list component DONE
// navigate back views using esc DONE
// cache configs per app in memory DONE
// cache config errors in memory too DONE
// pagination
// filter search
// add loading state
// add test mode

type view int

const (
	appList view = iota
	configList
	flagsTable
)

type Model struct {
	appconfigClient appconfig.Client
	cache           filecache.Cache

	activeView view

	appList     list.Model
	apps        []appconfig.App
	selectedApp string

	configList   list.Model
	configs      []appconfig.AppFlagConfig
	configsCache map[string]configsLoader

	flagsTable table.Model
}

func NewModel(appconfigClient *appconfig.Client, cache *filecache.Cache) Model {
	return Model{
		appconfigClient: *appconfigClient,
		configsCache:    make(map[string]configsLoader),
		cache:           *cache,
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
			return m, tea.Quit
		}
		m.apps = msg.apps
		m.appList = RenderAppList(msg.apps)
		return m, nil
	case configsLoader:
		if msg.err != nil {
			return m, m.appList.NewStatusMessage("Could not access configs")
		}
		m.configs = msg.configs
		m.activeView = configList
		m.configList = RenderConfigList(msg.configs)
		return m, nil
	case flagsLoader:
		m.activeView = flagsTable
		m.flagsTable = RenderFlagsTable(msg.flags)

	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if m.activeView == appList {
				// do nothing - because default behaviour is to exit the app
				return m, nil
			}
			if m.activeView == configList {
				m.activeView = appList
				return m, nil
			}
			if m.activeView == flagsTable {
				m.activeView = configList
				return m, nil
			}
		case "ctrl+c", "q":
			return m, tea.Quit
		case "enter":
			if m.activeView == appList {
				i, ok := m.appList.SelectedItem().(AppItem)
				if ok {
					m.selectedApp = i.FilterValue()
					return m, m.loadConfigsCmd(*i.Id)
				}
			}

			if m.activeView == configList {
				_, ok := m.configList.SelectedItem().(ConfigItem)
				if ok {
					selectedAppId := m.appList.SelectedItem().(AppItem).Id
					selectedConfigId := m.configList.SelectedItem().(ConfigItem).Id
					return m, m.loadFlagsCmd(*selectedAppId, *selectedConfigId)
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
		view := m.configList.View()
		return view
	case flagsTable:
		view := m.flagsTable.View()
		return view
	}
	return ""
}

type appsLoader struct {
	apps []appconfig.App
	err  error
}

func (m Model) loadAppsCmd() tea.Cmd {
	return func() tea.Msg {
		apps, err := m.appconfigClient.ListApps(context.TODO())
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
			log.Printf("from cache")
			return cachedConfig
		}

		configs, err := m.appconfigClient.ListAppFlagConfigs(context.TODO(), appId)
		result := configsLoader{appId: appId, configs: configs, err: err}

		m.configsCache[appId] = result
		return result
	}
}

type flagsLoader struct {
	flags []appconfig.Result
	err   error
}

func (m Model) loadFlagsCmd(appId string, configId string) tea.Cmd {
	return func() tea.Msg {
		cacheKey := fmt.Sprintf("%s:%s", appId, configId)
		cached, ok := m.cache.Get(cacheKey)
		if ok {
			log.Printf("flags from cache")
			return flagsLoader{
				flags: cached,
				err:   nil,
			}
		}
		flags, err := m.appconfigClient.GetFlags(context.TODO(), appId, configId)
		result := flagsLoader{
			flags: flags,
			err:   err,
		}
		m.cache.Add(cacheKey, flags)
		log.Printf("flags not from cache")
		return result
	}
}
