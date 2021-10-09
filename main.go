package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/ping", ping)

	http.ListenAndServe(":8090", nil)
}

func ping(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "pong\n")
}