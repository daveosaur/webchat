package main

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"sync"
	"time"

	ws "nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

var (
	messages    = make([]*Message, 0, 100)
	messageChan = make(chan *Message, 100)
	users       = make([]*User, 0, 10)
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
	conn *ws.Conn
}

type Message struct {
	Kind msgType `json:"kind"`
	Guy  string  `json:"guy"`
	Msg  string  `json:"msg"`
	conn *ws.Conn
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

func main() {
	http.HandleFunc("/", userHandler)

	go messagePasser()

	fmt.Println("listening on 3000")
	http.ListenAndServe("0.0.0.0:3000", nil)

}

func sendMessage(msg *Message, ctx context.Context, target *ws.Conn) {
	switch msg.Kind {
	case CONNECT:
		switch target {
		case nil:
			users = append(users, &User{Name: msg.Guy, conn: msg.conn})
			for _, user := range users {
				user.conn.Write(ctx, ws.MessageText, []byte(fmt.Sprintf("%s connected!", msg.Guy)))
			}
			// default:
			// 	target.Write(ctx, ws.MessageText, []byte(fmt.Sprintf("%s connected!", msg.Guy)))
		}
	case MESSAGE:
		switch target {
		case nil:
			for _, user := range users {
				user.conn.Write(ctx, ws.MessageText, []byte(fmt.Sprintf("<%s> %s", msg.Guy, msg.Msg)))
			}
		default:
			target.Write(ctx, ws.MessageText, []byte(fmt.Sprintf("<%s> %s", msg.Guy, msg.Msg)))
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
			for _, user := range users {
				user.conn.Write(ctx, ws.MessageText, []byte(fmt.Sprintf("%s is now %s", oldname, msg.Guy)))

			}
		}
	}
}

func userHandler(w http.ResponseWriter, r *http.Request) {

	fmt.Println("connected!")
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
			time.Sleep(50 * time.Millisecond)
			sendMessage(msg, ctx, con)
		}
	}(c)

	messageChan <- u

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

		fmt.Printf("got: <%s> %s\n", msg.Guy, msg.Msg)
	}

}
