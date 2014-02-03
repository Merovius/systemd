About
===

These are pure-go implementations of some systemd-APIs for daemon-authors
(reference implementation is [sd-daemon.h](http://www.freedesktop.org/software/systemd/man/sd-daemon.html).
The idea is, to make it as simple as possible to write systemd-aware daemons.
Logging can happen to stderr, so the `log`-package is sufficient there for most
cases. systemd also provides APIs for [socket activation](http://0pointer.de/blog/projects/socket-activation.html),
[startup notifications]() and [a software watchdog](http://0pointer.de/blog/projects/watchdog.html)), built on top of that.

We try to expose those features as simply (and idiomatically) as possible. For example, a socket activated http-server is as simple as
```go
package main

import (
	"github.com/Merovius/systemd"
	"log"
	"net"
	"net/http"
)

func Handle(res http.ResponseWriter, req *http.Request) {
	res.Write([]byte("Hello world"))
}

func main() {
	var l net.Listener
	n, err := systemd.GetPassedFiles(&l)
	if n < 1 {
		log.Fatal("Not enough sockets passed")
	}
	if err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("/", Handle)
	err = http.Serve(l, nil)
	log.Fatal(err)
}
```

A watchdog-aware service is as simple as
```go
package main

import (
	"github.com/Merovius/systemd"
	"log"
)

func main() {
	wd, _, err := systemd.AutoWatchdog()
	if err != nil {
		log.Fatal(err)
	}
	// Do awesome stuff
}
```

And the startup-notification protocol is handled as simple as
```go
package main

import (
	"github.com/Merovius/systemd"
)

func main() {
	// Do initialization stuff
	if err := systemd.NotifyReady(); err != nil {
		log.Println(err)
	}
}
```

Status
===

A while after I started working at this and a bit before I actually wanted to
push it out, Brandon Philips
[announced](https://groups.google.com/forum/#!topic/golang-nuts/XcGjI-qCgTs) a
similar library. His and this library mainly share the socket-activation
functionality. I'm currently evaluating whether it makes more sense to
contribute to his library or maintain my own approach. This repository is only
meant *to enable this evaluation*. Otherwise I would have polished the code as
well as the API a little bit more before publishing. So *please* don't use this
yet (and if you do, don't depend on it being actively maintained in the future
or that the API does not change, IOW do it at your own risk).
