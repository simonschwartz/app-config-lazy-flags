package app

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/simonschwartz/app-config-lazy-flags/internal/appconfig"
)

// Manages the rendering of the configs panel.
// Renders a navigatable list of AppConfig configuraition profiles
// as returned by appconfigx.ListAppFlagConfigs
//
// CLI output looks like this:
//
// ┌─ Configuration Profiles ───────────────────────┐
// │                                                │
// │  > Frontend Flags                              │
// │    Backend Flags                               │
// │    App Flags                                   │
// │                                                │
// └────────────────────────────────────────────────┘

// To satisfy bubbletea list interface
type ConfigItem appconfig.AppFlagConfig

func (i ConfigItem) FilterValue() string {
	return *i.Name
}

func NewConfigsPanel(height int, width int, configs []appconfig.AppFlagConfig) *ListPanel {
	var appConfigs []list.Item
	for _, config := range configs {
		appConfigs = append(appConfigs, ConfigItem(config))
	}

	return NewListPanel(height, width, "Configuration Profiles", appConfigs)
}
