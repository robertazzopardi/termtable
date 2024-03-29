package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type CurrentView string

const (
	DEFAULT         CurrentView = "DEFAULT"
	NEW_CONNECTION  CurrentView = "NEW_CONNECTION"
	EDIT_CONNECTION CurrentView = "EDIT_CONNECTION"
	JOIN_EXISTING   CurrentView = "JOIN_EXISTING"
	DATABASE_VIEW   CurrentView = "DATABASE_VIEW"
	QUITTING        CurrentView = "QUITTING"
)

const listHeight = 14

var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	quitTextStyle     = lipgloss.NewStyle().Margin(1, 0, 2, 4)

	focusedStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	cursorStyle         = focusedStyle.Copy()
	noStyle             = lipgloss.NewStyle()
	helpStyle           = blurredStyle.Copy()
	cursorModeHelpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))

	focusedButton = focusedStyle.Copy().Render("[ Submit ]")
	blurredButton = fmt.Sprintf("[ %s ]", blurredStyle.Render("Submit"))
)

type item string

func (i item) FilterValue() string { return "" }

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

type model struct {
	list               list.Model
	newConnectionModel NewConnectionModel
	currentView        CurrentView
}

type (
	tickMsg struct{}
)

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		if m.currentView == DEFAULT {
			switch keypress := msg.String(); keypress {
			case "q", "ctrl+c":
				m.currentView = QUITTING
				return m, nil

			case "enter":
				i, ok := m.list.SelectedItem().(item)
				if ok {
					// m.choice = string(i)
					switch string(i) {
					case "New Connection":
						m.currentView = NEW_CONNECTION
					case "Edit Connection":
						m.currentView = EDIT_CONNECTION
					case "Join Existing":
						m.currentView = JOIN_EXISTING
						// default:
						// 	m.currentView = DEFAULT
					}
				}
				return m, nil
			}
		}
	}

	var cmd tea.Cmd

	if m.currentView == NEW_CONNECTION {
		m.newConnectionModel, cmd = m.newConnectionModel.Update(msg)
		if m.newConnectionModel.submitted {
			m.currentView = DATABASE_VIEW
			return m, nil
		}
		return m, cmd
	}

	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	switch m.currentView {
	case NEW_CONNECTION:
		return quitTextStyle.Render(fmt.Sprintf("%4s", m.newConnectionModel.View()))
	case EDIT_CONNECTION:
		return quitTextStyle.Render("Edit Connection")
	case JOIN_EXISTING:
		return quitTextStyle.Render("Join Existing")
	case DATABASE_VIEW:
		return quitTextStyle.Render("Database View")
	case QUITTING:
		return quitTextStyle.Render("Are you sure you want to quit? (ctrl+c/esc again)")
	default:
		return "\n" + m.list.View()
	}
}

func main() {
	items := []list.Item{
		item("New Connection"),
		item("Edit Connection"),
		item("Join Existing"),
	}

	const defaultWidth = 20

	l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
	l.Title = "Welcome to TermTable"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	m := model{list: l, newConnectionModel: InitialNewConnectionModel(), currentView: DEFAULT}

	if _, err := tea.NewProgram(m, tea.WithAltScreen()).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
