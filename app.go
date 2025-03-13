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
	tview.Borders.HorizontalFocus = tview.BoxDrawingsLightHorizontal
	tview.Borders.VerticalFocus = tview.BoxDrawingsLightVertical
	tview.Borders.TopLeftFocus = tview.BoxDrawingsLightDownAndRight
	tview.Borders.TopRightFocus = tview.BoxDrawingsLightDownAndLeft
	tview.Borders.BottomLeftFocus = tview.BoxDrawingsLightUpAndRight
	tview.Borders.BottomRightFocus = tview.BoxDrawingsLightUpAndLeft

	app := tview.NewApplication()

	connectionHotkeys := GetConnectionHotkeys()
	databaseHotkeys := GetDatabaseHotkeys()

	hotkeys := tview.NewPages().
		AddPage("databaseHotkeys", databaseHotkeys, true, false).
		AddAndSwitchToPage("connectionHotkeys", connectionHotkeys, true)

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

	searchBar := tview.NewInputField().
		SetFieldWidth(0).
		SetAcceptanceFunc(tview.InputFieldInteger).
		SetDoneFunc(func(key tcell.Key) {
			app.Stop()
		})
	searchBar.SetBorder(true)
	searchBar.SetFieldBackgroundColor(tcell.ColorNone)
	container := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(searchBar, 3, 0, false).
		AddItem(connectionsView, 0, 1, false)

	container.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		connectionsView.InputHandler()(event, func(p tview.Primitive) {})
		return event
	})

	app.content.AddPage(SAVED_CONNECTIONS, container, true, true)
	app.connections = connectionsTable
}

