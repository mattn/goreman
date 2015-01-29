package main

import (
	"flag"
	"github.com/hoisie/web"
)

var addr = flag.String("a", ":8080", "address")

func main() {
	flag.Parse()
	web.Get("/", func() string {
		return "Hello World"
	})
	web.Run(*addr)
}
