package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

// prepared database statements
var (
	insertUser    *sql.Stmt
	getUser       *sql.Stmt
	insertMessage *sql.Stmt
	db            *sql.DB
)

func prepareDatabase() error {
	var err error
	db, err = sql.Open("sqlite3", "webchat.db")
	if err != nil {
		return err
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS users(id INTEGER PRIMARY KEY AUTOINCREMENT, name varchar(16) UNIQUE NOT NULL, uuid VARCHAR(64) NOT NULL)")
	if err != nil {
		return err
	}
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS messages(id INTEGER PRIMARY KEY AUTOINCREMENT, user VARCHAR(16) NOT NULL, msg VARCHAR(32768) NOT NULL)")
	if err != nil {
		return err
	}

	//prepare statements
	insertUser, _ = db.Prepare("INSERT INTO users (name, uuid) VALUES (?, ?)")
	getUser, _ = db.Prepare("SELECT name FROM users WHERE uuid = ?")
	insertMessage, _ = db.Prepare("INSERT INTO messages (user, msg) VALUES (?, ?)")

	return nil
}

func dbHandler() {
	for {
		data := <-dbChan

		switch data.(type) {
		case User:
		case Message:

		}

	}
}
