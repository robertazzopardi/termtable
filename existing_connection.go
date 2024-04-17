package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type ExistingConnectionsModel struct {
	list        list.Model
	connections []Connection
	choice      string
}

func NewExistingConnectionsModel() ExistingConnectionsModel {
	const defaultWidth = 20

	existingConnectionsModel := ExistingConnectionsModel{}

	connections, err := ListConnections()

	if err != nil {
		return existingConnectionsModel
	}

	items := []list.Item{}

	for _, conn := range connections {
		items = append(items, item(conn.Name))
	}

	l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
	l.Title = "Choose a connection"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	existingConnectionsModel.list = l
	existingConnectionsModel.connections = connections

	return existingConnectionsModel
}

func (m ExistingConnectionsModel) Init() tea.Cmd {
	return nil
}

func (m ExistingConnectionsModel) Update(msg tea.Msg) (ExistingConnectionsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "q", "ctrl+c":
			os.Exit(0)

		case "enter":
			i, ok := m.list.SelectedItem().(item)
			if ok {
				m.choice = string(i)
			}
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m ExistingConnectionsModel) View() string {
	if m.choice != "" {
		return quitTextStyle.Render(fmt.Sprintf("%s? Sounds good to me.", m.choice))
	}

	return "\n" + m.list.View()
}
