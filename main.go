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

func main() {
	app := NewApp()

	if err := app.Run(); err != nil {
		panic(err)
	}
}
