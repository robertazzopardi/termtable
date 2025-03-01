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
				app.pages.RemovePage("ConfirmDelete")
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
