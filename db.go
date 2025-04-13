package main

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type ConnectionStatus string

const (
	CONNECTED    ConnectionStatus = "CONNECTED"
	DISCONNECTED ConnectionStatus = "DISCONNECTED"
)

type Connection struct {
	ID       uuid.UUID
	Host     string
	Port     string
	User     string
	Password string
	Database string
	Name     string
	status   ConnectionStatus
}

func NewConnection() Connection {
	ID, err := uuid.NewV7()
	if err != nil {
		log.Fatal("Could not create id for connection", err)
	}
	return Connection{ID: ID}
}

func (c Connection) Row() []string {
	return []string{c.ID.String(), c.Name, c.Host, c.Port, c.User, c.Database}
}

func (params Connection) ConnectionString() string {
	// urlExample := "postgres://username:password@localhost:5432/database_name"
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s", params.User, params.Password, params.Host, params.Port, params.Database)
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

func (params Connection) GetSchemas() []string {
	connectionString := params.ConnectionString()
	conn, err := pgx.Connect(context.Background(), connectionString)
	if err != nil {
		return nil
	}

	rows, err := conn.Query(context.Background(), "SELECT schema_name FROM information_schema.schemata")
	if err != nil {
		return nil
	}

	var schemaNames []string

	for rows.Next() {
		var schemaName string

		err = rows.Scan(&schemaName)
		if err != nil {
			return nil
		}

		schemaNames = append(schemaNames, schemaName)
	}

	conn.Close(context.Background())

	return schemaNames
}

func (parmas Connection) GetTableNames(schema string) []string {
	connectionString := parmas.ConnectionString()
	conn, err := pgx.Connect(context.Background(), connectionString)
	if err != nil {
		return nil
	}

	query := fmt.Sprintf("SELECT table_name FROM information_schema.tables WHERE table_schema = '%s'", schema)
	rows, err := conn.Query(context.Background(), query)
	if err != nil {
		return []string{"No tables in schema"}
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
	name   string
	fields []string
	values [][]string
}

func (params Connection) SelectAll(table string) (Table, error) {
	connectionString := params.ConnectionString()
	conn, err := pgx.Connect(context.Background(), connectionString)
	if err != nil {
		return Table{}, err
	}

	rows, err := conn.Query(context.Background(), "SELECT * FROM "+table)
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
		values, err := rows.Values()
		if err != nil {
			return Table{}, err
		}

		strValues := make([]string, len(values))

		for i, value := range values {
			strValues[i] = fmt.Sprintf("%v", value)
		}

		tableData.values = append(tableData.values, strValues)
	}

	tableData.name = table

	conn.Close(context.Background())

	return tableData, nil
}
