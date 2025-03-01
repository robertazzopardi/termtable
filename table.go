package main

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
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

type DbTable struct {
	*tview.Table
	table Table
}

func newDbTable(table Table) *DbTable {
	t := tview.NewTable()

	for i, header := range table.fields {
		t.SetCell(0, i, tview.NewTableCell(header).SetExpansion(1))
	}

	t.SetBorderPadding(0, 0, 1, 1)
	t.SetSelectable(true, false).Select(1, 0)

	connectionsTable := DbTable{t, table}
	connectionsTable.getTableRows()

	return &connectionsTable
}

func (t *DbTable) getTableRows() {
	for i, conn := range t.table.values {
		for j, value := range conn {
			t.SetCell(i+1, j, tview.NewTableCell(value))
		}
	}
}

type ContentBox struct {
	*tview.Box
	content tview.Primitive
}

func newContentBox(title string, content tview.Primitive) *ContentBox {
	return &ContentBox{
		tview.NewBox().SetBorder(true).SetTitle(fmt.Sprintf(" %s ", title)),
		content,
	}
}

func (b *ContentBox) Draw(screen tcell.Screen) {
	b.Box.DrawForSubclass(screen, b)
	x, y, w, h := b.GetInnerRect()

	b.content.SetRect(x, y, w, h)
	b.content.Draw(screen)
}

func (b *ContentBox) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return b.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		b.content.InputHandler()(event, setFocus)
	})
}