func (app *App) setInputHandler() {
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		pageName, _ := app.pages.GetFrontPage()
		currentHotkeys, _ := app.hotkeys.GetFrontPage()
		contentView, _ := app.content.GetFrontPage()

		if pageName == NEW_CONNECTION_FORM {
			return event
		}

		if currentHotkeys == "connectionHotkeys" {
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
			case 'o':
				connection := app.connections.getConnection()
				if connection != nil {
					app.openConnection(*connection)
				}
				return nil
			case 't':
				connection := app.connections.getConnection()
				if connection != nil {
					testResult := connection.TestConnection()
					closeModal := func() {
						app.pages.RemovePage(CONNECTION_TEST)
					}

					switch testResult {
					case PASSED:
						app.showInfoModal(CONNECTION_TEST, "Connection successful!", closeModal)
					case FAILED:
						app.showInfoModal(CONNECTION_TEST, "Connection failed. Please check your settings.", closeModal)
					}
				}
				return nil
			case 'r':
				app.refreshConnections()
				return nil
			case '/':
				app.showSearchInput()
				return nil
			case 's':
				app.sortConnectionsByName()
				return nil
			case '?':
				app.showHelpView()
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
		} else if currentHotkeys == "databaseHotkeys" {
			// Handle database view hotkeys
			switch event.Rune() {
			case 'q':
				app.Stop()
				return nil
			case 'b':
				app.returnToConnectionsView()
				return nil
			case 'r':
				app.refreshCurrentConnection()
				return nil
			case 'e':
				app.showQueryEditor()
				return nil
			case 'x':
				app.showExportOptions()
				return nil
			case 'f':
				app.showFilterInput()
				return nil
			case 'c':
				app.copySelectedRow()
				return nil
			case 'y':
				app.copySelectedCell()
				return nil
			case 'n':
				app.goToNextPage()
				return nil
			case 'p':
				app.goToPreviousPage()
				return nil
			case 'v':
				app.toggleViewMode()
				return nil
			case '?':
				app.showHelpView()
				return nil
			}

			// Handle special keys
			switch event.Key() {
			case tcell.KeyESC:
				app.returnToConnectionsView()
				return nil
			}
		} else if currentHotkeys == "helpHotkeys" {
			// Handle help view hotkeys
			switch event.Rune() {
			case 'q':
				app.Stop()
				return nil
			case 'b':
				app.closeHelpView()
				return nil
			}

			// Handle special keys
			switch event.Key() {
			case tcell.KeyESC:
				app.closeHelpView()
				return nil
			}
		} else if currentHotkeys == "queryHotkeys" {
			// Handle query editor hotkeys
			switch event.Key() {
			case tcell.KeyESC:
				app.closeQueryEditor()
				return nil
			case tcell.KeyCtrlE:
				app.executeQuery()
				return nil
			case tcell.KeyCtrlS:
				app.saveQuery()
				return nil
			case tcell.KeyCtrlO:
				app.loadQuery()
				return nil
			case tcell.KeyCtrlL:
				app.clearQuery()
				return nil
			case tcell.KeyCtrlH:
				app.showQueryHistory()
				return nil
			}
		} else if currentHotkeys == "exportHotkeys" {
			// Handle export options hotkeys
			switch event.Rune() {
			case 'c':
				app.exportAsCSV()
				return nil
			case 'j':
				app.exportAsJSON()
				return nil
			case 's':
				app.exportAsSQL()
				return nil
			}

			// Handle special keys
			switch event.Key() {
			case tcell.KeyESC:
				app.closeExportOptions()
				return nil
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

func (app *App) refreshCurrentConnection() {
	connection := app.connections.getConnection()
	if connection != nil {
		app.openConnection(*connection)
	}
}

func (app *App) returnToConnectionsView() {
	// Switch back to connection hotkeys
	app.hotkeys.SwitchToPage("connectionHotkeys")

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

	app.hotkeys.SwitchToPage("databaseHotkeys")

	layoutHeader := headerPanel(connection, app.hotkeys)
	newHeader := newLayout(tview.FlexRow, layoutHeader, app.content)
	app.pages.AddPage(SAVED_CONNECTIONS, newHeader, true, true)

	app.content.AddAndSwitchToPage(DATABASE_VIEW, dbContent, true)
	app.SetFocus(app.content)
}

// Search functionality
func (app *App) showSearchInput() {
	var inputField *tview.InputField
	inputField = tview.NewInputField().
		SetLabel("Search: ").
		SetFieldWidth(30).
		SetDoneFunc(func(key tcell.Key) {
			if key == tcell.KeyEnter {
				searchTerm := inputField.GetText()
				app.searchConnections(searchTerm)
				app.pages.RemovePage("searchInput")
				app.SetFocus(app.content)
			} else if key == tcell.KeyEscape {
				app.pages.RemovePage("searchInput")
				app.SetFocus(app.content)
			}
		})

	modal := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(inputField, 3, 1, true).
			AddItem(nil, 0, 1, false), 40, 1, true).
		AddItem(nil, 0, 1, false)

	app.pages.AddPage("searchInput", modal, true, true)
	app.SetFocus(inputField)
}

func (app *App) searchConnections(term string) {
	// Implementation would filter the connections table based on the search term
	// This is a placeholder - actual implementation would depend on your data structure
}

// Sorting functionality
func (app *App) sortConnectionsByName() {
	// Implementation would sort the connections table by name
	// This is a placeholder - actual implementation would depend on your data structure
}

// Help view
func (app *App) showHelpView() {
	helpText := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWordWrap(true).
		SetText(`[yellow]TermTable Help[white]

[green]Connection View Hotkeys:[white]
<n> New Connection - Create a new database connection
<e> Edit Connection - Edit the selected connection
<d> Delete Connection - Delete the selected connection
<o> Open Connection - Open the selected connection
<t> Test Connection - Test the selected connection
<r> Refresh Connections - Refresh the list of connections
</> Search - Search for a connection
<s> Sort by Name - Sort connections by name
<h> Help - Show this help screen
<q> Quit - Exit the application

[green]Database View Hotkeys:[white]
<b> Back to Connections - Return to the connections view
<r> Refresh Data - Refresh the current data view
<e> Execute Query - Open the query editor
<x> Export Results - Export the current results
<f> Filter Results - Filter the current results
<c> Copy Row - Copy the selected row
<y> Copy Cell - Copy the selected cell
<n> Next Page - Go to the next page of results
<p> Previous Page - Go to the previous page of results
<v> Toggle View Mode - Toggle between different view modes
<h> Help - Show this help screen
<q> Quit - Exit the application

[green]Navigation:[white]
Use arrow keys to navigate tables and lists.
Press Enter to select or open an item.
Press Esc to go back or close a modal.`)

	helpView := tview.NewFrame(helpText).
		SetBorders(0, 0, 0, 0, 0, 0).
		AddText("Help", true, tview.AlignCenter, tcell.ColorYellow).
		AddText("Press 'b' to go back", false, tview.AlignCenter, tcell.ColorWhite)

	// Create help hotkeys
	helpHotkeys := GetHelpHotkeys()
	app.hotkeys.AddAndSwitchToPage("helpHotkeys", helpHotkeys, true)

	// Save current view to return to it later
	app.pages.AddPage("helpView", helpView, true, true)
	app.SetFocus(helpView)
}

func (app *App) closeHelpView() {
	app.pages.RemovePage("helpView")

	// Switch back to previous hotkeys
	contentView, _ := app.content.GetFrontPage()
	if contentView == DATABASE_VIEW {
		app.hotkeys.SwitchToPage("databaseHotkeys")
	} else {
		app.hotkeys.SwitchToPage("connectionHotkeys")
	}

	app.SetFocus(app.content)
}

// Query editor functionality
func (app *App) showQueryEditor() {
	queryEditor := tview.NewTextArea().
		SetPlaceholder("Enter SQL query here...").
		SetWordWrap(true)

	queryFrame := tview.NewFrame(queryEditor).
		SetBorders(0, 0, 0, 0, 0, 0).
		AddText("SQL Query Editor", true, tview.AlignCenter, tcell.ColorYellow).
		AddText("Ctrl+E: Execute | Ctrl+S: Save | Ctrl+O: Load | Ctrl+L: Clear | Esc: Exit", false, tview.AlignCenter, tcell.ColorWhite)

	// Create query hotkeys
	queryHotkeys := GetQueryHotkeys()
	app.hotkeys.AddAndSwitchToPage("queryHotkeys", queryHotkeys, true)

	app.pages.AddPage("queryEditor", queryFrame, true, true)
	app.SetFocus(queryEditor)
}

func (app *App) closeQueryEditor() {
	app.pages.RemovePage("queryEditor")
	app.hotkeys.SwitchToPage("databaseHotkeys")
	app.SetFocus(app.content)
}

func (app *App) executeQuery() {
	// Implementation would execute the query in the editor
	// This is a placeholder
}

func (app *App) saveQuery() {
	// Implementation would save the current query
	// This is a placeholder
}

func (app *App) loadQuery() {
	// Implementation would load a saved query
	// This is a placeholder
}

func (app *App) clearQuery() {
	// Implementation would clear the query editor
	// This is a placeholder
}

func (app *App) showQueryHistory() {
	// Implementation would show query history
	// This is a placeholder
}

// Export functionality
func (app *App) showExportOptions() {
	exportModal := tview.NewModal().
		SetText("Export Options").
		AddButtons([]string{"CSV", "JSON", "SQL", "Cancel"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			switch buttonLabel {
			case "CSV":
				app.exportAsCSV()
			case "JSON":
				app.exportAsJSON()
			case "SQL":
				app.exportAsSQL()
			}
			app.pages.RemovePage("exportOptions")
			app.SetFocus(app.content)
		})

	// Create export hotkeys
	exportHotkeys := GetExportHotkeys()
	app.hotkeys.AddAndSwitchToPage("exportHotkeys", exportHotkeys, true)

	app.pages.AddPage("exportOptions", exportModal, true, true)
}

func (app *App) closeExportOptions() {
	app.pages.RemovePage("exportOptions")
	app.hotkeys.SwitchToPage("databaseHotkeys")
	app.SetFocus(app.content)
}

func (app *App) exportAsCSV() {
	// Implementation would export data as CSV
	// This is a placeholder
}

func (app *App) exportAsJSON() {
	// Implementation would export data as JSON
	// This is a placeholder
}

func (app *App) exportAsSQL() {
	// Implementation would export data as SQL
	// This is a placeholder
}

// Filter functionality
func (app *App) showFilterInput() {
	var inputField *tview.InputField
	inputField = tview.NewInputField().
		SetLabel("Filter: ").
		SetFieldWidth(30).
		SetDoneFunc(func(key tcell.Key) {
			if key == tcell.KeyEnter {
				filterTerm := inputField.GetText()
				app.filterResults(filterTerm)
				app.pages.RemovePage("filterInput")
				app.SetFocus(app.content)
			} else if key == tcell.KeyEscape {
				app.pages.RemovePage("filterInput")
				app.SetFocus(app.content)
			}
		})

	modal := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(inputField, 3, 1, true).
			AddItem(nil, 0, 1, false), 40, 1, true).
		AddItem(nil, 0, 1, false)

	app.pages.AddPage("filterInput", modal, true, true)
	app.SetFocus(inputField)
}

func (app *App) filterResults(term string) {
	// Implementation would filter the results based on the filter term
	// This is a placeholder
}

// Copy functionality
func (app *App) copySelectedRow() {
	// Implementation would copy the selected row to clipboard
	// This is a placeholder
}

func (app *App) copySelectedCell() {
	// Implementation would copy the selected cell to clipboard
	// This is a placeholder
}

// Pagination functionality
func (app *App) goToNextPage() {
	// Implementation would go to the next page of results
	// This is a placeholder
}

func (app *App) goToPreviousPage() {
	// Implementation would go to the previous page of results
	// This is a placeholder
}

// View mode functionality
func (app *App) toggleViewMode() {
	// Implementation would toggle between different view modes
	// This is a placeholder
}
