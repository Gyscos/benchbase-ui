package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

func main() {
	var port int
	var host string

	flag.IntVar(&port, "p", 80, "Port to listen to.")
	flag.StringVar(&host, "host", "http://localhost:6666", "Benchbase host to connect to.")

	flag.Parse()

	setupHandlers()

	log.Println("Now listening to port", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", port), nil))
}
