package main

import (
	"log"
)

type ViewMode string

type OpenDatabase struct {
	params    Connection
	openTable Table
	schema    string
}

func NewOpenDatabase(connParams Connection) OpenDatabase {
	openDatabase := OpenDatabase{
		params: connParams,
	}

	openDatabase.setSchemas()

	return openDatabase
}

func (db *OpenDatabase) setSchemas() {
	schemas := db.params.GetSchemas()

	db.setTable(schemas, "schema", "schema_names")
}

func (db *OpenDatabase) getTablesInSchema() {
	tables := db.params.GetTableNames(db.schema)

	db.setTable(tables, "table", "table_names")
}

func (db *OpenDatabase) setTable(tables []string, name, title string) {
	if len(tables) == 0 {
		return
	}

	rows := make([][]string, len(tables))

	for i, table := range tables {
		rows[i] = []string{table}
	}

	db.openTable = Table{name: name, fields: []string{title}, values: rows}
}

func (db *OpenDatabase) setOpenTable(index int) {
	tables := db.params.GetTableNames(db.schema)

	if len(tables) <= index {
		return
	}

	tableName := tables[index]

	table, err := db.params.SelectAll(tableName)
	if err != nil {
		log.Fatal("Could not connect to db", err)

		return
	}

	db.openTable = table
}

func (db *OpenDatabase) setSchema(schema string) {
	db.schema = schema
}
