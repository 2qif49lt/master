package main

import (
	"fmt"
	"io"
	"net/http"
)

var (
	count = 0
)

func HelloServer(w http.ResponseWriter, req *http.Request) {
	count++
	s := fmt.Sprintf("hello %d  %s\n", count, req.RequestURI)
	io.WriteString(w, s)
}

func main() {
	err := http.ListenAndServe(":8181", http.HandlerFunc(HelloServer))
	fmt.Print(err)
}
