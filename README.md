Ditaa-live
==========

[Ditaa](http://ditaa.sourceforge.net/) is command-line utility written in
Java, that can convert diagrams drawn using ascii art into proper bitmap
graphics. Ditaa-live is a golang web server that show the generated images
from ditaa. But the most important thing is that the image is automatically
reloaded when the ditaa file is modified. It works by opening a websocket
connection to the server that will monitor the file with inotify (Linux-only).

**Note**: ditaa-live only works in a browser that supports WebSocket,
i.e. Chrome or Firefox.


How to use it?
--------------

Install golang-weekly (2012-02-07) and don't forget to set `$GOROOT`.
And then, run these instructions:

    go get github.com/nono/ditaa-live
    ditaa-live path/to/a/directory
    x-www-browser http://127.0.0.1:4444/


Credits
-------

♡2012 by Bruno Michel. Copying is an act of love. Please copy and share.

Released under the MIT license
