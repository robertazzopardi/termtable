package main

import (
	"fmt"
	"log"

	tea "github.com/charmbracelet/bubbletea"
)

type ViewMode string

const (
	TABLES ViewMode = "TABLES"
	OPEN   ViewMode = "OPEN"
	QUIT   ViewMode = "QUIT"
)

type OpenDatabase struct {
	tables    []string
	viewMode  ViewMode
	params    Connection
	openTable Table
}

func NewOpenDatabase(connParams Connection) OpenDatabase {
	databaseTables := connParams.GetTableNames()

	openDatabase := OpenDatabase{
		tables:   databaseTables,
		viewMode: TABLES,
		params:   connParams,
	}

	openDatabase.setOpenTable()

	return openDatabase
}

func (db *OpenDatabase) setOpenTable() {
	tableName := db.tables[0]

	table, err := db.params.SelectAll(tableName)

	if err != nil {
		log.Fatal("Could not connect to db", err)
		return
	}

	db.openTable = table
}

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
