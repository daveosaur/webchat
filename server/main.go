/*
TODO: implement cookies on user connect, verify with db?
*/
package main

import (
	"context"
	// "database/sql"
	// "errors"

	"fmt"
	"net/http"
	"slices"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
	ws "nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

// globals
var (
	messages    = make([]*Message, 0, 100)
	messageChan = make(chan *Message, 100)
	users       = make([]*User, 0, 10)
	dbChan      = make(chan interface{}, 100)
	mut         sync.Mutex
)

type msgType int

const (
	CONNECT msgType = iota
	MESSAGE
	RENAME
)

type User struct {
	Name string
	UUID string
	Conn *ws.Conn
}

type Message struct {
	Kind      msgType `json:"kind"`
	Guy       string  `json:"guy"`
	Msg       string  `json:"msg"`
	conn      *ws.Conn
	timestamp time.Time
}

func main() {

	mux := http.NewServeMux()
	mux.HandleFunc("/", userHandler)
	err := prepareDatabase()
	defer db.Close()
	if err != nil {
		panic(err)
	}

	// run routines for message passer and serving the client
	go messagePasser()
	go serveClient("8096")

	fmt.Println("listening on 3000")
	err = http.ListenAndServe("0.0.0.0:3000", mux)
	if err != nil {
		panic(err)
	}

}

func messagePasser() {
	ctx := context.Background()

	for {
		msg := <-messageChan

		mut.Lock()
		if msg.Kind == MESSAGE {
			messages = append(messages, msg)
		}

		sendMessage(msg, ctx, nil)

		mut.Unlock()
	}
}

func sendMessage(msg *Message, ctx context.Context, target *ws.Conn) {
	switch msg.Kind {
	case CONNECT:
		switch target {
		case nil:
			users = append(users, &User{Name: msg.Guy, Conn: msg.conn})
			for _, user := range users {
				user.Conn.Write(ctx, ws.MessageText, []byte(fmt.Sprintf("%s connected!", msg.Guy)))
			}
			// default:
			// 	target.Write(ctx, ws.MessageText, []byte(fmt.Sprintf("%s connected!", msg.Guy)))
		}
	case MESSAGE:
		switch target {
		case nil:
			for _, user := range users {
				user.Conn.Write(ctx, ws.MessageText, []byte(fmt.Sprintf("<%s> %s", msg.Guy, msg.Msg)))
			}
		default:
			target.Write(ctx, ws.MessageText, []byte(fmt.Sprintf("<%s> %s", msg.Guy, msg.Msg)))
		}
	case RENAME:
		switch target {
		case nil:
			var oldname string
			for _, user := range users {
				if user.Conn == msg.conn {
					oldname = user.Name
					user.Name = msg.Guy
				}
			}
			for _, user := range users {
				user.Conn.Write(ctx, ws.MessageText, []byte(fmt.Sprintf("%s is now %s", oldname, msg.Guy)))

			}
		}
	}
}

func userHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("connected!")

	// TODO:do db stuff
	// TODO: add user to db if they dont exist.

	//lol cors
	opts := ws.AcceptOptions{InsecureSkipVerify: true}
	c, err := ws.Accept(w, r, &opts)
	if err != nil {
		fmt.Println(err)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), time.Hour)
	defer cancel()

	//recieve username
	var u *Message
	wsjson.Read(ctx, c, &u)
	u.conn = c

	//send new user all messages?
	go func(con *ws.Conn) {
		for _, msg := range messages {
			//sleep to not send packets too fast?
			time.Sleep(50 * time.Millisecond)
			sendMessage(msg, ctx, con)
		}
	}(c)

	messageChan <- u
	dbChan <- u

	for {
		var msg Message
		err := wsjson.Read(ctx, c, &msg)
		// conn closed
		if err != nil {
			fmt.Println(err)
			mut.Lock()
			defer mut.Unlock()
			for i, user := range users {
				if user.Conn == c {
					users = slices.Delete(users, i, i+1)
				}
			}
			return
		}
		msg.conn = c
		messageChan <- &msg

		fmt.Printf("<%s> %s\n", msg.Guy, msg.Msg)
	}
}

func serveClient(port string) {
	fs := http.FileServer(http.Dir("./dist"))

	mux := http.NewServeMux()
	mux.Handle("/", fs)

	fmt.Println("serving client 8096")
	err := http.ListenAndServe(":"+port, mux)
	if err != nil {
		panic(err)
	}
}
