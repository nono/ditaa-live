package main

import (
	"code.google.com/p/go.net/websocket"
	"github.com/howeyc/fsnotify"
	"flag"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

type Page struct {
	Filename string
	Addr     string
}

const (
	execname  = "ditaa"
	index_tpl = `
<html lang="en">
<head>
  <meta charset="utf-8">
  <title>ditaa-live</title>
</head>
<body>
  <h1>Ditaa-live</h1>
  <ul id="listing">
  {{range .}}
    <li><a href="/html/{{.}}">{{.}}</a></li>
  {{end}}
  </ul>
</body>
</html>
`
	page_tpl  = `
{{define "page"}}<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <title>{{.Filename}}</title>
</head>
<body>
  <h1>{{.Filename}}</h1>
  <img id="ditaa" src="/png/{{.Filename}}" />
  <script src="http://code.jquery.com/jquery-1.7.1.min.js"></script>
  <script>
  if (window["MozWebSocket"]) window.WebSocket = window.MozWebSocket;
  var ws = new WebSocket("ws://{{.Addr}}/notify")  // XXX
    , img = $("#ditaa")
    , path = $("h1").text();
  ws.onopen = function() {
    ws.send('"' + path + '"');
  };
  ws.onmessage = function(msg) {
    var ts = +new Date();
    img.attr({ src: "/png/" + path + "?" + ts});
  };
  </script>
</body>
</html>
{{end}}
`
)

var addr string

func listing() []string {
	entries, err := ioutil.ReadDir(".")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	files := make([]string, len(entries))
	for i, f := range entries {
		files[i] = f.Name()
	}
	return files
}

func notify(ws *websocket.Conn) {
	var filename string
	defer ws.Close()
	err := websocket.JSON.Receive(ws, &filename)
	if err != nil {
		fmt.Println(err)
		return
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer watcher.Close()

	for {
		time.Sleep(10 * time.Millisecond)
		err = watcher.Watch(filename)
		if err != nil {
			fmt.Println(err)
			return
		}

		select {
		case ev := <-watcher.Event:
			watcher.RemoveWatch(filename)
			err = websocket.JSON.Send(ws, ev)
			if err != nil {
				fmt.Println(err)
				return
			}
		case err = <-watcher.Error:
			fmt.Println(err)
			return
		}
	}
}

func image(w http.ResponseWriter, filename string) {
	_, err := os.Lstat(filename)
	if err != nil {
		notfound := fmt.Sprintf("Error 404: %s has not been found!\n", filename)
		http.Error(w, notfound, 404)
		return
	}

	tmpfile, err := ioutil.TempFile("", execname)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	tmpname := tmpfile.Name()
	defer os.Remove(tmpname)
	defer tmpfile.Close()

	cmd := exec.Command(execname, filename, "-o", tmpname)
	err = cmd.Run()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	tmpfile.Close()
	tmpfile, err = os.Open(tmpname)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	io.Copy(w, tmpfile)
}

func page(w http.ResponseWriter, filename string) {
	t, err := template.New("page").Parse(page_tpl)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
	t.ExecuteTemplate(w, "page", &Page{filename, addr})
}

func index(w http.ResponseWriter) {
	t, err := template.New("index").Parse(index_tpl)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
	t.ExecuteTemplate(w, "index", listing())
}

func dispatch(w http.ResponseWriter, r *http.Request) {
	png := "/png/"
	html := "/html/"
	path := r.URL.Path
	switch {
	case path == "/":
		index(w)
	case strings.HasPrefix(path, html):
		filename := r.URL.Path[len(html):]
		page(w, filename)
	case strings.HasPrefix(path, png):
		filename := r.URL.Path[len(png):]
		image(w, filename)
	default:
		http.NotFound(w, r)
	}
}

func main() {
	flag.StringVar(&addr, "addr", "127.0.0.1:4444", "Bind to this address:port")
	flag.Parse()
	args := flag.Args()
	if len(args) > 0 {
		os.Chdir(args[0])
	}
	fmt.Printf("Listening on http://%s/\n", addr)

	http.Handle("/notify", websocket.Handler(notify))
	http.HandleFunc("/", dispatch)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Can't listen on %s\n", addr)
		os.Exit(1)
	}
}
