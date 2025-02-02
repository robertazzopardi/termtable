package main

import (
	"fmt"
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
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
	// DATABASE_VIEW   CurrentView = "DATABASE_VIEW"
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

// type model struct {
// 	list                list.Model
// 	newConnectionModel  NewConnectionModel
// 	currentView         CurrentView
// 	currentConnection   Connection
// 	openDatabase        OpenDatabase
// 	existingConnections ExistingConnectionsModel
// }

// func (m model) updateEvents(msg tea.Msg) (model, tea.Cmd) {
// 	switch msg := msg.(type) {
// 	case tea.WindowSizeMsg:
// 		m.list.SetWidth(msg.Width)
// 		width = msg.Width
// 		height = msg.Height
// 		return m, nil

// 	case tea.KeyMsg:
// 		switch keypress := msg.String(); keypress {
// 		case "q", "ctrl+c":
// 			return m, tea.Quit

// 		case "enter":
// 			i, ok := m.list.SelectedItem().(item)
// 			if ok {
// 				switch string(i) {
// 				case "New Connection":
// 					m.currentView = NEW_CONNECTION
// 					m.newConnectionModel =
// 						InitialNewConnectionModel()
// 				case "Edit Connection":
// 					m.currentView = EDIT_CONNECTION
// 				case "Join Existing":
// 					m.currentView = JOIN_EXISTING
// 					m.existingConnections = NewExistingConnectionsModel()
// 				}
// 			}
// 			return m, nil
// 		}
// 	}

// 	var cmd tea.Cmd
// 	m.list, cmd = m.list.Update(msg)
// 	return m, cmd
// }

// func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
// 	var cmd tea.Cmd

// 	switch m.currentView {
// 	case NEW_CONNECTION:
// 		m.newConnectionModel, cmd = m.newConnectionModel.Update(msg)
// 		if m.newConnectionModel.connection.status == CONNECTED {
// 			m.currentView = DATABASE_VIEW
// 			m.currentConnection = m.newConnectionModel.connection
// 			m.openDatabase = NewOpenDatabase(m.currentConnection)

// 			SaveConnectionInKeyring(m.currentConnection)
// 		}

// 		if m.newConnectionModel.action == CANCEL {
// 			m.currentView = DEFAULT
// 		}

// 	case DATABASE_VIEW:
// 		m.openDatabase, cmd = m.openDatabase.Update(msg)
// 		if m.openDatabase.viewMode == QUIT {
// 			m.currentView = DEFAULT
// 			m.openDatabase = OpenDatabase{}
// 		}

// 	case JOIN_EXISTING:
// 		m.existingConnections, cmd = m.existingConnections.Update(msg)
// 		if m.existingConnections.selectedConnection != nil {
// 			m.currentView = DATABASE_VIEW
// 			m.currentConnection = *m.existingConnections.selectedConnection
// 			m.openDatabase = NewOpenDatabase(m.currentConnection)
// 		}

// 		if m.existingConnections.back {
// 			m.currentView = DEFAULT
// 		}

// 	case DEFAULT, EDIT_CONNECTION:
// 		m, cmd = m.updateEvents(msg)
// 	}

// 	return m, cmd
// }

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

func currentConnectionInfo() *tview.List {
	list := tview.NewList().
		ShowSecondaryText(false).
		SetSelectedFocusOnly(true).
		AddItem("Name: ", "", 0, nil).
		AddItem("Host: ", "", 0, nil).
		AddItem("PORT: ", "", 0, nil).
		AddItem("USER: ", "", 0, nil).
		AddItem("Database: ", "", 0, nil)

	return list
}

func headerPanel(hotkeys *tview.Pages) *tview.Flex {
	connection := currentConnectionInfo()

	appName := tview.NewTextView().SetText(APP_NAME).SetTextAlign(tview.AlignRight)

	headerView := tview.NewFlex().
		AddItem(connection, 0, 1, false).
		AddItem(hotkeys, 0, 1, false).
		AddItem(appName, 0, 1, false)
	headerView.SetBorderPadding(0, 0, 1, 1)

	return headerView
}

func newConnectionForm(conn Connection, escapeFunc func()) *tview.Flex {
	form := tview.NewForm().
		AddInputField("Name", conn.Name, 26, nil, func(text string) { conn.Name = text }).
		AddInputField("Host", conn.Host, 26, nil, func(text string) { conn.Host = text }).
		AddInputField("Port", conn.Port, 26, func(textToCheck string, lastChar rune) bool {
			_, err := strconv.Atoi(textToCheck)
			return err == nil
		}, func(text string) { conn.Port = text }).
		AddInputField("User", conn.User, 26, nil, func(text string) { conn.User = text }).
		AddPasswordField("Password", conn.Password, 26, '*', func(text string) { conn.Password = text }).
		AddInputField("Database", conn.Database, 26, nil, func(text string) { conn.Database = text }).
		AddButton("Save", nil).
		AddButton("Test", func() {
			testResult := conn.TestConnection()
			switch testResult {
			case PASSED:
				// TODO show modal here for passed and failed jobs
			case FAILED:
			}
		}).
		AddButton("Connect", func() {
			// TODO save and connect
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

	modal.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyESC {
			escapeFunc()
		}
		return event
	})

	return modal
}

const (
	MAIN_PAGE           = "main"
	NEW_CONNECTION_FORM = "newConnection"
	SAVED_CONNECTIONS   = "savedConnections"
	DATABASE_VIEW       = "databaseView"
)

var CONNECTION_TABLE_HEADERS = []string{"NAME", "HOST", "PORT", "USER", "DATABASE"}

type DisplayTable struct {
	*tview.Table
	columns []string
	rows    []Connection
}

func newConnectionsTable(columns []string) *DisplayTable {
	table := tview.NewTable()

	for i, header := range columns {
		table.SetCell(0, i, tview.NewTableCell(header).SetExpansion(1))
	}

	table.SetBorderPadding(0, 0, 1, 1)
	table.SetSelectable(true, false).Select(1, 0)

	connectionsTable := DisplayTable{table, columns, []Connection{}}
	connectionsTable.getConnections()

	return &connectionsTable
}

func (t *DisplayTable) getConnections() {
	connections, err := ListConnections()

	if err != nil {
		return
	}

	for i, conn := range connections {
		values := conn.Row()
		for j, value := range values {
			t.SetCell(i+1, j, tview.NewTableCell(value))
		}
	}

	t.rows = connections
}

func (t *DisplayTable) getConnection() *Connection {
	row, _ := t.GetSelection()

	if row == 0 {
		return nil
	}

	return &t.rows[row-1]
}

type DbTable struct {
	*tview.Table
	table Table
}

func newDbTable(table Table) *DbTable {
	t := tview.NewTable()

	for i, header := range table.fields {
		t.SetCell(0, i, tview.NewTableCell(header).SetExpansion(1))
	}

	t.SetBorderPadding(0, 0, 1, 1)
	t.SetSelectable(true, false).Select(1, 0)

	connectionsTable := DbTable{t, table}
	connectionsTable.getTableRows()

	return &connectionsTable
}

func (t *DbTable) getTableRows() {
	for i, conn := range t.table.values {
		for j, value := range conn {
			t.SetCell(i+1, j, tview.NewTableCell(value))
		}
	}
}

type ContentBox struct {
	*tview.Box
	content tview.Primitive
}

func newContentBox(title string, content tview.Primitive) *ContentBox {
	return &ContentBox{
		tview.NewBox().SetBorder(true).SetTitle(fmt.Sprintf(" %s ", title)),
		content,
	}
}

func (b *ContentBox) Draw(screen tcell.Screen) {
	b.Box.DrawForSubclass(screen, b)
	x, y, w, h := b.GetInnerRect()

	b.content.SetRect(x, y, w, h)
	b.content.Draw(screen)
}

func (b *ContentBox) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return b.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		b.content.InputHandler()(event, setFocus)
	})
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

	hotkeyView := NewHotkeys().
		AddHotKey("New Connection", 'n').
		AddHotKey("Edit Connection", 'e').
		AddHotKey("Quit", 'q')
	hotkeyPages := tview.NewPages().AddAndSwitchToPage("connectionHotkeys", hotkeyView, true)
	header := headerPanel(hotkeyPages)

	connectionsTable := newConnectionsTable(CONNECTION_TABLE_HEADERS)
	connectionsView := newContentBox("Connections", connectionsTable)

	contentPages := tview.NewPages().
		AddAndSwitchToPage(SAVED_CONNECTIONS, connectionsView, true)
	mainView := newLayout(header, contentPages)

	mainPages := tview.NewPages().
		AddAndSwitchToPage(MAIN_PAGE, mainView, true)

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		pageName, _ := mainPages.GetFrontPage()
		// contentName, _ := mainView.content.GetFrontPage()
		currentHotkeys, _ := hotkeyPages.GetFrontPage()

		switch currentHotkeys {
		case "connectionHotkeys":
			if pageName == NEW_CONNECTION_FORM {
				return event
			}

			switch event.Rune() {
			case 'q':
				app.Stop()
			case 'n':
				if pageName == NEW_CONNECTION_FORM {
					return event
				}

				addConnectionForm := newConnectionForm(Connection{}, func() {
					mainPages.RemovePage(NEW_CONNECTION_FORM)
					app.SetFocus(contentPages)
				})
				mainPages.AddPage(NEW_CONNECTION_FORM, addConnectionForm, true, true)
				return nil
			case 'e':
				connection := connectionsTable.getConnection()
				if connection == nil {
					return event
				}

				addConnectionForm := newConnectionForm(*connection, func() {
					mainPages.RemovePage(NEW_CONNECTION_FORM)
					app.SetFocus(contentPages)
				})
				mainPages.AddPage(NEW_CONNECTION_FORM, addConnectionForm, true, true)
				return nil
			}

			switch event.Key() {
			case tcell.KeyESC:
				mainPages.RemovePage(NEW_CONNECTION_FORM)
				app.SetFocus(contentPages)
			case tcell.KeyEnter:
				connection := connectionsTable.getConnection()
				if connection == nil {
					return event
				}

				db := NewOpenDatabase(*connection)
				dbTable := newDbTable(db.openTable)
				dbContent := newContentBox(db.openTable.name, dbTable)
				contentPages.AddAndSwitchToPage(DATABASE_VIEW, dbContent, true)
			}
		default:
		}

		return event
	})

	// TODO somthing with focus is causing the extra border outline
	if err := app.SetRoot(mainPages, true).SetFocus(contentPages).Run(); err != nil {
		panic(err)
	}
}
