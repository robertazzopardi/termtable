package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/zalando/go-keyring"
	_ "modernc.org/sqlite"
)

const (
	SERVICE = "termtable-app"
)

func getAndOrCreateLocalDb() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", errors.New("could not get home directory")
	}

	dbDir := filepath.Join(homeDir, ".termtable")
	localDb := filepath.Join(dbDir, "termtable.db")

	if _, err := os.Stat(dbDir); os.IsNotExist(err) {
		err := os.MkdirAll(dbDir, 0o755)
		if err != nil {
			log.Fatal("Could not create directory to store local db: ", err)
		}
	}

	// Create the database file if it doesn't exist
	if _, err := os.Stat(localDb); os.IsNotExist(err) {
		file, err := os.Create(localDb)
		if err != nil {
			return "", fmt.Errorf("could not create database file: %v", err)
		}
		file.Close()
	}

	return localDb, nil
}

func initDb(db *sql.DB) error {
	// Create the table if it doesn't exist
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS database_connections (
			name TEXT PRIMARY KEY,
			host TEXT NOT NULL,
			port TEXT NOT NULL,
			database_name TEXT NOT NULL
		)
	`)
	return err
}

func updateLocalDbConn(conn Connection) error {
	localDb, err := getAndOrCreateLocalDb()
	if err != nil {
		return err
	}

	db, err := sql.Open("sqlite", localDb)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = initDb(db)
	if err != nil {
		return err
	}

	// Insert or replace the connection
	_, err = db.Exec(
		"INSERT OR REPLACE INTO database_connections (name, host, port, database_name) VALUES (?, ?, ?, ?)",
		conn.Name, conn.Host, conn.Port, conn.Database,
	)

	return err
}

func deleteLocalDbConn(name string) error {
	localDb, err := getAndOrCreateLocalDb()
	if err != nil {
		return err
	}

	db, err := sql.Open("sqlite", localDb)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = initDb(db)
	if err != nil {
		return err
	}

	_, err = db.Exec("DELETE FROM database_connections WHERE name = ?", name)
	return err
}

func listLocalDbConn() (map[string]string, error) {
	connections := make(map[string]string)
	localDb, err := getAndOrCreateLocalDb()

	if err != nil {
		return connections, err
	}

	db, err := sql.Open("sqlite", localDb)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = initDb(db)
	if err != nil {
		return connections, err
	}

	rows, err := db.Query("SELECT name, host, port, database_name FROM database_connections")
	if err != nil {
		return connections, err
	}
	defer rows.Close()

	for rows.Next() {
		var name, host, port, database string
		if err := rows.Scan(&name, &host, &port, &database); err != nil {
			log.Fatal(err)
		}
		connections[name] = fmt.Sprintf("%s:%s:%s", host, port, database)
	}

	return connections, rows.Err()
}

func createKeyringPassword(username string, password string) string {
	return fmt.Sprintf("%s:%s", username, password)
}

func parseKeyringPassword(password string) (string, string, error) {
	passwordComponents := strings.Split(password, ":")

	if len(passwordComponents) != 2 {
		log.Fatal("Expected saved password to contain 2 components")
	}

	return passwordComponents[0], passwordComponents[1], nil
}

func SaveConnectionInKeyring(conn Connection) {
	// Save keyring part
	password := createKeyringPassword(conn.User, conn.Password)
	err := keyring.Set(SERVICE, conn.Name, password)
	if err != nil {
		log.Fatal("Could not save db credentials in keyring: ", err)
	}

	// Save rest to local storage
	err = updateLocalDbConn(conn)
	if err != nil {
		log.Fatal("Could not set keyring info into local db: ", err)
	}
}

func GetConnectionFromKeyring(name string) (string, string, error) {
	password, err := keyring.Get(SERVICE, name)
	if err != nil {
		log.Fatal("Could not get credentials for connection: ", err)
	}

	return parseKeyringPassword(password)
}

func ListConnections() ([]Connection, error) {
	connections, err := listLocalDbConn()

	var conns []Connection

	if err != nil {
		return conns, errors.New("Could not list connections")
	}

	for k, v := range connections {
		hostPortDb := strings.Split(v, ":")
		if len(hostPortDb) != 3 {
			continue
		}

		user, password, err := GetConnectionFromKeyring(k)
		if err != nil {
			log.Fatal("Could not get user and password for db", err)
		}

		conn := Connection{
			Name:     k,
			User:     user,
			Password: password,
			Host:     hostPortDb[0],
			Port:     hostPortDb[1],
			Database: hostPortDb[2],
		}
		conns = append(conns, conn)
	}

	return conns, nil
}
