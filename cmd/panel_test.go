package app_test

import (
	"strings"
	"testing"

	"github.com/simonschwartz/app-config-lazy-flags/cmd"
)

func TestRenderPanel(t *testing.T) {
	tests := []struct {
		name     string
		view     string
		title    string
		width    int
		expected string
	}{
		{
			name:  "should render border box",
			view:  "hello world",
			title: "title",
			width: 50,
			expected: strings.Join([]string{
				"┌─ title ────────────────────────────────────────┐",
				"│                                                │",
				"│  hello world                                   │",
				"│                                                │",
				"└────────────────────────────────────────────────┘",
			}, "\n"),
		},
		{
			name:  "should render border box small width",
			view:  "hello world",
			title: "title",
			width: 20,
			expected: strings.Join([]string{
				"┌─ title ──────────┐",
				"│                  │",
				"│  hello world     │",
				"│                  │",
				"└──────────────────┘",
			}, "\n"),
		},
		{
			name:  "should render border box with multi line view",
			view:  strings.Join([]string{"hello", "world"}, "\n"),
			title: "title",
			width: 50,
			expected: strings.Join([]string{
				"┌─ title ────────────────────────────────────────┐",
				"│                                                │",
				"│  hello                                         │",
				"│  world                                         │",
				"│                                                │",
				"└────────────────────────────────────────────────┘",
			}, "\n"),
		},
		{
			name:  "should truncate title longer than width",
			view:  "hello world",
			title: strings.Repeat("a", 100),
			width: 50,
			expected: strings.Join([]string{
				"┌─ aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa... ┐",
				"│                                                │",
				"│  hello world                                   │",
				"│                                                │",
				"└────────────────────────────────────────────────┘",
			}, "\n"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := app.RenderPanel(tt.view, tt.title, tt.width)

			if result != tt.expected {
				t.Errorf("result: \n %v, expected \n %v \n test", result, tt.expected)
			}
		})
	}
}
