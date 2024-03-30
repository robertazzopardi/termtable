package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type ConnectionParams struct {
	Host string
	Port string
	User string
	Pass string
	Name string
}

func (params ConnectionParams) ConnectionString() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s", params.User, params.Pass, params.Host, params.Port, params.Name)
}

func (params ConnectionParams) TestConnection() bool {
	// urlExample := "postgres://username:password@localhost:5432/database_name"
	connectionString := params.ConnectionString()
	conn, err := pgx.Connect(context.Background(), connectionString)

	if err != nil {
		// fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		return false
	}

	conn.Close(context.Background())

	return true
}
