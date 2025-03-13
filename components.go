package main

import (
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

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
				app.showInfoModal(CONNECTION_TEST, "Could not save connection because connection could not be established.", func() {
					app.pages.RemovePage(CONNECTION_TEST)
				})
				return
			}

			SaveConnectionInKeyring(conn)

			escapeFunc()

			app.refreshConnections()
		}).
		AddButton("Test", func() {
			testResult := conn.TestConnection()
			closeModal := func() {
				app.pages.RemovePage(CONNECTION_TEST)
			}

			switch testResult {
			case PASSED:
				app.showInfoModal(CONNECTION_TEST, "Connection successful!", closeModal)
			case FAILED:
				app.showInfoModal(CONNECTION_TEST, "Connection failed. Please check your settings.", closeModal)
			}
		}).
		AddButton("Connect", func() {
			testResult := conn.TestConnection()
			if testResult == FAILED {
				app.showInfoModal(CONNECTION_TEST, "Could not connect because connection could not be established", func() {
					app.pages.RemovePage(CONNECTION_TEST)
				})
				return
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
