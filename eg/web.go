package main

import (
	"flag"
	"fmt"
	"net/http"
)

var addr = flag.String("a", ":8080", "address")

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "")
	})
	http.ListenAndServe(*addr, nil)
}
