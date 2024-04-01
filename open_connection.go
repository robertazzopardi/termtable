package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	modelStyle = lipgloss.
			NewStyle().
			Align(lipgloss.Center, lipgloss.Center).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color(GREY))
	focusedModelStyle = lipgloss.
				NewStyle().
				Align(lipgloss.Center, lipgloss.Center).
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color(WHITE))

	modelSideBarStyle = lipgloss.
				NewStyle().
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color(GREY))
	focusedModelSideBarStyle = lipgloss.
					NewStyle().
					BorderStyle(lipgloss.NormalBorder()).
					BorderForeground(lipgloss.Color(WHITE))

	selectedTableStyle = lipgloss.
				NewStyle().
				Foreground(lipgloss.Color(MAGENTA))
)

type ViewMode string

const (
	COLUMNS ViewMode = "COLUMNS"
	OPEN    ViewMode = "OPEN"
	QUIT    ViewMode = "QUIT"
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

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedTableStyle.Render(strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(string(str)))
}

type OpenDatabase struct {
	tables   list.Model
	viewMode ViewMode
	name     string
}

func NewOpenDatabase(connParams *Connection) OpenDatabase {
	databaseTables := connParams.GetTableNames()

	listItems := []list.Item{}
	for _, value := range databaseTables {
		listItems = append(listItems, tableItem(value))
	}

	openDatabase := OpenDatabase{
		tables:   list.New(listItems, tableItemDelegate{}, 14, 14),
		viewMode: COLUMNS,
		name:     connParams.Name,
	}

	openDatabase.tables.SetShowHelp(false)
	openDatabase.tables.SetShowTitle(false)
	openDatabase.tables.SetShowStatusBar(false)

	return openDatabase
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
			case COLUMNS:
				db.viewMode = OPEN
			case OPEN:
				db.viewMode = COLUMNS
			}
		}
	case tea.WindowSizeMsg:
		db.tables.SetWidth(msg.Width)
		db.tables.SetHeight(msg.Height)
		return db, nil
	}

	var cmd tea.Cmd
	db.tables, cmd = db.tables.Update(msg)
	return db, cmd
}

func (db OpenDatabase) View() string {
	s := fmt.Sprintf("%s\n\n", db.name)

	tableLabels := db.tables.View()

	if db.viewMode == COLUMNS {
		s += lipgloss.JoinHorizontal(lipgloss.Top,
			focusedModelSideBarStyle.Render(fmt.Sprintf("%4s", tableLabels)),
			modelStyle.Render("Table"))
	} else {
		s += lipgloss.JoinHorizontal(lipgloss.Top,
			modelStyle.Render(fmt.Sprintf("%4s", tableLabels)),
			focusedModelStyle.Render("Table"))
	}

	return s
}
