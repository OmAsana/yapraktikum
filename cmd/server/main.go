package main

import (
	"fmt"
	"net/http"
)

func logIncoming(w http.ResponseWriter, req *http.Request) {
	fmt.Println(req.URL)

}

func main() {
	http.HandleFunc("/", logIncoming)
	http.ListenAndServe("127.0.0.1:8080", nil)

}
