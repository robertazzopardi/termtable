package main

import (
	"fmt"
	"io"
	"log"
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
)

const (
	defaultWidth = 20
	listHeight   = 14
)

// Primary ansi colours
const (
	WHITE      = "15"
	RED        = "1"
	GREEN      = "2"
	YELLOW     = "3"
	BLUE       = "4"
	MAGENTA    = "5"
	GREY       = "240"
	LIGHT_GREY = "244"
)

var (
	titleStyle      = lipgloss.NewStyle()
	itemStyle       = lipgloss.NewStyle().PaddingLeft(4)
	paginationStyle = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	quitTextStyle   = lipgloss.NewStyle().Margin(1, 0, 2, 0)
	helpStyle       = blurredStyle.Copy()
	cursorStyle     = focusedItemStyle.Copy()
	noStyle         = lipgloss.NewStyle()

	selectedItemStyle   = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color(MAGENTA))
	focusedItemStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color(RED))
	focusedStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color(WHITE))
	blurredStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color(GREY))
	cursorModeHelpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(LIGHT_GREY))
	successStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color(GREEN))
	errorStyle          = lipgloss.NewStyle().Foreground(lipgloss.Color(RED))

	width  int = 100
	height int = 100
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
	list                list.Model
	newConnectionModel  NewConnectionModel
	currentView         CurrentView
	currentConnection   Connection
	openDatabase        OpenDatabase
	existingConnections ExistingConnectionsModel
}

func (m model) updateEvents(msg tea.Msg, cmd tea.Cmd) (model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		width = msg.Width
		height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "enter":
			i, ok := m.list.SelectedItem().(item)
			if ok {
				switch string(i) {
				case "New Connection":
					m.currentView = NEW_CONNECTION
					m.newConnectionModel =
						InitialNewConnectionModel()
				case "Edit Connection":
					m.currentView = EDIT_CONNECTION
				case "Join Existing":
					m.currentView = JOIN_EXISTING
					m.existingConnections = NewExistingConnectionsModel()
				}
			}
			return m, nil
		}
	}

	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch m.currentView {
	case NEW_CONNECTION:
		m.newConnectionModel, cmd = m.newConnectionModel.Update(msg)
		if m.newConnectionModel.connection.status == CONNECTED {
			m.currentView = DATABASE_VIEW
			m.currentConnection = m.newConnectionModel.connection
			m.openDatabase = NewOpenDatabase(m.currentConnection)

			SaveConnectionInKeyring(m.currentConnection)
		}

		if m.newConnectionModel.action == CANCEL {
			m.currentView = DEFAULT
		}

	case DATABASE_VIEW:
		m.openDatabase, cmd = m.openDatabase.Update(msg)
		if m.openDatabase.viewMode == QUIT {
			m.currentView = DEFAULT
			m.openDatabase = OpenDatabase{}
		}

	case JOIN_EXISTING:
		m.existingConnections, cmd = m.existingConnections.Update(msg)
		if m.existingConnections.selectedConnection != nil {
			m.currentView = DATABASE_VIEW
			m.currentConnection = *m.existingConnections.selectedConnection
			m.openDatabase = NewOpenDatabase(m.currentConnection)
		}

		if m.existingConnections.back {
			m.currentView = DEFAULT
		}

	case DEFAULT:
		m, cmd = m.updateEvents(msg, cmd)
	}

	return m, cmd
}

func (m model) View() string {
	switch m.currentView {
	case NEW_CONNECTION:
		return quitTextStyle.Render(m.newConnectionModel.View())
	case EDIT_CONNECTION:
		return quitTextStyle.Render("Edit Connection")
	case JOIN_EXISTING:
		return quitTextStyle.Render(m.existingConnections.View())
	case DATABASE_VIEW:
		return quitTextStyle.Render(m.openDatabase.View())
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

	l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
	l.Title = "Welcome to TermTable"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	m := model{list: l, currentView: DEFAULT}

	if _, err := tea.NewProgram(m, tea.WithAltScreen()).Run(); err != nil {
		log.Fatal("Error running program:", err)
	}
}
