package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/zalando/go-keyring"
	bolt "go.etcd.io/bbolt"
)

const (
	SERVICE = "termtable-app"
	USER    = "termtable-user-anon"

	LOCAL_BUCKET_NAME = "database_connections"
)

func updateLocalDbConn(name string, host string, port string) error {
	homeDir, err := os.UserHomeDir()

	if err != nil {
		return errors.New("Could not get home directory")
	}

	localDb := fmt.Sprintf("%s/.termtable/connections.db", homeDir)

	db, err := bolt.Open(localDb, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(LOCAL_BUCKET_NAME))
		err := b.Put([]byte(host), []byte(fmt.Sprintf("%s:%s", host, port)))
		return err
	})
}

func getLocalDbConn(name string) (string, error) {
	homeDir, err := os.UserHomeDir()

	if err != nil {
		return "", errors.New("Could not get home directory")
	}

	localDb := fmt.Sprintf("%s/.termtable/connections.db", homeDir)

	db, err := bolt.Open(localDb, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

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
	homeDir, err := os.UserHomeDir()

	if err != nil {
		return errors.New("Could not get home directory")
	}

	localDb := fmt.Sprintf("%s/.termtable/connections.db", homeDir)

	db, err := bolt.Open(localDb, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(LOCAL_BUCKET_NAME))
		err := b.Delete([]byte(name))
		return err
	})

	return nil
}

func listLocalDbConn() (map[string]string, error) {
	homeDir, err := os.UserHomeDir()

	if err != nil {
		return make(map[string]string), errors.New("Could not get home directory")
	}

	localDb := fmt.Sprintf("%s/.termtable/connections.db", homeDir)

	db, err := bolt.Open(localDb, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	connections := make(map[string]string)

	err = db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("MyBucket"))

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
