package main

// systemd socket activation example
// This is a minimal web server, that listenes to all sockets passed by systemd
// and also provides a control-socket for use of the system-administrator.
import (
	"net/http"
	"net"
	"log"
	"github.com/Merovius/systemd"
)

type SimpleHandler struct{}

func (h SimpleHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	// We serve a very simple http-page
	res.Write([]byte(`<!DOCTYPE html><html><body>systemd and golang rock \o/</body></html>`))
}

func main() {
	var ctl net.Listener
	var https []net.Listener

	n, err := systemd.GetPassedFiles(&ctl, &https)
	if err != nil {
		log.Fatal(err)
	}
	if n < 2 {
		log.Fatal("Not enoug file descriptors passed")
	}

	for _, l := range https {
		go http.Serve(l, SimpleHandler{})
	}

	// Our control-socket is very primitive. We just exit, when someone
	// connects.
	ctl.Accept()

	// We close all http-listeners
	for _, l := range https {
		l.Close()
	}

	// And done
}
