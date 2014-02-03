// package systemd implements some systemd APIs
package systemd

import (
	"errors"
	"fmt"
	"net"
	"os"
	"reflect"
	"strconv"
	"sync"
	"syscall"
)

const (
	listenFdsStart = 3
)

var (
	num_fds = -1
	lock    = sync.Mutex{}
)

// listenFds checks for file descriptors passed by the system manager. It
// returns the number of file descriptors received and an error, if one
// occured.
func listenFds() (num_fds int, err error) {
	// Get PID the fds are for
	e := osm.Getenv("LISTEN_PID")
	if e == "" {
		return 0, nil
	}
	l, err := strconv.Atoi(e)
	if err != nil {
		return 0, err
	}

	// Are they for us?
	if l != osm.Getpid() {
		return 0, nil
	}

	// Get number of filedescriptors
	e = osm.Getenv("LISTEN_FDS")
	if e == "" {
		return 0, nil
	}
	l, err = strconv.Atoi(e)
	if err != nil {
		return 0, err
	}
	if l < 0 {
		return 0, fmt.Errorf("Got negative number of filedescriptors")
	}

	for i := listenFdsStart; i < listenFdsStart+l; i++ {
		flags, _, err := osm.Syscall(syscall.SYS_FCNTL, uintptr(i), syscall.F_GETFL, 0)
		if err != 0 {
			return 0, err
		}

		if flags&syscall.FD_CLOEXEC != 0 {
			continue
		}

		flags |= syscall.FD_CLOEXEC

		_, _, err = osm.Syscall(syscall.SYS_FCNTL, uintptr(i), syscall.F_SETFL, flags)
		if err != 0 {
			return 0, err
		}
	}

	return l, nil
}

// getFile checks if the file descriptor at the given offset is a FIFO or a
// special file and returns an open file to this file descriptor (But see BUGS
// below)
// BUG(aw): Currently the path to the file is not correctly set, i.e. f.Name()
// does not work.
func getFile(num int) (f *os.File, err error) {
	fd := listenFdsStart + num

	var st syscall.Stat_t
	err = osm.Fstat(fd, &st)
	if err != nil {
		return nil, err
	}

	if st.Mode & syscall.S_IFIFO == 0 && st.Mode & syscall.S_IFREG == 0 && st.Mode & syscall.S_IFCHR == 0 {
		return nil, errors.New("File descriptor is not a fifo or special file")
	}

	return osm.NewFile(uintptr(fd), ""), nil
}

// getConn checks if the file descriptor at the given offset is a socket and
// returns a net.Conn bound to this file descriptor.
func getConn(num int) (conn net.Conn, err error) {
	fd := listenFdsStart + num

	f := osm.NewFile(uintptr(fd), "")
	conn, err = net.FileConn(f)
	f.Close()

	return conn, err
}

// getIPConn checks if the file descriptor at the given offset is a tcp-socket
// and returns a net.IPConn bound to this file descriptor.
func getIPConn(num int) (conn *net.IPConn, err error) {
	c, err := getConn(num)
	if err != nil {
		return nil, err
	}

	var ok bool
	if conn, ok = c.(*net.IPConn); !ok {
		return nil, errors.New("Not a IPConn")
	}

	return conn, nil
}

// getTCPConn checks if the file descriptor at the given offset is a tcp-socket
// and returns a net.TCPConn bound to this file descriptor.
func getTCPConn(num int) (conn *net.TCPConn, err error) {
	c, err := getConn(num)
	if err != nil {
		return nil, err
	}

	var ok bool
	if conn, ok = c.(*net.TCPConn); !ok {
		return nil, errors.New("Not a TCPConn")
	}

	return conn, nil
}

// getUDPConn checks if the file descriptor at the given offset is a tcp-socket
// and returns a net.UDPConn bound to this file descriptor.
func getUDPConn(num int) (conn *net.UDPConn, err error) {
	c, err := getConn(num)
	if err != nil {
		return nil, err
	}

	var ok bool
	if conn, ok = c.(*net.UDPConn); !ok {
		return nil, errors.New("Not a UDPConn")
	}

	return conn, nil
}

// getUnixConn checks if the file descriptor at the given offset is a tcp-socket
// and returns a net.UnixConn bound to this file descriptor.
func getUnixConn(num int) (conn *net.UnixConn, err error) {
	c, err := getConn(num)
	if err != nil {
		return nil, err
	}

	var ok bool
	if conn, ok = c.(*net.UnixConn); !ok {
		return nil, errors.New("Not a UnixConn")
	}

	return conn, nil
}

// getListener checks if the file descriptor at the given offset is a socket
// and returns a net.Listener bound to this file descriptor.
func getListener(num int) (conn net.Listener, err error) {
	fd := listenFdsStart + num

	return net.FileListener(osm.NewFile(uintptr(fd), ""))
}

// getTCPListener checks if the file descriptor at the given offset is a
// tcp-socket in accepting mode and returns a net.TCPListener bound to this file
// descriptor.
func getTCPListener(num int) (conn *net.TCPListener, err error) {
	l, err := getListener(num)
	if err != nil {
		return nil, err
	}

	var ok bool
	if conn, ok = l.(*net.TCPListener); !ok {
		return nil, errors.New("Not a TCPListener")
	}

	return conn, nil
}

