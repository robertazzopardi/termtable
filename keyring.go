package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/zalando/go-keyring"
	bolt "go.etcd.io/bbolt"
)

const (
	SERVICE = "termtable-app"

	LOCAL_BUCKET_NAME = "database_connections"
)

func getAndOrCreateLocalDb() (string, error) {
	homeDir, err := os.UserHomeDir()

	if err != nil {
		return "", errors.New("Could not get home directory")
	}

	localDb := fmt.Sprintf("%s/.termtable/connections.db", homeDir)

	if _, err := os.Stat(localDb); os.IsNotExist(err) {
		err := os.Mkdir(filepath.Dir(localDb), 0755)
		if err != nil {
			log.Fatal("Could not create directory to store local db: ", err)
		}
	}

	return localDb, nil
}

func createBucket(db *bolt.DB) error {
	// Start a writable transaction.
	tx, err := db.Begin(true)
	if err != nil {
		return err
	}
	defer func() {
		err = tx.Rollback()
	}()

	// Use the transaction...
	_, err = tx.CreateBucket([]byte(LOCAL_BUCKET_NAME))
	if err != nil {
		return err
	}

	// Commit the transaction and check for error.
	if err := tx.Commit(); err != nil {
		return err
	}

	return err
}

func updateLocalDbConn(conn Connection) error {
	localDb, err := getAndOrCreateLocalDb()

	if err != nil {
		return err
	}

	db, err := bolt.Open(localDb, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = createBucket(db)
	if err != nil {
		return err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(LOCAL_BUCKET_NAME))
		err := b.Put([]byte(conn.Name), []byte(fmt.Sprintf("%s:%s:%s", conn.Host, conn.Port, conn.Database)))
		return err
	})

	return err
}

func deleteLocalDbConn(name string) error {
	localDb, err := getAndOrCreateLocalDb()

	if err != nil {
		return err
	}

	db, err := bolt.Open(localDb, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = createBucket(db)
	if err != nil {
		return err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(LOCAL_BUCKET_NAME))
		err := b.Delete([]byte(name))
		return err
	})

	return err
}

func listLocalDbConn() (map[string]string, error) {
	localDb, err := getAndOrCreateLocalDb()

	connections := make(map[string]string)

	if err != nil {
		return connections, err
	}

	db, err := bolt.Open(localDb, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = createBucket(db)
	if err != nil {
		return connections, err
	}

	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(LOCAL_BUCKET_NAME))

		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			key := string(k)
			value := string(v)
			connections[key] = value
		}

		return nil
	})

	return connections, err
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
	password := createKeyringPassword(conn.User, conn.Pass)
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
		conn := Connection{
			Name:     k,
			Host:     hostPortDb[0],
			Port:     hostPortDb[1],
			Database: hostPortDb[2],
		}
		conns = append(conns, conn)
	}

	return conns, nil
}
