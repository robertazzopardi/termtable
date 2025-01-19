package main

import (
	"fmt"
	"io"

	// "log"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	// "github.com/gdamore/tcell/v2"
)

type CurrentView string

const TERMTABLE_TEXT = `
  __                        __        ___.   .__          
_/  |_  ___________  ______/  |______ \_ |__ |  |   ____  
\   __\/ __ \_  __ \/     \   __\__  \ | __ \|  | _/ __ \ 
 |  | \  ___/|  | \/  Y Y  \  |  / __ \| \_\ \  |_\  ___/ 
 |__|  \___  >__|  |__|_|  /__| (____  /___  /____/\___  >
           \/            \/          \/    \/          \/ 
`

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
	helpStyle       = blurredStyle.Copy().PaddingLeft(2)
	cursorStyle     = focusedItemStyle.Copy()
	noStyle         = lipgloss.NewStyle()

	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color(MAGENTA))
	focusedItemStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color(RED))
	focusedStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color(WHITE))
	blurredStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color(GREY))
	successStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color(GREEN))
	errorStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color(RED))

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

func (m model) updateEvents(msg tea.Msg) (model, tea.Cmd) {
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

	var cmd tea.Cmd
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

	case DEFAULT, EDIT_CONNECTION:
		m, cmd = m.updateEvents(msg)
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

func hotkeys(app *tview.Application) *tview.List {
	// hotkeyTable := tview.NewTable()

	// lorem := strings.Split("Lorem ipsum dolor", " ")
	// // cols, rows := 5, 2
	// // word := 0
	// // for r := 0; r < rows; r++ {
	// // 	for c := 0; c < cols; c++ {
	// // 		color := tcell.ColorWhite
	// // 		if c < 1 || r < 1 {
	// // 			color = tcell.ColorYellow
	// // 		}
	// // 		word = (word + 1) % len(lorem)
	// // 	}
	// // }

	list := tview.NewList().ShowSecondaryText(false).
		SetSelectedFocusOnly(true).
		SetShortcutStyle(tcell.StyleDefault)

	list.
		AddItem("List item 1", "", 0, nil).
		AddItem("List item 2", "", 0, nil).
		AddItem("List item 3", "", 0, nil).
		AddItem("List item 4", "", 0, nil).
		AddItem("<q> quit", "", 0, func() {
			app.Stop()
		})

	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'a':
			list.SetCurrentItem(0)
			app.Stop()
			return nil
		case 'b':
			list.SetCurrentItem(1)
			app.Stop()
			return nil
		}
		return event
	})

	// 		hotkeyTable.SetCell(0, 0,
	// 			tview.NewTableCell(lorem[word]).
	// 				SetTextColor(color).
	// 				SetAlign(tview.AlignCenter))

	return list
}

func currentConnectionInfo() *tview.List {
	list := tview.NewList().ShowSecondaryText(false).
		SetSelectedFocusOnly(true).
		AddItem(" Name: ", "", 0, nil).
		AddItem(" Database: ", "", 0, nil).
		AddItem(" Host: ", "", 0, nil).
		AddItem(" PORT: ", "", 0, nil).
		AddItem(" USER: ", "Press to exit", 0, nil)

	return list
}

func header(app *tview.Application) *tview.Flex {
	connection := currentConnectionInfo()

	table := hotkeys(app)

	headerView := tview.NewFlex().
		AddItem(connection, 0, 1, false).
		AddItem(table, 0, 2, false)
		// AddItem(tview.NewBox(), 0, 2, false).
		// AddItem(tview.NewBox(), 0, 3, false)
	return headerView
}

func mainView(app *tview.Application) *tview.Flex {
	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(header(app), 0, 1, false).
		AddItem(tview.NewBox().SetBorder(true).SetTitle("Connections"), 0, 6, false)

	return flex
}

func main() {
	app := tview.NewApplication()
	mainView := mainView(app)

	if err := app.SetRoot(mainView, true).SetFocus(mainView).Run(); err != nil {
		panic(err)
	}

	// items := []list.Item{
	// 	item("New Connection"),
	// 	item("Edit Connection"),
	// 	item("Join Existing"),
	// }

	// l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
	// l.Title = "Welcome to TermTable"
	// l.SetShowStatusBar(false)
	// l.SetFilteringEnabled(false)
	// l.Styles.Title = titleStyle
	// l.Styles.PaginationStyle = paginationStyle
	// l.Styles.HelpStyle = helpStyle

	// m := model{list: l, currentView: DEFAULT}

	// if _, err := tea.NewProgram(m, tea.WithAltScreen()).Run(); err != nil {
	// 	log.Fatal("Error running program:", err)
	// }
}
