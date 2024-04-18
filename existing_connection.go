package main

import (
	"log"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type ExistingConnectionsModel struct {
	list               list.Model
	connections        []Connection
	selectedConnection *Connection
	back               bool
}

func NewExistingConnectionsModel() ExistingConnectionsModel {
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
			m.back = true
			return m, nil

		case "enter":
			i, ok := m.list.SelectedItem().(item)
			if ok {
				choice := string(i)

				for _, v := range m.connections {
					if v.Name == choice {
						m.selectedConnection = &v

						user, pass, err := GetConnectionFromKeyring(v.Name)

						if err != nil {
							log.Fatal("Could not get user and password for connection from keyring: ", err)
						}

						m.selectedConnection.User = user
						m.selectedConnection.Pass = pass

						break
					}
				}
			}
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m ExistingConnectionsModel) View() string {
	return m.list.View()
}
