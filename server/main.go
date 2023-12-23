/*
TODO: implement cookies on user connect, verify with db?
*/
package main

import (
	"context"
	"crypto/rand"

	// "database/sql"
	// "errors"

	"fmt"
	"net/http"
	"slices"
	"sync"
	"time"

	"github.com/gorilla/sessions"
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
	store       sessions.Store
)

type msgType int

const (
	CONNECT msgType = iota
	MESSAGE
	RENAME
	SERVERMSG
)

type User struct {
	Name string
	UUID string
	conn *ws.Conn
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
	if err != nil {
		panic(err)
	}
	defer db.Close()
	store = sessions.NewCookieStore([]byte("test"))

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
			users = append(users, &User{Name: msg.Guy, conn: msg.conn})
			wsjson.Write(ctx, msg.conn, msg)
			messageChan <- &Message{Kind: SERVERMSG, Msg: fmt.Sprintf("%s joined!", msg.Guy)}
			// for _, user := range users {
			// 	// user.conn.Write(ctx, ws.MessageText, []byte(fmt.Sprintf("%s connected!", msg.Guy)))
			// }
			// default:
			// 	target.Write(ctx, ws.MessageText, []byte(fmt.Sprintf("%s connected!", msg.Guy)))
		}
	case MESSAGE:
		switch target {
		case nil:
			for _, user := range users {
				// user.Conn.Write(ctx, ws.MessageText, []byte(fmt.Sprintf("<%s> %s", msg.Guy, msg.Msg)))
				wsjson.Write(ctx, user.conn, msg)
			}
		default:
			// target.Write(ctx, ws.MessageText, []byte(fmt.Sprintf("<%s> %s", msg.Guy, msg.Msg)))
			wsjson.Write(ctx, msg.conn, msg)
		}
	case RENAME:
		switch target {
		case nil:
			var oldname string
			for _, user := range users {
				if user.conn == msg.conn {
					oldname = user.Name
					user.Name = msg.Guy
				}
			}
			changeName.Exec(msg.Guy, oldname)
			for _, user := range users {
				// user.conn.Write(ctx, ws.MessageText, []byte(fmt.Sprintf("%s is now %s", oldname, msg.Guy)))
				wsjson.Write(ctx, user.conn, Message{Kind: RENAME, Msg: fmt.Sprintf("%s is now %s", oldname, msg.Guy)})

			}
		}
	case SERVERMSG:
		for _, user := range users {
			fmt.Println(user.Name)
			wsjson.Write(ctx, user.conn, msg)
		}
	}
}

func userHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("connected!")
	// set/get session
	session, err := store.Get(r, "daveChat")
	if err != nil {
		session, _ = store.New(r, "daveChat")
	}
	// set session token
	if session.Values["token"] == nil {
		sesToken := make([]byte, 16)
		rand.Read(sesToken)
		session.Values["token"] = string(sesToken)
		session.Save(r, w)
	}

	// db query stuff
	rows, err := getSession.Query(session.Values["token"])
	if err != nil {
		fmt.Println(err)
	}
	var userName string
	var expiration int
	var exists bool
	for rows.Next() {
		exists = true
		rows.Scan(&userName, &expiration)
	}
	fmt.Println(userName, expiration)

	// accept websocket upgrade
	opts := ws.AcceptOptions{InsecureSkipVerify: true}
	c, err := ws.Accept(w, r, &opts)
	if err != nil {
		fmt.Println(err)
		return
	}

	ctx := r.Context()

	u := Message{conn: c, Kind: CONNECT, Guy: userName}
	// if no database entry for user, accept name from user
	if !exists {
		wsjson.Read(ctx, c, &u)
		expiration = int(time.Now().Add(24 * time.Hour).Unix())
		_, err := newSession.Exec(session.Values["token"], u.Guy, expiration)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
	messageChan <- &u

	//send new user all messages?
	// go func(con *ws.Conn) {
	// 	for _, msg := range messages {
	// 		//sleep to not send packets too fast?
	// 		time.Sleep(50 * time.Millisecond)
	// 		sendMessage(msg, ctx, con)
	// 	}
	// }(c)

	// TODO: add message to database at some point
	// dbChan <- u

	// loop to deal with messages
	for {
		var msg Message
		err := wsjson.Read(ctx, c, &msg)
		// conn closed
		if err != nil {
			fmt.Println(err)
			mut.Lock()
			defer mut.Unlock()
			for i, user := range users {
				if user.conn == c {
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
