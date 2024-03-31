package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type Connection struct {
	Host string
	Port string
	User string
	Pass string
	Name string
}

func (params Connection) ConnectionString() string {
	// urlExample := "postgres://username:password@localhost:5432/database_name"
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s", params.User, params.Pass, params.Host, params.Port, params.Name)
}

func (params Connection) TestConnection() bool {
	connectionString := params.ConnectionString()
	conn, err := pgx.Connect(context.Background(), connectionString)

	if err != nil {
		return false
	}

	conn.Close(context.Background())

	return true
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
