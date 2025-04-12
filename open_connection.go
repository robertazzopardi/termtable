package main

import "log"

type ViewMode string

type OpenDatabase struct {
	tables    []string
	params    Connection
	openTable Table
}

func NewOpenDatabase(connParams Connection) OpenDatabase {
	databaseTables := connParams.GetTableNames()

	openDatabase := OpenDatabase{
		tables: databaseTables,
		params: connParams,
	}

	openDatabase.setOpenTable()

	return openDatabase
}

func (db *OpenDatabase) setOpenTable() {
	if len(db.tables) == 0 {
		return
	}

	tableName := db.tables[0]

	table, err := db.params.SelectAll(tableName)
	if err != nil {
		log.Fatal("Could not connect to db", err)

		return
	}

	db.openTable = table
}
