package app

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func RenderPanel(view string, title string, width int) string {
	lines := strings.Split(view, "\n")

	var result strings.Builder

	if title != "" {
		titleA := title
		maxTitleWidth := width - 8
		if len(title) > maxTitleWidth {
			titleA = title[:maxTitleWidth] + "..."
		}
		result.WriteString("┌─ ")
		result.WriteString(titleA)
		result.WriteString(" ")
		remainingWidth := width - len(title) - 5
		if remainingWidth > 0 {
			result.WriteString(strings.Repeat("─", remainingWidth))
		}
		result.WriteString("┐\n")
	} else {
		result.WriteString("┌")
		result.WriteString(strings.Repeat("─", width-2))
		result.WriteString("┐\n")
	}

	result.WriteString("│")
	result.WriteString(strings.Repeat(" ", width-2))
	result.WriteString("│\n")

	for _, line := range lines {
		if line == "" {
			// Preserve empty lines
			result.WriteString("│")
			result.WriteString(strings.Repeat(" ", width-2))
			result.WriteString("│\n")
			continue
		}

		result.WriteString("│ ")

		// Add internal margin (1 space) and truncate or pad the line to fit width
		contentWidth := width - 5 // 2 for borders, 2 for padding, 1 for internal margin
		result.WriteString(" ")   // internal margin
		lineWidth := lipgloss.Width(line)
		if lineWidth > contentWidth {
			// Truncate
			result.WriteString(line[:contentWidth])
		} else {
			// Pad
			result.WriteString(line)
			result.WriteString(strings.Repeat(" ", contentWidth-lineWidth))
		}

		result.WriteString(" │\n")
	}

	// Add bottom padding (empty line)
	result.WriteString("│")
	result.WriteString(strings.Repeat(" ", width-2))
	result.WriteString("│\n")

	// Bottom border
	result.WriteString("└")
	result.WriteString(strings.Repeat("─", width-2))
	result.WriteString("┘")

	return result.String()
}
