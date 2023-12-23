package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

// prepared database statements
var (
	insertUser    *sql.Stmt
	changeName    *sql.Stmt
	getSession    *sql.Stmt
	insertMessage *sql.Stmt
	newSession    *sql.Stmt
	db            *sql.DB
)

func prepareDatabase() error {
	var err error
	db, err = sql.Open("sqlite3", "webchat.db")
	if err != nil {
		return err
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS users(id INTEGER PRIMARY KEY AUTOINCREMENT, name VARCHAR(16) UNIQUE NOT NULL)")
	if err != nil {
		return err
	}
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS messages(id INTEGER PRIMARY KEY AUTOINCREMENT, user VARCHAR(16) NOT NULL, msg VARCHAR(32768) NOT NULL)")
	if err != nil {
		return err
	}
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS sessions(id INTEGER PRIMARY KEY AUTOINCREMENT, token VARCHAR(16) NOT NULL, user VARCHAR(16), expiration INTEGER NOT NULL)")
	if err != nil {
		return err
	}

	//prepare statements
	insertUser, err = db.Prepare("INSERT INTO users (name) VALUES (?)")
	if err != nil {
		return err
	}
	newSession, err = db.Prepare("INSERT INTO sessions (token, user, expiration) VALUES (?, ?, ?)")
	if err != nil {
		return err
	}
	changeName, err = db.Prepare("UPDATE sessions SET user = ? WHERE user = ?")
	if err != nil {
		return err
	}
	getSession, err = db.Prepare("SELECT user, expiration FROM sessions WHERE token = ?")
	if err != nil {
		return err
	}
	insertMessage, err = db.Prepare("INSERT INTO messages (user, msg) VALUES (?, ?)")
	if err != nil {
		return err
	}

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
