package systemd

import (
	"os"
	"syscall"
	"testing"
)

func TestListenFds(t *testing.T) {

	var testcases = []struct {
		ListenPid  string
		pid        int
		ListenFds  string
		GetflFlags uintptr
		GetflErr   syscall.Errno
		SetflFlags uintptr
		SetflErr   syscall.Errno

		NumFds int
		Err    bool
	}{
		{ListenPid: "", NumFds: 0, Err: false},
		{ListenPid: "unparseable pid", NumFds: 0, Err: true},
		{ListenPid: "1234", pid: 5678, NumFds: 0, Err: false},
		{ListenPid: "1234", pid: 1234, ListenFds: "", NumFds: 0, Err: false},
		{ListenPid: "1234", pid: 1234, ListenFds: "unparseable", NumFds: 0, Err: true},
		{ListenPid: "1234", pid: 1234, ListenFds: "-1", NumFds: 0, Err: true},
		{ListenPid: "1234", pid: 1234, ListenFds: "0", NumFds: 0, Err: false},
		{ListenPid: "1234", pid: 1234, ListenFds: "1", GetflErr: 1, NumFds: 0, Err: true},
		{ListenPid: "1234", pid: 1234, ListenFds: "1", GetflErr: 0, GetflFlags: syscall.FD_CLOEXEC, NumFds: 1, Err: false},
		{ListenPid: "1234", pid: 1234, ListenFds: "1", GetflErr: 0, GetflFlags: 0, SetflFlags: syscall.FD_CLOEXEC, SetflErr: 1, NumFds: 0, Err: true},
		{ListenPid: "1234", pid: 1234, ListenFds: "1", GetflErr: 0, GetflFlags: 0, SetflFlags: syscall.FD_CLOEXEC, SetflErr: 0, NumFds: 1, Err: false},
	}

	for _, tc := range testcases {
		osm = &mock{
			{"Getenv", []interface{}{"LISTEN_PID"}, []interface{}{tc.ListenPid}},
			{"Getpid", nil, []interface{}{tc.pid}},
			{"Getenv", []interface{}{"LISTEN_FDS"}, []interface{}{tc.ListenFds}},
			{"Syscall", []interface{}{uintptr(syscall.SYS_FCNTL), uintptr(3), uintptr(syscall.F_GETFL), uintptr(0)}, []interface{}{tc.GetflFlags, uintptr(0), tc.GetflErr}},
			{"Syscall", []interface{}{uintptr(syscall.SYS_FCNTL), uintptr(3), uintptr(syscall.F_SETFL), tc.SetflFlags}, []interface{}{uintptr(0), uintptr(0), tc.SetflErr}},
		}

		num_fds, err := listenFds()
		if num_fds != tc.NumFds {
			t.Fail()
		}
		if tc.Err != (err != nil) {
			t.Fail()
		}
	}
}

func TestGetFile(t *testing.T) {

	var testcases = []struct {
		Num      int
		Fstat    syscall.Stat_t
		FstatErr syscall.Errno
		Outfd    uintptr
		Err      bool
	}{
		{Num: 1234, FstatErr: 1, Err: true},
		{Num: 1234, FstatErr: 0, Fstat: syscall.Stat_t{Mode: 0}, Err: true},
		{Num: 1234, FstatErr: 0, Fstat: syscall.Stat_t{Mode: syscall.S_IFIFO}, Err: false, Outfd: 1237},
		{Num: 1234, FstatErr: 0, Fstat: syscall.Stat_t{Mode: syscall.S_IFREG}, Err: false, Outfd: 1237},
		{Num: 1234, FstatErr: 0, Fstat: syscall.Stat_t{Mode: syscall.S_IFCHR}, Err: false, Outfd: 1237},
	}

	for _, tc := range testcases {
		osm = &mock{
			{"Fstat", []interface{}{1237}, []interface{}{tc.Fstat, tc.FstatErr}},
			{"NewFile", []interface{}{uintptr(1237), ""}, []interface{}{os.NewFile(1237, "")}},
		}

		f, err := getFile(tc.Num)

		if (err != nil) != tc.Err {
			t.Fail()
		} else if err == nil && (f == nil || f.Fd() != tc.Outfd) {
			t.Fail()
		}
	}
}
