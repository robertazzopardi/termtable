package main

import "log"

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
