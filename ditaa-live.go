package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
)

const (
	execname = "ditaa"
)

type Server struct {
	filename string
}

func (s Server) fail(w http.ResponseWriter, err error) {
	fmt.Fprintln(os.Stderr, err)
	fmt.Fprintln(w, "Error")
}

func (s Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tmpfile, err := ioutil.TempFile("", "ditaa")
	if err != nil {
		s.fail(w, err)
		return
	}
	tmpname := tmpfile.Name()
	defer os.Remove(tmpname)

	cmdline := fmt.Sprintf("%s %s -o %s", execname, s.filename, tmpname)
	fmt.Printf("cmdline: %s\n", cmdline)

	cmd := exec.Command(execname, s.filename, "-o", tmpname)
	err = cmd.Run()
	if err != nil {
		s.fail(w, err)
		return
	}

	tmpfile.Close()
	tmpfile, err = os.Open(tmpname)
	if err != nil {
		s.fail(w, err)
		return
	}
	io.Copy(w, tmpfile)
	tmpfile.Close()
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

	server := &Server{args[0]}
	fmt.Printf("Listening on http://%s/\n", addr)
	fmt.Printf("Watching %s\n", server.filename)

	err := http.ListenAndServe(addr, server)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Can't listen on %s\n", addr)
		os.Exit(1)
	}
}
