package main

import (
	"code.google.com/p/go.net/websocket"
	"exp/inotify"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

const execname = "ditaa"

var addr string

func listing(ws *websocket.Conn) {
	defer ws.Close()
	entries, err := ioutil.ReadDir(".")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	files := make([]string, len(entries))
	for i, f := range entries {
		files[i] = f.Name()
	}
	err = websocket.JSON.Send(ws, files)
	if err != nil {
		fmt.Println(err)
	}
}

func notify(ws *websocket.Conn) {
	var filename string
	defer ws.Close()
	err := websocket.JSON.Receive(ws, &filename)
	if err != nil {
		fmt.Println(err)
		return
	}

	watcher, err := inotify.NewWatcher()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer watcher.Close()

	for {
		time.Sleep(10 * time.Millisecond)
		err = watcher.AddWatch(filename, inotify.IN_MODIFY)
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
	fmt.Fprintf(w, `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <title>%s</title>
</head>
<body>
  <h1>%s</h1>
  <img id="ditaa" src="/png/%s" />
  <script src="http://code.jquery.com/jquery-1.7.1.min.js"></script>
  <script>
  var ws = new WebSocket("ws://%s/notify")
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
`, filename, filename, filename, addr)
}

func index(w http.ResponseWriter) {
	fmt.Fprintf(w, `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <title>ditaa-live</title>
</head>
<body>
  <h1>Ditaa-live</h1>
  <ul id="listing"></ul>
  <script src="http://code.jquery.com/jquery-1.7.1.min.js"></script>
  <script>
  var ws = new WebSocket("ws://%s/ls")
    , ls = $("#listing");
  ws.onmessage = function(msg) {
    var files = jQuery.parseJSON(msg.data)
    for (var i in files) {
      $("<li><a href='/html/" + files[i] + "'>" + files[i] + "</a>")
        .appendTo(ls);
    }
  };
  </script>
</body>
</html>
`, addr)
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
	http.Handle("/ls", websocket.Handler(listing))
	http.HandleFunc("/", dispatch)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Can't listen on %s\n", addr)
		os.Exit(1)
	}
}
