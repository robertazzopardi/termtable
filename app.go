package main

import (
	"log"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type App struct {
	*tview.Application
	conn        Connection
	pages       *tview.Pages
	content     *tview.Pages
	hotkeys     *tview.Pages
	connections *DisplayTable
}

func NewApp() App {
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

		// Skip handling if we're in the connection form
		if pageName == NEW_CONNECTION_FORM {
			return event
		}

		// Handle connection hotkeys
		if currentHotkeys == "connectionHotkeys" {
			// Handle rune-based hotkeys
			switch event.Rune() {
			case 'q':
				app.Stop()
				return nil
			case 'n':
				app.showNewConnectionForm(NewConnection())
				return nil
			case 'e':
				connection := app.connections.getConnection()
				if connection != nil {
					app.showNewConnectionForm(*connection)
				}
				return nil
			case 'd':
				connection := app.connections.getConnection()
				if connection != nil {
					app.showDeleteConfirmation(*connection)
				}
				return nil
			}

			// Handle special keys
			switch event.Key() {
			case tcell.KeyESC:
				if contentView == DATABASE_VIEW {
					app.returnToConnectionsView()
					return nil
				}
				app.closeModals()
				return nil
			case tcell.KeyEnter:
				if contentView != DATABASE_VIEW && pageName != CONFIRM_DELETE && pageName != CONNECTION_TEST {
					connection := app.connections.getConnection()
					if connection != nil {
						app.openConnection(*connection)
					}
				}
				return event
			}
		}

		return event
	})
}

// Helper methods to clean up the input handler
func (app *App) showNewConnectionForm(connection Connection) {
	addConnectionForm := newConnectionForm(app, connection, func() {
		app.pages.RemovePage(NEW_CONNECTION_FORM)
		app.SetFocus(app.content)
	})
	app.pages.AddPage(NEW_CONNECTION_FORM, addConnectionForm, true, true)
}

func (app *App) showDeleteConfirmation(connection Connection) {
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
			app.pages.RemovePage(CONFIRM_DELETE)
			app.SetFocus(app.content)
		})
	app.pages.AddPage(CONFIRM_DELETE, confirmDeleteModal, true, true)
}

func (app *App) showInfoModal(page, body string, escapeFunc func()) {
	confirmDeleteModal := tview.NewModal().
		SetText(body).
		AddButtons([]string{"Ok"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			escapeFunc()
		})

	app.pages.AddPage(page, confirmDeleteModal, true, true)
}

func (app *App) returnToConnectionsView() {
	newHeader := newLayout(tview.FlexRow, headerPanel(NewConnection(), app.hotkeys), app.content)
	app.pages.AddPage(SAVED_CONNECTIONS, newHeader, true, true)
	app.content.RemovePage(DATABASE_VIEW)
	app.SetFocus(app.content)
}

func (app *App) closeModals() {
	app.pages.RemovePage(NEW_CONNECTION_FORM)
	app.pages.RemovePage(CONFIRM_DELETE)
	app.SetFocus(app.content)
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
