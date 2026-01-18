package app_test

import (
	"strings"
	"testing"

	"github.com/simonschwartz/app-config-lazy-flags/cmd"
	"github.com/simonschwartz/app-config-lazy-flags/internal/appconfig"
)

func TestFlagsTableRender(t *testing.T) {
	tests := []struct {
		name     string
		flags    []appconfig.Result
		expected string
	}{
		{
			name: "should render flags",
			flags: []appconfig.Result{
				{
					EnvName: "development",
					Flags: appconfig.Flags{
						"dark_mode":    {Enabled: true},
						"new_checkout": {Enabled: false},
						"beta_feature": {Enabled: true},
					},
				},
				{
					EnvName: "staging",
					Flags: appconfig.Flags{
						"dark_mode":    {Enabled: true},
						"new_checkout": {Enabled: true},
						"beta_feature": {Enabled: false},
					},
				},
				{
					EnvName: "production",
					Flags: appconfig.Flags{
						"dark_mode":    {Enabled: false},
						"new_checkout": {Enabled: false},
						"beta_feature": {Enabled: false},
					},
				},
			},
			expected: strings.Join([]string{
				"┌─ Feature Flags ─────────────────────────────────────────────────────────┐",
				"│                                                                         │",
				"│  Flag Name             development      staging          production     │",
				"│  beta_feature          on               off              off            │",
				"│  dark_mode             on               on               off            │",
				"│  new_checkout          off              on               off            │",
				"│                                                                         │",
				"│                                                                         │",
				"│                                                                         │",
				"│                                                                         │",
				"│                                                                         │",
				"│                                                                         │",
				"│                                                                         │",
				"│                                                                         │",
				"│                                                                         │",
				"│                                                                         │",
				"│                                                                         │",
				"│                                                                         │",
				"│                                                                         │",
				"│                                                                         │",
				"│                                                                         │",
				"│                                                                         │",
				"│                                                                         │",
				"│                                                                         │",
				"└─────────────────────────────────────────────────────────────────────────┘",
			}, "\n"),
		},
		{
			name:  "should render flags table with no flags",
			flags: []appconfig.Result{},
			expected: strings.Join([]string{
				"┌─ Feature Flags ────────────────────────────────┐",
				"│                                                │",
				"│  You have no flags                             │",
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
			flagsTable := app.NewFlagsTable(20, 50, tt.flags)
			result := flagsTable.Render()

			if result != tt.expected {
				t.Errorf("result: \n %v, expected \n %v", result, tt.expected)
			}
		})
	}
}

// func TestConfigsPanelRenderError(t *testing.T) {
// 	tests := []struct {
// 		name     string
// 		errMsg   string
// 		expected string
// 	}{
// 		{
// 			name:   "should render config list with error",
// 			errMsg: "Access: Permission Denied",
// 			expected: strings.Join([]string{
// 				"┌─ Configuration Profiles ───────────────────────┐",
// 				"│                                                │",
// 				"│  Access: Permission Denied                     │",
// 				"│                                                │",
// 				"│                                                │",
// 				"│                                                │",
// 				"│                                                │",
// 				"│                                                │",
// 				"│                                                │",
// 				"│                                                │",
// 				"│                                                │",
// 				"│                                                │",
// 				"│                                                │",
// 				"│                                                │",
// 				"│                                                │",
// 				"│                                                │",
// 				"│                                                │",
// 				"│                                                │",
// 				"│                                                │",
// 				"│                                                │",
// 				"│                                                │",
// 				"│                                                │",
// 				"│                                                │",
// 				"└────────────────────────────────────────────────┘",
// 			}, "\n"),
// 		},
// 	}
//
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			configsPanel := app.NewConfigsPanel(20, 50, []appconfig.AppFlagConfig{})
// 			result := configsPanel.RenderError(tt.errMsg)
//
// 			if result != tt.expected {
// 				t.Errorf("result: \n %v, expected \n %v", result, tt.expected)
// 			}
// 		})
// 	}
// }
