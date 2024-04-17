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
	USER    = "termtable-user-anon"

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
			os.Exit(0)
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
	defer tx.Rollback()

	// Use the transaction...
	_, err = tx.CreateBucket([]byte(LOCAL_BUCKET_NAME))
	if err != nil {
		return err
	}

	// Commit the transaction and check for error.
	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func updateLocalDbConn(name string, host string, port string) error {
	localDb, err := getAndOrCreateLocalDb()

	if err != nil {
		return err
	}

	db, err := bolt.Open(localDb, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	createBucket(db)

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(LOCAL_BUCKET_NAME))
		err := b.Put([]byte(name), []byte(fmt.Sprintf("%s:%s", host, port)))
		return err
	})

	return err
}

func getLocalDbConn(name string) (string, error) {
	localDb, err := getAndOrCreateLocalDb()

	if err != nil {
		return "", err
	}

	db, err := bolt.Open(localDb, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	createBucket(db)

	var value string
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(LOCAL_BUCKET_NAME))
		v := b.Get([]byte(name))
		value = fmt.Sprintf("%s", v)
		return nil
	})

	return value, err
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
	createBucket(db)

	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(LOCAL_BUCKET_NAME))
		err := b.Delete([]byte(name))
		return err
	})

	return nil
}

func listLocalDbConn() (map[string]string, error) {
	localDb, err := getAndOrCreateLocalDb()

	if err != nil {
		return make(map[string]string), err
	}

	db, err := bolt.Open(localDb, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	createBucket(db)

	connections := make(map[string]string)

	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(LOCAL_BUCKET_NAME))

		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			fmt.Printf("key=%s, value=%s\n", k, v)
			key := fmt.Sprintf("%s", k)
			value := fmt.Sprintf("%s", v)
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
		return "", "", errors.New("Expected saved password to contain 2 components")
	}

	return passwordComponents[0], passwordComponents[1], nil
}

func SaveConnectionInKeyring(conn Connection) error {
	password := createKeyringPassword(conn.User, conn.Pass)
	err := keyring.Set(SERVICE, USER, password)

	if err != nil {
		return err
	}

	return nil
}

func ListConnections() ([]Connection, error) {
	connections, err := listLocalDbConn()

	var conns []Connection

	if err != nil {
		return conns, errors.New("Could not list connections")
	}

	for k, v := range connections {
		hostPort := strings.Split(v, ":")
		if len(hostPort) != 2 {
			log.Fatal(fmt.Sprintf("Connection %s does not have the correct host and port value", k))
			continue
		}
		conn := Connection{
			Name: k,
			Host: hostPort[0],
			Port: hostPort[1],
		}
		conns = append(conns, conn)
	}

	return conns, nil
}
