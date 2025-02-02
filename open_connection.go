package main

import (
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	// "github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	modelStyle = lipgloss.
			NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color(GREY))
	focusedModelStyle = lipgloss.
				NewStyle().
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color(WHITE))

	focusedModelSideBarStyle = lipgloss.
					NewStyle().
					BorderStyle(lipgloss.NormalBorder()).
					BorderForeground(lipgloss.Color(WHITE))

	blurredModelSideBarStyle = lipgloss.
					NewStyle().
					Foreground(lipgloss.Color(GREY))
	selectedTableStyle = lipgloss.
				NewStyle().
				Foreground(lipgloss.Color(MAGENTA))
)

type ViewMode string

const (
	TABLES ViewMode = "TABLES"
	OPEN   ViewMode = "OPEN"
	QUIT   ViewMode = "QUIT"
)

type tableItem string

func (i tableItem) FilterValue() string { return "" }

type tableItemDelegate struct{}

func (d tableItemDelegate) Height() int                             { return 1 }
func (d tableItemDelegate) Spacing() int                            { return 0 }
func (d tableItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d tableItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	str, ok := listItem.(tableItem)
	if !ok {
		return
	}

	fn := blurredModelSideBarStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedTableStyle.Render(strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(string(str)))
}

type OpenDatabase struct {
	tables   []string
	viewMode ViewMode
	// selectedTable table.Model
	params    Connection
	openTable Table
}

func NewOpenDatabase(connParams Connection) OpenDatabase {
	databaseTables := connParams.GetTableNames()

	// listItems := []list.Item{}
	// for _, value := range databaseTables {
	// 	listItems = append(listItems, tableItem(value))
	// }

	openDatabase := OpenDatabase{
		// tables:   list.New(listItems, tableItemDelegate{}, 14, 14),
		tables:   databaseTables,
		viewMode: TABLES,
		params:   connParams,
	}

	// openDatabase.tables.SetShowHelp(false)
	// openDatabase.tables.SetShowTitle(false)
	// openDatabase.tables.SetShowStatusBar(false)

	openDatabase.setOpenTable()

	return openDatabase
}

func (db *OpenDatabase) setOpenTable() {
	tableName := db.tables[0]

	table, err := db.params.SelectAll(tableName)
	// selectedTable, err := db.openTable(tableName)

	if err != nil {
		// db.params.status = DISCONNECTED
		log.Fatal("Could not connect to db", err)
		return
	}

	db.openTable = table
}

// func (db OpenDatabase) openTable(tableName string) (Table, error) {
// 	return db.params.SelectAll(tableName)

// 	// if err != nil {
// 	// 	return Table{}, err
// 	// }

// 	// return tableData, nil
// }

func (db OpenDatabase) Init() tea.Cmd {
	return nil
}

func (db OpenDatabase) Update(msg tea.Msg) (OpenDatabase, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			db.viewMode = QUIT
			return db, nil

		case "left", "right":
			switch db.viewMode {
			case TABLES:
				db.viewMode = OPEN
			case OPEN:
				db.viewMode = TABLES
			}
		}
	}

	var cmd tea.Cmd

	switch db.viewMode {
	case TABLES:
		// db.tables, cmd = db.tables.Update(msg)
	case OPEN:
		// db.selectedTable, cmd = db.selectedTable.Update(msg)
	}

	return db, cmd
}

func (db OpenDatabase) View() string {
	s := fmt.Sprintf("%s / %s\n\n", db.params.Name, db.params.Database)

	// tableLabels := db.tables.View()

	// db.setOpenTable()
	// openTable := db.selectedTable.View()

	// if db.viewMode == TABLES {
	// 	s += lipgloss.JoinHorizontal(lipgloss.Top,
	// 		focusedModelSideBarStyle.Render(tableLabels),
	// 		modelStyle.Render(openTable))
	// } else {
	// 	s += lipgloss.JoinHorizontal(lipgloss.Top,
	// 		modelStyle.Render(tableLabels),
	// 		focusedModelStyle.Render(openTable))
	// }

	return paginationStyle.Render(s)
}
