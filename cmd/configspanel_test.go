package app_test

import (
	"strings"
	"testing"

	"github.com/simonschwartz/app-config-lazy-flags/cmd"
	"github.com/simonschwartz/app-config-lazy-flags/internal/appconfig"
)

func stringPtr(s string) *string {
	return &s
}

func TestConfigsPanelRender(t *testing.T) {
	tests := []struct {
		name     string
		configs  []appconfig.AppFlagConfig
		expected string
	}{
		{
			name: "should render config list with multiple items",
			configs: []appconfig.AppFlagConfig{
				{
					ApplicationId: stringPtr("1234"),
					Name:          stringPtr("Frontend Flags"),
					Id:            stringPtr("dummyID"),
				},
				{
					ApplicationId: stringPtr("1234"),
					Name:          stringPtr("Backend Flags"),
					Id:            stringPtr("dummyID"),
				},
				{
					ApplicationId: stringPtr("1234"),
					Name:          stringPtr("App Flags"),
					Id:            stringPtr("dummyID"),
				},
			},
			expected: strings.Join([]string{
				"┌─ Configuration Profiles ───────────────────────┐",
				"│                                                │",
				"│  > Frontend Flags                              │",
				"│    Backend Flags                               │",
				"│    App Flags                                   │",
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
			name:    "should render config list with no items",
			configs: []appconfig.AppFlagConfig{},
			expected: strings.Join([]string{
				"┌─ Configuration Profiles ───────────────────────┐",
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
			configsPanel := app.NewConfigsPanel(20, 50, tt.configs)
			result := configsPanel.Render()

			if result != tt.expected {
				t.Errorf("result: \n %v, expected \n %v", result, tt.expected)
			}
		})
	}
}

func TestConfigsPanelRenderError(t *testing.T) {
	tests := []struct {
		name     string
		errMsg   string
		expected string
	}{
		{
			name:   "should render config list with error",
			errMsg: "Access: Permission Denied",
			expected: strings.Join([]string{
				"┌─ Configuration Profiles ───────────────────────┐",
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
			configsPanel := app.NewConfigsPanel(20, 50, []appconfig.AppFlagConfig{})
			result := configsPanel.RenderError(tt.errMsg)

			if result != tt.expected {
				t.Errorf("result: \n %v, expected \n %v", result, tt.expected)
			}
		})
	}
}
