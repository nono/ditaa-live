package main

import (
	"flag"
	"fmt"
	"github.com/bmizerany/pat.go"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
)

const (
	execname = "ditaa"
)

func fail(w http.ResponseWriter, err error) {
	fmt.Fprintln(os.Stderr, err)
	fmt.Fprintln(w, "Error")
}

func index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Index\n")
}

func image(w http.ResponseWriter, r *http.Request) {
	tmpfile, err := ioutil.TempFile("", execname)
	if err != nil {
		fail(w, err)
		return
	}
	tmpname := tmpfile.Name()
	defer os.Remove(tmpname)
	defer tmpfile.Close()

	filename := r.URL.Query().Get(":filename")
	cmd := exec.Command(execname, filename, "-o", tmpname)
	err = cmd.Run()
	if err != nil {
		fail(w, err)
		return
	}

	tmpfile.Close()
	tmpfile, err = os.Open(tmpname)
	if err != nil {
		fail(w, err)
		return
	}
	io.Copy(w, tmpfile)
}

func main() {
	var addr string
	flag.StringVar(&addr, "addr", "127.0.0.1:4444", "Bind to this address:port")
	flag.Parse()
	args := flag.Args()
	if len(args) > 0 {
		os.Chdir(args[0])
	}
	fmt.Printf("Listening on http://%s/\n", addr)

	m := pat.New()
	m.Get("/", http.HandlerFunc(index))
	m.Get("/png/:filename", http.HandlerFunc(image))
	err := http.ListenAndServe(addr, m)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Can't listen on %s\n", addr)
		os.Exit(1)
	}
}
