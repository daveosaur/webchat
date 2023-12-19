package main

import (
	"fmt"
	"net/http"
)

func main() {
	fs := http.FileServer(http.Dir("./files"))

	http.Handle("/", fs)

	fmt.Println("listening on 8096")
	err := http.ListenAndServe(":8096", nil)
	if err != nil {
		panic(err)
	}
}
