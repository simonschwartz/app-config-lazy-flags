package app

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	overlay "github.com/rmhubbert/bubbletea-overlay"
)

// EnvItem represents an environment checkbox item
type EnvItem struct {
	envName string
	enabled bool
}

func (e EnvItem) FilterValue() string { return e.envName }

// checkboxDelegate renders list items as checkboxes
type checkboxDelegate struct{}

func (d checkboxDelegate) Height() int                             { return 1 }
func (d checkboxDelegate) Spacing() int                            { return 0 }
func (d checkboxDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d checkboxDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	item, ok := listItem.(EnvItem)
	if !ok {
		return
	}

	checkbox := "[ ]"
	if item.enabled {
		checkbox = "[x]"
	}

	str := checkbox + " " + item.envName

	// Style for selected vs unselected
	if index == m.Index() {
		str = "> " + str
		str = lipgloss.NewStyle().Foreground(nordfoxBlue).Render(str)
	} else {
		str = "  " + str
	}

	io.WriteString(w, str)
}

type FlagDetail struct {
	flagData         FlagRowData
	envOrder         []string
	model            list.Model
	renderBackground func() string

	// Confirmation state
	confirming      bool
	confirmEnvName  string
	confirmNewState bool // what state we're confirming to change to
	confirmBtnIdx   int  // 0 = Yes, 1 = Cancel
}

// Manages the rendering of the flag detail modal.
// Renders checkbox list of environments with flag states.
//
// CLI output looks like this:
//
// ┌─ dark_mode ──────────────────────────┐
// │                                      │
// │  > [x] development                   │
// │    [x] staging                       │
// │    [ ] production                    │
// │                                      │
// └──────────────────────────────────────┘
func NewFlagDetail(renderBackground func() string) *FlagDetail {
	delegate := checkboxDelegate{}
	l := list.New([]list.Item{}, delegate, 36, 10)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetShowFilter(false)
	l.SetShowHelp(false)
	l.SetShowPagination(false)

	return &FlagDetail{
		model:            l,
		renderBackground: renderBackground,
	}
}

func (f *FlagDetail) SetData(data FlagRowData, envOrder []string) tea.Cmd {
	f.flagData = data
	f.envOrder = envOrder
	f.confirming = false

	// Build list items from env states
	items := make([]list.Item, 0, len(envOrder))
	for _, envName := range envOrder {
		state := data.GetEnvState(envName)
		items = append(items, EnvItem{
			envName: envName,
			enabled: state == "on",
		})
	}

	return f.model.SetItems(items)
}

func (f *FlagDetail) HandleMsg(msg tea.Msg) tea.Cmd {
	if f.confirming {
		// Handle confirmation dialog navigation
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			switch keyMsg.String() {
			case "left", "h":
				f.confirmBtnIdx = 0
			case "right", "l":
				f.confirmBtnIdx = 1
			case "enter":
				if f.confirmBtnIdx == 0 {
					// Yes - apply the toggle
					f.applyToggle()
				}
				// Either way, close confirmation
				f.confirming = false
			case "esc":
				f.confirming = false
			}
		}
		return nil
	}

	var cmd tea.Cmd
	f.model, cmd = f.model.Update(msg)
	return cmd
}

func (f *FlagDetail) ShowConfirm() {
	idx := f.model.Index()
	items := f.model.Items()
	if idx >= 0 && idx < len(items) {
		if item, ok := items[idx].(EnvItem); ok {
			f.confirming = true
			f.confirmEnvName = item.envName
			f.confirmNewState = !item.enabled
			f.confirmBtnIdx = 1 // Default to Cancel for safety
		}
	}
}

func (f *FlagDetail) applyToggle() {
	idx := f.model.Index()
	items := f.model.Items()
	if idx >= 0 && idx < len(items) {
		if item, ok := items[idx].(EnvItem); ok {
			item.enabled = f.confirmNewState
			f.model.SetItem(idx, item)
		}
	}
}

func (f *FlagDetail) IsConfirming() bool {
	return f.confirming
}

func (f *FlagDetail) CancelConfirm() {
	f.confirming = false
}

func (f *FlagDetail) SelectedEnv() (string, bool) {
	idx := f.model.Index()
	items := f.model.Items()
	if idx >= 0 && idx < len(items) {
		if item, ok := items[idx].(EnvItem); ok {
			return item.envName, item.enabled
		}
	}
	return "", false
}

func (f *FlagDetail) Render() string {
	var modal string

	if f.confirming {
		modal = f.renderConfirmView()
	} else {
		modal = f.renderListView()
	}

	return overlay.Composite(modal, f.renderBackground(), overlay.Center, overlay.Center, 0, 0)
}

func (f *FlagDetail) renderListView() string {
	return RenderPanel(f.model.View(), f.flagData.FlagName, 40)
}

func (f *FlagDetail) renderConfirmView() string {
	action := "Enable"
	if !f.confirmNewState {
		action = "Disable"
	}

	// Button styles
	selectedStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#000")).
		Background(nordfoxBlue).
		Padding(0, 1)

	normalStyle := lipgloss.NewStyle().
		Padding(0, 1)

	yesBtn := fmt.Sprintf("Yes, %s", strings.ToLower(action))
	cancelBtn := "Cancel"

	if f.confirmBtnIdx == 0 {
		yesBtn = selectedStyle.Render(yesBtn)
		cancelBtn = normalStyle.Render(cancelBtn)
	} else {
		yesBtn = normalStyle.Render(yesBtn)
		cancelBtn = selectedStyle.Render(cancelBtn)
	}

	var content strings.Builder
	content.WriteString(fmt.Sprintf("%s %s in\n", action, f.flagData.FlagName))
	content.WriteString(fmt.Sprintf("%s?\n\n", f.confirmEnvName))
	content.WriteString(yesBtn + "     " + cancelBtn)

	return RenderPanel(content.String(), "Confirm", 40)
}
