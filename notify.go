package systemd

import (
	"fmt"
	"net"
	"strings"
	"sync"
)

var (
	notifyConn *net.UnixConn
	notifyMtx  sync.Mutex
)

// NotifyConn returns a *net.UnixConn that can be used for special stuff that
// is not covered by other functions. If you do not have good reasons to need
// this, you should probably use the Notify* functions.
func NotifyConn() (*net.UnixConn, error) {
	e := osm.Getenv("NOTIFY_SOCKET")
	if e == "" {
		return nil, fmt.Errorf("No notification socket found")
	}

	if (e[0] != '@' && e[0] != '/') {
		return nil, fmt.Errorf("Notification socket must be an abstract socket or an absolute path")
	}

	addr, err := net.ResolveUnixAddr("unixgram", e)
	if err != nil {
		return nil, fmt.Errorf("Could not resolve socket addres: %v", err.Error())
	}

	conn, err := net.DialUnix("unixgram", nil, addr)
	if err != nil {
		return nil, fmt.Errorf("Could not connect to notification socket: %v", err.Error())
	}
	return conn, nil
}

// Notify sends a custom message to the system manager (see the manpage of
// sd_notify for more information). It can be used to implement own extensions
// to the startup-notification protocol. For everything else it is recommended
// to use one of the special notification-functions.
func Notify(state string) (err error) {
	notifyMtx.Lock()
	defer notifyMtx.Unlock()

	if notifyConn == nil {
		notifyConn, err = NotifyConn()
		if err != nil {
			return err
		}
	}

	defer func() {
		if err != nil {
			notifyConn.Close()
			notifyConn = nil
		}
	}()

	_, err = notifyConn.Write([]byte(state))
	if err != nil {
		return err
	}
	return nil
}

// NotifyReady tells the init system that daemon startup is finished.
func NotifyReady() error {
	return Notify("READY=1")
}

// NotifyStatus passes a single-line statusu string back to the init system
// that describes the daemon state.
func NotifyStatus(status string) error {
	if strings.ContainsAny(status, "\n\r") {
		return fmt.Errorf("Status may not contain newlines")
	}
	return Notify(fmt.Sprintf("STATUS=%s", status))
}

// NotifyErrno sends an errno-style error code to the init system.
func NotifyErrno(errno uint) error {
	return Notify(fmt.Sprintf("ERRNO=%d", errno))
}

// NotifyMainPid sends the main pid of the daemon, in case this process isn't it.
func NotifyMainPid(pid int) error {
	return Notify(fmt.Sprintf("MAINPID=%d", pid))
}

// NotifyWatchdog sends a keep-alive ping to the system manager.
func NotifyWatchdog() error {
	return Notify("WATCHDOG=1")
}
