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

const APP_NAME = `_________________       
\______ \______  \______
 |    |  \  /    /  ___/
 |    ` + "`" + `   \/    /\___ \ 
/_______  /____//____  >
        \/           \/ 
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

type HotKey struct {
	desc     string
	shortcut rune
}

type HotKeys struct {
	*tview.List
	values []HotKey
}

func NewHotkeys() *HotKeys {
	list := tview.NewList().
		ShowSecondaryText(false).SetSelectedFocusOnly(true)
	return &HotKeys{
		List:   list,
		values: []HotKey{},
	}
}

func (r *HotKeys) AddHotKey(desc string, shortcut rune) *HotKeys {
	r.values = append(r.values, HotKey{desc, shortcut})
	return r
}

func (r *HotKeys) Draw(screen tcell.Screen) {
	r.Box.DrawForSubclass(screen, r)
	x, y, width, height := r.GetInnerRect()

	for index, hotkey := range r.values {
		if index >= height {
			break
		}

		line := fmt.Sprintf("<%s> %s", string(hotkey.shortcut), hotkey.desc)
		tview.Print(screen, line, x, y+index, width, tview.AlignLeft, tcell.ColorYellow)
	}
}

// func (r *HotKeys) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
// 	return r.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
// 		for _, hotkey := range r.values {
// 			if event.Rune() == hotkey.shortcut && hotkey.action != nil {
// 				hotkey.action()
// 			}
// 		}
// 	})
// }

func currentConnectionInfo() *tview.List {
	list := tview.NewList().
		ShowSecondaryText(false).
		SetSelectedFocusOnly(true).
		AddItem("Name: ", "", 0, nil).
		AddItem("Host: ", "", 0, nil).
		AddItem("PORT: ", "", 0, nil).
		AddItem("USER: ", "Press to exit", 0, nil).
		AddItem("Database: ", "", 0, nil)

	return list
}

func header(hotkeyView *HotKeys) *tview.Flex {
	connection := currentConnectionInfo()

	appName := tview.NewTextView().SetText(APP_NAME).SetTextAlign(tview.AlignRight)

	headerView := tview.NewFlex().
		AddItem(connection, 0, 1, false).
		AddItem(hotkeyView, 0, 1, false).
		AddItem(appName, 0, 1, false)
	headerView.SetBorderPadding(0, 0, 1, 1)

	return headerView
}

func newConnectionForm(app *tview.Application) *tview.Flex {
	form := tview.NewForm().
		AddInputField("Name", "", 26, nil, nil).
		AddInputField("Host", "", 26, nil, nil).
		AddInputField("Port", "", 26, nil, nil).
		AddInputField("User", "", 26, nil, nil).
		AddPasswordField("Password", "", 26, '*', nil).
		AddInputField("Database", "", 26, nil, nil).
		AddButton("Save", nil).
		AddButton("Test", nil).
		AddButton("Connect", func() {
			app.Stop()
		})
	form.SetBorder(true)
	form.SetButtonsAlign(tview.AlignRight)

	modal := tview.NewFlex().
		AddItem(nil, 0, 3, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(form, 0, 1, true).
			AddItem(nil, 0, 1, false), 0, 2, true).
		AddItem(nil, 0, 3, false)

	return modal
}

const (
	MAIN_PAGE          = "main"
	NEW_CONNETION_FORM = "newConnection"
	SAVED_CONNECTIONS  = "savedConnections"
)

func savedConnections() *tview.Box {
	connectionsView := tview.NewBox().SetBorder(true).SetTitle("Connections")

	connections, err := ListConnections()

	if err != nil {
		return connectionsView
	}

	table := tview.NewTable()

	table.SetCell(0, 0, tview.NewTableCell("Name").SetExpansion(1))
	table.SetCell(0, 1, tview.NewTableCell("Host").SetExpansion(1))
	table.SetCell(0, 2, tview.NewTableCell("Port").SetExpansion(1))
	table.SetCell(0, 3, tview.NewTableCell("User").SetExpansion(1))
	table.SetCell(0, 4, tview.NewTableCell("Database").SetExpansion(1))

	for i, conn := range connections {
		table.SetCell(i+1, 0, tview.NewTableCell(conn.Name))
		table.SetCell(i+1, 1, tview.NewTableCell(conn.Host))
		table.SetCell(i+1, 2, tview.NewTableCell(conn.Port))
		table.SetCell(i+1, 3, tview.NewTableCell(conn.User))
		table.SetCell(i+1, 4, tview.NewTableCell(conn.Database))
	}

	table.SetBorderPadding(0, 0, 1, 1)

	table.SetSelectable(true, false)
	// table.Select(0, 0).SetFixed(1, 1).SetDoneFunc(func(key tcell.Key) {
	// 	if key == tcell.KeyEscape {
	// 		app.Stop()
	// 	}
	// 	if key == tcell.KeyEnter {
	// 		table.SetSelectable(true, true)
	// 	}
	// }).SetSelectedFunc(func(row int, column int) {
	// 	table.GetCell(row, column).SetTextColor(tcell.ColorRed)
	// 	table.SetSelectable(false, false)
	// })

	connectionsView.SetDrawFunc(func(screen tcell.Screen, x int, y int, w int, h int) (int, int, int, int) {
		centerY := y + 1
		centerX := x + 1

		table.SetRect(centerX, centerY, w-2, h-2)
		table.Draw(screen)

		// Space for other content.
		return x + 1, centerY + 1, w - 2, h - (centerY + 1 - y)
	})

	return connectionsView
}

func mainContent() *tview.Pages {
	pages := tview.NewPages().
		AddPage(SAVED_CONNECTIONS, savedConnections(), true, true)

	return pages
}

type Layout struct {
	*tview.Flex
	header  *tview.Flex
	content *tview.Pages
}

func newLayout(header *tview.Flex, content *tview.Pages) Layout {
	view := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(header, 0, 1, false).
		AddItem(content, 0, 6, false)

	return Layout{view, header, content}
}

func main() {
	app := tview.NewApplication()

	hotkeyView := NewHotkeys()
	header := header(hotkeyView)
	pages := mainContent()
	mainView := newLayout(header, pages)

	mainPanels := tview.NewPages().
		AddPage(MAIN_PAGE, mainView, true, true)

	hotkeyView.
		AddHotKey("New Connection", 'n').
		AddHotKey("Quit", 'q')

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		pageName, _ := mainPanels.GetFrontPage()
		contentName, _ := mainView.content.GetFrontPage()

		switch contentName {
		case SAVED_CONNECTIONS:
			switch event.Rune() {
			case 'q':
				app.Stop()
			case 'n':
				if pageName == NEW_CONNETION_FORM {
					return event
				}
				addConnectionForm := newConnectionForm(app)
				mainPanels.AddPage(NEW_CONNETION_FORM, addConnectionForm, true, true)
				return nil
			}

			if event.Key() == tcell.KeyESC {
				mainPanels.RemovePage(NEW_CONNETION_FORM)
			}

		default:

		}

		return event
	})

	if err := app.SetRoot(mainPanels, true).SetFocus(mainView).Run(); err != nil {
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
