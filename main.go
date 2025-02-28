package main

import (
	"fmt"
	"log"
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
	// DATABASE_VIEW   CurrentView = "DATABASE_VIEW".
)

// Primary ansi colours.
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

func currentConnectionInfo(conn Connection) *tview.List {
	list := tview.NewList().
		ShowSecondaryText(false).
		SetSelectedFocusOnly(true).
		AddItem("Name: "+conn.Name, "", 0, nil).
		AddItem("Host: "+conn.Host, "", 0, nil).
		AddItem("PORT: "+conn.Port, "", 0, nil).
		AddItem("USER: "+conn.User, "", 0, nil).
		AddItem("Database: "+conn.Database, "", 0, nil)

	return list
}

func headerPanel(conn Connection, hotkeys *tview.Pages) *tview.Flex {
	connection := currentConnectionInfo(conn)

	appName := tview.NewTextView().SetText(APP_NAME).SetTextAlign(tview.AlignRight)

	headerView := tview.NewFlex().
		AddItem(connection, 0, 1, false).
		AddItem(hotkeys, 0, 1, false).
		AddItem(appName, 0, 1, false)
	headerView.SetBorderPadding(0, 0, 1, 1)

	return headerView
}

func newConnectionForm(app *App, conn Connection, escapeFunc func()) *tview.Flex {
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
		AddButton("Save", func() {
			testResult := conn.TestConnection()
			if testResult == FAILED {
				log.Fatal("Could not save connection because connection could not be established")
			}

			SaveConnectionInKeyring(conn)

			escapeFunc()

			app.refreshConnections()
		}).
		AddButton("Test", func() {
			testResult := conn.TestConnection()
			switch testResult {
			case PASSED:
				// TODO show modal here for passed and failed jobs
			case FAILED:
			}
		}).
		AddButton("Connect", func() {
			testResult := conn.TestConnection()
			if testResult == FAILED {
				log.Fatal("Could not connect because connection could not be established")
			}

			escapeFunc()

			app.refreshConnections()

			app.openConnection(conn)
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

var CONNECTION_TABLE_HEADERS = []string{"ID", "NAME", "HOST", "PORT", "USER", "DATABASE"}

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

func newLayout(direction int, header *tview.Flex, content *tview.Pages) Layout {
	view := tview.NewFlex().SetDirection(direction).
		AddItem(header, 0, 1, false).
		AddItem(content, 0, 6, false)

	return Layout{view, header, content}
}

type App struct {
	*tview.Application
	conn        Connection
	pages       *tview.Pages
	content     *tview.Pages
	hotkeys     *tview.Pages
	connections *DisplayTable
}

func newApp() App {
	app := tview.NewApplication()

	hotkeyView := NewHotkeys().
		AddHotKey("New Connection", 'n').
		AddHotKey("Edit Connection", 'e').
		AddHotKey("Delete Connection", 'd').
		AddHotKey("Quit", 'q')
	hotkeys := tview.NewPages().
		AddAndSwitchToPage("connectionHotkeys", hotkeyView, true)
	header := headerPanel(NewConnection(), hotkeys)

	content := tview.NewPages()
	mainView := newLayout(tview.FlexRow, header, content)

	pages := tview.NewPages().
		AddAndSwitchToPage(MAIN_PAGE, mainView, true)

	app.SetRoot(pages, true).SetFocus(content)

	ctx := App{app, NewConnection(), pages, content, hotkeys, &DisplayTable{}}
	ctx.setInputHandler()
	ctx.refreshConnections()

	return ctx
}

func (app *App) refreshConnections() {
	connectionsTable := newConnectionsTable(CONNECTION_TABLE_HEADERS)
	connectionsView := newContentBox("Connections", connectionsTable)
	app.content.AddPage(SAVED_CONNECTIONS, connectionsView, true, true)
	app.connections = connectionsTable
}

func (app *App) setInputHandler() {
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		pageName, _ := app.pages.GetFrontPage()
		currentHotkeys, _ := app.hotkeys.GetFrontPage()
		contentView, _ := app.content.GetFrontPage()

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

				addConnectionForm := newConnectionForm(app, NewConnection(), func() {
					app.pages.RemovePage(NEW_CONNECTION_FORM)
					app.SetFocus(app.content)
				})
				app.pages.AddPage(NEW_CONNECTION_FORM, addConnectionForm, true, true)

				return nil
			case 'e':
				connection := app.connections.getConnection()
				if connection == nil {
					return event
				}

				addConnectionForm := newConnectionForm(app, *connection, func() {
					app.pages.RemovePage(NEW_CONNECTION_FORM)
					app.SetFocus(app.content)
				})
				app.pages.AddPage(NEW_CONNECTION_FORM, addConnectionForm, true, true)

				return nil
			case 'd':
				// Delete connection
				connection := app.connections.getConnection()
				if connection == nil {
					return event
				}

				confirmDeleteModal := tview.NewModal().
					SetText("Are you sure you want to delete: " + connection.Name + "?").
					AddButtons([]string{"Cancel", "Confirm"}).
					SetDoneFunc(func(buttonIndex int, buttonLabel string) {
						if buttonLabel == "Confirm" {
							err := DeleteConnection(connection.Name)
							if err != nil {
								log.Fatal("Could not delete connection", err)
							}

							app.refreshConnections()
						}

						app.pages.RemovePage("ConfirmDelete")
						app.SetFocus(app.content)
					})
				app.pages.AddPage("ConfirmDelete", confirmDeleteModal, true, true)
				return nil
			}

			switch event.Key() {
			case tcell.KeyESC:
				if contentView == DATABASE_VIEW {
					newHeader := newLayout(tview.FlexRow, headerPanel(NewConnection(), app.hotkeys), app.content)
					app.pages.AddPage(SAVED_CONNECTIONS, newHeader, true, true)

					app.content.RemovePage(DATABASE_VIEW)
					app.SetFocus(app.content)

					return event
				}

				app.pages.RemovePage(NEW_CONNECTION_FORM)
				app.SetFocus(app.content)
			case tcell.KeyEnter:
				if contentView == DATABASE_VIEW || pageName == "ConfirmDelete" {
					return event
				}

				connection := app.connections.getConnection()
				if connection == nil {
					return event
				}

				app.openConnection(*connection)
			}
		default:
		}

		return event
	})
}

func (app App) openConnection(connection Connection) {
	db := NewOpenDatabase(connection)
	dbTable := newDbTable(db.openTable)
	dbContent := newContentBox(db.openTable.name, dbTable)

	newHeader := newLayout(tview.FlexRow, headerPanel(connection, app.hotkeys), app.content)
	app.pages.AddPage(SAVED_CONNECTIONS, newHeader, true, true)

	app.content.AddAndSwitchToPage(DATABASE_VIEW, dbContent, true)
	app.SetFocus(app.content)
}

func main() {
	app := newApp()

	if err := app.Run(); err != nil {
		panic(err)
	}
}
