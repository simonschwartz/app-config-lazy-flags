package app_test

import (
	"strings"
	"testing"

	"github.com/simonschwartz/app-config-lazy-flags/cmd"
	"github.com/simonschwartz/app-config-lazy-flags/internal/appconfig"
)

func TestAppsPanelRender(t *testing.T) {
	tests := []struct {
		name     string
		apps     []appconfig.App
		expected string
	}{
		{
			name: "should render config list with multiple items",
			apps: []appconfig.App{
				{
					Description: stringPtr("Dummy Description"),
					Name:        stringPtr("Customer Portal"),
					Id:          stringPtr("dummyID"),
				},
				{
					Description: stringPtr("Dummy Description"),
					Name:        stringPtr("Intranet"),
					Id:          stringPtr("dummyID"),
				},
				{
					Description: stringPtr("Dummy Description"),
					Name:        stringPtr("Online Shop"),
					Id:          stringPtr("dummyID"),
				},
			},
			expected: strings.Join([]string{
				"┌─ Applications ─────────────────────────────────┐",
				"│                                                │",
				"│  > Customer Portal                             │",
				"│    Intranet                                    │",
				"│    Online Shop                                 │",
				"│                                                │",
				"│                                                │",
				"│                                                │",
				"│                                                │",
				"│                                                │",
				"│                                                │",
				"│                                                │",
				"│                                                │",
				"│                                                │",
				"│                                                │",
				"│                                                │",
				"│                                                │",
				"│                                                │",
				"│                                                │",
				"│                                                │",
				"│                                                │",
				"│                                                │",
				"│                                                │",
				"└────────────────────────────────────────────────┘",
			}, "\n"),
		},
		{
			name: "should render config list with no items",
			apps: []appconfig.App{},
			expected: strings.Join([]string{
				"┌─ Applications ─────────────────────────────────┐",
				"│                                                │",
				"│  No items.                                     │",
				"│                                                │",
				"│                                                │",
				"│                                                │",
				"│                                                │",
				"│                                                │",
				"│                                                │",
				"│                                                │",
				"│                                                │",
				"│                                                │",
				"│                                                │",
				"│                                                │",
				"│                                                │",
				"│                                                │",
				"│                                                │",
				"│                                                │",
				"│                                                │",
				"│                                                │",
				"│                                                │",
				"│                                                │",
				"│                                                │",
				"└────────────────────────────────────────────────┘",
			}, "\n"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configsPanel := app.NewAppsPanel(20, 50, tt.apps)
			result := configsPanel.Render()

			if result != tt.expected {
				t.Errorf("result: \n %v, expected \n %v", result, tt.expected)
			}
		})
	}
}

func TestAppsPanelRenderError(t *testing.T) {
	tests := []struct {
		name     string
		errMsg   string
		expected string
	}{
		{
			name:   "should render apps list with error",
			errMsg: "Access: Permission Denied",
			expected: strings.Join([]string{
				"┌─ Applications ─────────────────────────────────┐",
				"│                                                │",
				"│  Access: Permission Denied                     │",
				"│                                                │",
				"│                                                │",
				"│                                                │",
				"│                                                │",
				"│                                                │",
				"│                                                │",
				"│                                                │",
				"│                                                │",
				"│                                                │",
				"│                                                │",
				"│                                                │",
				"│                                                │",
				"│                                                │",
				"│                                                │",
				"│                                                │",
				"│                                                │",
				"│                                                │",
				"│                                                │",
				"│                                                │",
				"│                                                │",
				"└────────────────────────────────────────────────┘",
			}, "\n"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configsPanel := app.NewAppsPanel(20, 50, []appconfig.App{})
			result := configsPanel.RenderError(tt.errMsg)

			if result != tt.expected {
				t.Errorf("result: \n %v, expected \n %v", result, tt.expected)
			}
		})
	}
}
