package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
)

type MyServer struct{}

func (my MyServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	fmt.Fprint(w, "hello, world!\n")
}

func main() {
	var addr string
	flag.StringVar(&addr, "addr", "127.0.0.1:4444", "Bind to this address:port")
	flag.Parse()
	args := flag.Args()
	if len(args) != 1 {
		fmt.Println("Usage: splint [options] <go file>...")
		flag.PrintDefaults()
		os.Exit(1)
	}
	filename := args[0]

	var server MyServer
	err := http.ListenAndServe(addr, server)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Can't listen on %s\n", addr)
		os.Exit(1)
	}

	fmt.Printf("Listening on http://%s/\n", addr)
	fmt.Printf("Watching %s\n", filename)
}
