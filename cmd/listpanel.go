package app

import (
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ListPanel struct {
	height int
	width  int
	title  string
	model  list.Model
}

// controls the display of active list item, and satisfies some
// interface requirements of bubbletea list.
type itemDelegate struct{}
func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	str := listItem.FilterValue()

	maxWidth := m.Width() - 4
	if len(str) > maxWidth {
		str = str[:maxWidth-3] + "..."
	}

	// Style for selected vs unselected
	if index == m.Index() {
		str = "> " + str
		str = lipgloss.NewStyle().Foreground(lipgloss.Color("#8cafd2")).Render(str)
	} else {
		str = "  " + str
	}

	io.WriteString(w, str)
}

// Render a navigatable bubbletea list inside a beautiful ascii panel.
func NewListPanel(height int, width int, title string, items []list.Item) *ListPanel {
	delegate := itemDelegate{}

	// Create the list model
	l := list.New(items, delegate, width-4, height)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetShowFilter(false)
	l.SetShowHelp(false)
	l.SetShowPagination(false)

	return &ListPanel{
		height: height,
		width:  width,
		title:  title,
		model:  l,
	}
}

func (p *ListPanel) Render() string {
	return RenderPanel(p.model.View(), p.title, p.width)
}

func (p *ListPanel) RenderError(errMsg string) string {
	paddedErrMsg := errMsg + strings.Repeat("\n", p.height-1)
	return RenderPanel(paddedErrMsg, p.title, p.width)
}

func (p *ListPanel) SetItems(items []list.Item) tea.Cmd {
	return p.model.SetItems(items)
}

func (p *ListPanel) SelectedItem() (list.Item, bool) {
	item := p.model.SelectedItem()
	return item, item != nil
}

func (p *ListPanel) HandleMsg(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	p.model, cmd = p.model.Update(msg)
	return cmd
}
