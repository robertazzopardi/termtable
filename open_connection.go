package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	ltb "github.com/charmbracelet/lipgloss/table"
)

var (
	unFocusedBorderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(GREY))
	focusedBorderStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color(WHITE))

	modelStyle = lipgloss.
			NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color(GREY))

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
	tables        list.Model
	viewMode      ViewMode
	selectedTable *ltb.Table
	params        Connection
	viewport      *viewport.Model
}

func NewOpenDatabase(connParams Connection, viewport *viewport.Model) OpenDatabase {
	databaseTables := connParams.GetTableNames()

	listItems := []list.Item{}
	for _, value := range databaseTables {
		listItems = append(listItems, tableItem(value))
	}

	openDatabase := OpenDatabase{
		tables:   list.New(listItems, tableItemDelegate{}, 14, 14),
		viewMode: TABLES,
		params:   connParams,
		viewport: viewport,
	}

	openDatabase.tables.SetShowHelp(false)
	openDatabase.tables.SetShowTitle(false)
	openDatabase.tables.SetShowStatusBar(false)

	openDatabase.setOpenTable()

	return openDatabase
}

func (db *OpenDatabase) setOpenTable() {
	selectedItem := db.tables.SelectedItem()
	tableName := string(selectedItem.(tableItem))

	selectedTable, err := db.openTable(tableName)

	if err != nil {
		db.params.status = DISCONNECTED
		return
	}

	db.selectedTable = selectedTable
	// db.selectedTable.SetWidth(db.viewport.Width / 2)
	// db.selectedTable.SetHeight(db.viewport.Height / 2)
}

func (db OpenDatabase) openTable(tableName string) (*ltb.Table, error) {
	tableData, err := db.params.SelectAll(tableName)

	if err != nil {
		return db.selectedTable, err
	}

	// max_len := db.selectedTable.Width() / len(tableData.fields)
	// columns := make([]table.Column, len(tableData.fields))
	cols := make([]string, len(tableData.fields))
	for i, field := range tableData.fields {
		// columns[i] = table.Column{Title: field, Width: max_len}
		cols[i] = field
	}

	rows := make([]table.Row, len(tableData.values))
	rowss := make([][]string, len(tableData.values))
	for i, value := range tableData.values {
		rows[i] = make(table.Row, len(value))
		copy(rows[i], value)
		rowss[i] = make([]string, len(value))
		copy(rowss[i], value)
	}

	t := ltb.New().
		Headers(cols...).
		Rows(rowss...)

	// s := table.DefaultStyles()
	// s.Header = s.Header.
	// 	BorderStyle(lipgloss.NormalBorder()).
	// 	BorderForeground(lipgloss.Color("240")).
	// 	BorderBottom(true).
	// 	Bold(false)
	// s.Selected = s.Selected.
	// 	Foreground(lipgloss.Color("229")).
	// 	Background(lipgloss.Color("57")).
	// 	Bold(false)
	// t.SetStyles(s)

	return t, nil
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

	switch db.viewMode {
	case TABLES:
		db.tables, cmd = db.tables.Update(msg)
	case OPEN:
		// db.selectedTable, cmd = db.selectedTable.Update(msg)
	}

	return db, cmd
}

func (db OpenDatabase) View() string {
	s := fmt.Sprintf("%s / %s\n\n", db.params.Name, db.params.Database)

	tableLabels := db.tables.View()

	db.setOpenTable()
	openTable := db.selectedTable //.View()

	if db.viewMode == TABLES {
		s += lipgloss.JoinHorizontal(lipgloss.Top,
			focusedModelSideBarStyle.Render(tableLabels),
			openTable.BorderStyle(unFocusedBorderStyle).String())
	} else {
		s += lipgloss.JoinHorizontal(lipgloss.Top,
			modelStyle.Render(tableLabels),
			openTable.BorderStyle(focusedBorderStyle).String())
	}

	return paginationStyle.Render(s)
}
