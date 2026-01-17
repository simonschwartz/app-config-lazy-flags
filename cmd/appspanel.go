package app

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/simonschwartz/app-config-lazy-flags/internal/appconfig"
)

// Manages the rendering of the Applications panel.
// Renders a navigatable list of AppConfig Apps.
// as returned by appconfigx.ListApps
//
// CLI output looks like this:
//
// ┌─ Applications ─────────────────────────────────┐
// │                                                │
// │  > Customer Portal                             │
// │    Intranet                                    │
// │    Online Shop                                 │
// │                                                │
// └────────────────────────────────────────────────┘

// To satisfy bubbletea list interface
type AppItem appconfig.App

func (i AppItem) FilterValue() string {
	return *i.Name
}

func NewAppsPanel(height int, width int, apps []appconfig.App) *ListPanel {
	var appItems []list.Item
	for _, app := range apps {
		appItems = append(appItems, AppItem(app))
	}

	return NewListPanel(height, width, "Applications", appItems)
}
