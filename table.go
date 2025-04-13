package main

import (
	"strings"

	"github.com/rivo/tview"
)

type DisplayTable struct {
	*tview.Table
	columns []string
	rows    []Connection
}

func newConnectionsTable(columns []string) *DisplayTable {
	table := tview.NewTable()

	for i, header := range columns {
		table.SetCell(0, i, tview.NewTableCell(header).SetExpansion(1))
	}

	table.SetBorderPadding(0, 0, 1, 1)
	table.SetSelectable(true, false).Select(1, 0)

	connectionsTable := DisplayTable{table, columns, []Connection{}}
	connectionsTable.getConnections()

	return &connectionsTable
}

func (t *DisplayTable) getConnections() {
	connections, err := ListConnections()
	if err != nil {
		return
	}

	for i, conn := range connections {
		values := conn.Row()
		for j, value := range values {
			t.SetCell(i+1, j, tview.NewTableCell(value))
		}
	}

	t.rows = connections
}

func (t *DisplayTable) getConnection() *Connection {
	row, _ := t.GetSelection()

	if row == 0 {
		return nil
	}

	return &t.rows[row-1]
}

func (t *DisplayTable) updateTable(filter string) {
	t.Clear()

	for i, header := range t.columns {
		t.SetCell(0, i, tview.NewTableCell(header).SetExpansion(1))
	}

	row := 1
	for _, conn := range t.rows {
		lowerFilter := strings.ToLower(filter)

		if filter == "" ||
			strings.Contains(strings.ToLower(conn.Host), lowerFilter) ||
			strings.Contains(strings.ToLower(conn.User), lowerFilter) ||
			strings.Contains(strings.ToLower(conn.Database), lowerFilter) ||
			strings.Contains(strings.ToLower(conn.Name), lowerFilter) {

			values := conn.Row()
			for j, value := range values {
				t.SetCell(row, j, tview.NewTableCell(value))
			}

			row++
		}
	}
}

type DbTable struct {
	*tview.Table
	db OpenDatabase
}

func newDbTable(db OpenDatabase) *DbTable {
	t := tview.NewTable()
	t.SetBorderPadding(0, 0, 1, 1)
	t.SetSelectable(true, false).Select(1, 0)

	connectionsTable := DbTable{t, db}
	connectionsTable.setTableRows()

	return &connectionsTable
}

func (t *DbTable) setTableRows() {
	table := t.db.openTable

	for i, header := range table.fields {
		t.SetCell(0, i, tview.NewTableCell(header).SetExpansion(1))
	}

	for i, value := range t.db.openTable.values {
		for j, value := range value {
			t.SetCell(i+1, j, tview.NewTableCell(value))
		}
	}
}

func (t *DbTable) showSchema() {
	row, col := t.GetSelection()

	cell := t.GetCell(row, col)
	t.db.setSchema(cell.Text)

	t.db.getTablesInSchema()

	t.Clear().ScrollToBeginning()

	t.setTableRows()
}

func (t *DbTable) showTables() {
	row, _ := t.GetSelection()

	t.db.setOpenTable(row)

	t.Clear().ScrollToBeginning()

	t.setTableRows()
}
