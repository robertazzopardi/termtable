package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	modelStyle = lipgloss.NewStyle().
			Width(15).
			Height(5).
			Align(lipgloss.Center, lipgloss.Center).
			BorderStyle(lipgloss.HiddenBorder())
	focusedModelStyle = lipgloss.NewStyle().
				Width(15).
				Height(5).
				Align(lipgloss.Center, lipgloss.Center).
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("69"))
)

type ViewMode string

const (
	COLUMNS ViewMode = "COLUMNS"
	OPEN    ViewMode = "OPEN"
)

type OpenDatabase struct {
	tables     []textinput.Model
	connParams *Connection
	viewMode   ViewMode
}

func NewOpenDatabase(connParams *Connection) OpenDatabase {
	databaseTables := connParams.GetTableNames()

	openDatabase := OpenDatabase{
		tables:     make([]textinput.Model, len(databaseTables)),
		connParams: connParams,
	}

	var t textinput.Model
	for i, value := range databaseTables {
		t = textinput.New()
		t.Cursor.Style = cursorStyle
		t.CharLimit = 32
		t.Placeholder = value

		if i == 0 {
			t.Focus()
			t.PromptStyle = focusedItemStyle
			t.TextStyle = focusedItemStyle
		}

		openDatabase.tables[i] = t
	}

	return openDatabase
}

func (db OpenDatabase) tableView() string {
	var tableLabels strings.Builder
	for i := range db.tables {
		tableLabels.WriteString(db.tables[i].View())
		if i < len(db.tables)-1 {
			tableLabels.WriteRune('\n')
		}
	}
	return tableLabels.String()
}

func (db OpenDatabase) Init() tea.Cmd {
	return textinput.Blink
}

func (db OpenDatabase) Update(msg tea.Msg) (OpenDatabase, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			db.connParams = nil
			return db, nil

		case "left", "right":
			if db.viewMode == COLUMNS {
				db.viewMode = OPEN
			} else {
				db.viewMode = COLUMNS
			}

		}
	}

	return db, nil
}

func (db OpenDatabase) View() string {
	s := fmt.Sprintf("%s\n\n", db.connParams.Name)

	tableLabels := db.tableView()

	if db.viewMode == COLUMNS {
		s += lipgloss.JoinHorizontal(lipgloss.Top, modelStyle.Render(tableLabels), focusedModelStyle.Render("Table"))
	} else {
		s += lipgloss.JoinHorizontal(lipgloss.Top, focusedModelStyle.Render(tableLabels), modelStyle.Render("Table"))
	}

	return s
}
