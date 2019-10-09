// +build ignore

package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
)

var addr = flag.String("a", ":8080", "address")

func main() {
	flag.Parse()
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "hello "+os.Getenv("AUTHOR"))
	})
	http.ListenAndServe(*addr, nil)
}
