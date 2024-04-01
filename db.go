package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type ConnectionStatus string

const (
	CONNECTED    ConnectionStatus = "CONNECTED"
	DISCONNECTED ConnectionStatus = "DISCONNECTED"
)

type Connection struct {
	Host   string
	Port   string
	User   string
	Pass   string
	Name   string
	status ConnectionStatus
}

func (params Connection) ConnectionString() string {
	// urlExample := "postgres://username:password@localhost:5432/database_name"
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s", params.User, params.Pass, params.Host, params.Port, params.Name)
}

func (params *Connection) TestConnection() TestStatus {
	connectionString := params.ConnectionString()
	conn, err := pgx.Connect(context.Background(), connectionString)

	if err != nil {
		params.status = DISCONNECTED
		return FAILED
	}

	conn.Close(context.Background())

	params.status = CONNECTED
	return PASSED
}

func (parmas Connection) GetTableNames() []string {
	connectionString := parmas.ConnectionString()
	conn, err := pgx.Connect(context.Background(), connectionString)

	if err != nil {
		return nil
	}

	rows, err := conn.Query(context.Background(), "SELECT table_name FROM information_schema.tables WHERE table_schema='public'")

	if err != nil {
		return nil
	}

	var tableNames []string
	for rows.Next() {
		var tableName string
		err = rows.Scan(&tableName)
		if err != nil {
			return nil
		}
		tableNames = append(tableNames, tableName)
	}

	conn.Close(context.Background())

	return tableNames
}

type Table struct {
	fields []string
	values [][]string
}

func (params Connection) SelectAll(table string) (Table, error) {
	connectionString := params.ConnectionString()
	conn, err := pgx.Connect(context.Background(), connectionString)

	if err != nil {
		return Table{}, err
	}

	rows, err := conn.Query(context.Background(), fmt.Sprintf("SELECT * FROM %s", table))

	if err != nil {
		return Table{}, err
	}

	var tableData Table

	fieldDescriptions := rows.FieldDescriptions()
	tableData.fields = make([]string, len(fieldDescriptions))
	for i, field := range fieldDescriptions {
		tableData.fields[i] = field.Name
	}

	for rows.Next() {
		var values []string
		err = rows.Scan(values)
		if err != nil {
			return Table{}, err
		}
		tableData.values = append(tableData.values, values)
	}

	conn.Close(context.Background())

	return tableData, nil
}
