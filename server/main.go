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
	messages = make(chan *Message, 100)
	users    = make([]*User, 0)
	mut      sync.Mutex
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
		msg := <-messages
		mut.Lock()
		switch msg.Kind {
		case CONNECT:
			users = append(users, &User{Name: msg.Guy, conn: msg.conn})
			for _, user := range users {
				user.conn.Write(ctx, ws.MessageText, []byte(fmt.Sprintf("%s connected!", msg.Guy)))
			}

		case MESSAGE:
			//get name
			var name string
			for _, user := range users {
				if user.conn == msg.conn {
					name = user.Name
				}
			}
			for _, user := range users {
				user.conn.Write(ctx, ws.MessageText, []byte(fmt.Sprintf("<%s> %s", name, msg.Msg)))

			}

		case RENAME:
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
		mut.Unlock()

	}

}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("connected!")
		opts := ws.AcceptOptions{InsecureSkipVerify: true}
		c, err := ws.Accept(w, r, &opts)
		if err != nil {
			fmt.Println(err)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), time.Hour)
		defer cancel()

		//recieve username
		// _, name, err := c.Read(ctx)
		var u *Message
		wsjson.Read(ctx, c, &u)
		u.conn = c

		messages <- u

		// var v interface{}
		// err = wskson.Read(ctx, c, &v)
		// if err != nil {
		// 	fmt.Println(err)
		// 	return
		// }
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
			// if err != nil {
			// 	panic(err)
			// }
			// fmt.Println(v)
			msg.conn = c

			messages <- &msg

			fmt.Printf("got: <%s> %s\n", msg.Guy, msg.Msg)
			// c.Write(ctx, ws.MessageText, msg)
			// fmt.Printf("got %v\n", v)
		}

	})

	go messagePasser()

	fmt.Println("starting server...")
	http.ListenAndServe("0.0.0.0:3000", nil)

}