// getUnixListener checks if the file descriptor at the given offset is a
// tcp-socket in accepting mode and returns a net.UnixListener bound to this file
// descriptor.
func getUnixListener(num int) (conn *net.UnixListener, err error) {
	l, err := getListener(num)
	if err != nil {
		return nil, err
	}

	var ok bool
	if conn, ok = l.(*net.UnixListener); !ok {
		return nil, errors.New("Not a UnixListener")
	}

	return conn, nil
}

// GetPassedFiles acquires a number of open filedescriptors using the systemd
// socket activation protocol. The passed fds will be put into the given
// targets in ascending order. The targets are interpreted as follows:
//
// If the target is a pointer to a non-slice, the respective file descriptor is
// checked against the type of the value pointed to and stored in it.
//
// If the target is a slice of pointers, then every element of t[0:len(t)]
// will have one file descriptor placed in it.
//
// If the target is a pointer to a slice of pointers, and it is nil, all remaining
// filedescriptors will be placed in a newly allocated slice which is then
// written to the pointer location. If it is not nil, the filedescriptors will
// be placed in the slice, as above
//
// Everything else is an error. It is not an error to provide more targets then
// there are passed file descriptors, but it is an error to provide fewer.
//
// For example, a web server, that wants to provide a local administrative unix
// domain socket and have systemd open an arbitrary number of connections for
// it, might call:
//
//		var control net.UnixListener
//		var listeners []*net.TCPListener
//		fds, err := SockedActivation(true, &control, &listeners)
func GetPassedFiles(targets ...interface{}) (n int, err error) {
	fds, err := listenFds()
	if err != nil {
		return fds, err
	}

	k := 0

	for i := 0; i < fds; i++ {
		rv := reflect.ValueOf(targets[k])

		// Target is a pointer to a non-slice
		if rv.Kind() == reflect.Ptr && rv.Elem().Kind() != reflect.Slice {
			// Why zero value?
			err = storeFd(i, rv.Elem())
			if err != nil {
				return 0, err
			}
			continue
		}

		// Target is a slice
		if rv.Kind() == reflect.Slice {
			for j := 0; j < rv.Len() && i < fds; j++ {
				rv2 := rv.Index(j)
				err = storeFd(i, rv2.Addr())
				if err != nil {
					return 0, err
				}
				i++
			}
			continue
		}

		// Target is a pointer to a slice
		if rv.Kind() == reflect.Ptr && rv.Elem().Kind() == reflect.Slice {
			if rv.Elem().IsNil() {
				rv.Elem().Set(reflect.MakeSlice(rv.Elem().Type(), fds - i, fds - i))
			}

			rv2 := rv.Elem()
			for j := 0; j < rv2.Len() && i < fds; j++ {
				rv3 := rv.Elem().Index(j)
				err = storeFd(i, rv3.Addr())
				if err != nil {
					return 0, err
				}
				i++
			}
			continue
		}

		return 0, fmt.Errorf("Unhandled type %v", rv.Type())
	}
	return fds, nil
}

// storeFd assumes, that target is a settable Value and tries to store one fd in it
func storeFd(num int, target reflect.Value) error {
	// Target is a net.Conn
	if target.Type() == reflect.TypeOf((*net.Conn)(nil)).Elem() {
		c, err := getConn(num)
		if err != nil {
			return err
		}
		target.Set(reflect.ValueOf(c))
		return nil
	}

	// Target is a *net.IPConn
	if target.Type() == reflect.TypeOf((*net.IPConn)(nil)) {
		c, err := getIPConn(num)
		if err != nil {
			return err
		}
		target.Set(reflect.ValueOf(c))
		return nil
	}

	// Target is a *net.TCPConn
	if target.Type() == reflect.TypeOf((*net.TCPConn)(nil)) {
		c, err := getTCPConn(num)
		if err != nil {
			return err
		}
		target.Set(reflect.ValueOf(c))
		return nil
	}

	// Target is a *net.UDPConn
	if target.Type() == reflect.TypeOf((*net.UDPConn)(nil)) {
		c, err := getUDPConn(num)
		if err != nil {
			return err
		}
		target.Set(reflect.ValueOf(c))
		return nil
	}

	// Target is a *net.UnixConn
	if target.Type() == reflect.TypeOf((*net.UnixConn)(nil)) {
		c, err := getUnixConn(num)
		if err != nil {
			return err
		}
		target.Set(reflect.ValueOf(c))
		return nil
	}

	// Target is a net.Listener
	if target.Type() == reflect.TypeOf((*net.Listener)(nil)).Elem() {
		c, err := getListener(num)
		if err != nil {
			return err
		}
		target.Set(reflect.ValueOf(c))
		return nil
	}

	// Target is a *net.TCPListener
	if target.Type() == reflect.TypeOf((*net.TCPListener)(nil)) {
		l, err := getTCPListener(num)
		if err != nil {
			return err
		}
		target.Set(reflect.ValueOf(l))
		return nil
	}

	// Target is a *net.UnixListener
	if target.Type() == reflect.TypeOf((*net.UnixListener)(nil)) {
		l, err := getUnixListener(num)
		if err != nil {
			return err
		}
		target.Set(reflect.ValueOf(l))
		return nil
	}

	// Target is a *os.File
	if target.Type() == reflect.TypeOf((*os.File)(nil)) {
		f, err := getFile(num)
		if err != nil {
			return err
		}
		target.Set(reflect.ValueOf(f))
		return nil
	}

	return errors.New("Unknown target type")
}
