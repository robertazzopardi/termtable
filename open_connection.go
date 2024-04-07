package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/table"
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
	tables        list.Model
	viewMode      ViewMode
	selectedTable table.Model
	params        Connection
}

func NewOpenDatabase(connParams Connection) OpenDatabase {
	databaseTables := connParams.GetTableNames()

	listItems := []list.Item{}
	for _, value := range databaseTables {
		listItems = append(listItems, tableItem(value))
	}

	openDatabase := OpenDatabase{
		tables:   list.New(listItems, tableItemDelegate{}, 14, 14),
		viewMode: TABLES,
		params:   connParams,
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
}

func (db OpenDatabase) openTable(tableName string) (table.Model, error) {
	tableData, err := db.params.SelectAll(tableName)

	if err != nil {
		return table.Model{}, err
	}

	columns := make([]table.Column, len(tableData.fields))
	for i, field := range tableData.fields {
		columns[i] = table.Column{Title: field, Width: len(field) + 2}
	}

	rows := make([]table.Row, len(tableData.values))
	for i, value := range tableData.values {
		rows[i] = make(table.Row, len(value))
		copy(rows[i], value)
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(7),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	return t, nil
}

func (db OpenDatabase) Init() tea.Cmd {
	return nil
}

func (db OpenDatabase) Update(msg tea.Msg) (OpenDatabase, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			db.viewMode = QUIT

		case "left", "right":
			switch db.viewMode {
			case TABLES:
				db.viewMode = OPEN
			case OPEN:
				db.viewMode = TABLES
			}
		}
	case tea.WindowSizeMsg:
		db.tables.SetWidth(msg.Width)
		db.tables.SetHeight(msg.Height)
		return db, nil
	}

	var cmd tea.Cmd

	switch db.viewMode {
	case TABLES:
		db.tables, cmd = db.tables.Update(msg)
		db.setOpenTable()
	case OPEN:
		db.selectedTable, cmd = db.selectedTable.Update(msg)
	}

	return db, cmd
}

func (db OpenDatabase) View() string {
	s := fmt.Sprintf("%s\n\n", db.params.Name)

	tableLabels := db.tables.View()

	openTable := db.selectedTable.View()

	if db.viewMode == TABLES {
		s += lipgloss.JoinHorizontal(lipgloss.Top,
			focusedModelSideBarStyle.Render(tableLabels),
			modelStyle.Render(openTable))
	} else {
		s += lipgloss.JoinHorizontal(lipgloss.Top,
			modelStyle.Render(tableLabels),
			focusedModelStyle.Render(openTable))
	}

	return s
}
