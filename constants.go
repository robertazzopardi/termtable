package main

// CurrentView represents the current view state of the application
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

type PageName string

const (
	MAIN_PAGE           = "main"
	NEW_CONNECTION_FORM = "newConnection"
	SAVED_CONNECTIONS   = "savedConnections"
	DATABASE_VIEW       = "databaseView"
	CONNECTION_TEST     = "ConntectionTest"
	CONFIRM_DELETE      = "ConfirmDelete"
)

var CONNECTION_TABLE_HEADERS = []string{"ID", "NAME", "HOST", "PORT", "USER", "DATABASE"}
