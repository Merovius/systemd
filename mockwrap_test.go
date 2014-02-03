package systemd

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"syscall"
	"time"
)

type mockedCall struct {
	Name   string
	Args   []interface{}
	Return []interface{}
}

type mock []mockedCall

type argError struct {
	Fun      string
	Expected interface{}
	Got      interface{}
}

func (e argError) Error() string {
	return fmt.Sprintf("Expected %v, got %v in mock.%s", e.Expected, e.Got, e.Fun)
}

type mockFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

func (fi *mockFileInfo) Name() string {
	return fi.name
}

func (fi *mockFileInfo) Size() int64 {
	return fi.size
}

func (fi *mockFileInfo) Mode() os.FileMode {
	return fi.mode
}

func (fi *mockFileInfo) ModTime() time.Time {
	return fi.modTime
}

func (fi *mockFileInfo) IsDir() bool {
	return fi.mode.IsDir()
}

func (fi *mockFileInfo) Sys() interface{} {
	return nil
}

// getCall is a helper, that returns the first mocked call, advances m and
// tests, whether the called function is the correct one
func (m *mock) getCall() mockedCall {
	if len(*m) == 0 {
		panic("Too many calls to mock object")
	}
	c := (*m)[0]
	(*m) = (*m)[1:]

	pc, _, _, _ := runtime.Caller(1)
	fun := runtime.FuncForPC(pc)
	name := fun.Name()
	sep := strings.LastIndex(name, ".")
	name = name[sep+1:]

	if c.Name != name {
		panic(fmt.Errorf("Expected mock.%s call, got mock.%s", name, c.Name))
	}
	return c
}

func (m *mock) Getenv(key string) string {
	c := m.getCall()

	if key != c.Args[0].(string) {
		panic(argError{"Getenv", c.Args[0], key})
	}

	return c.Return[0].(string)
}

func (m *mock) Getpid() int {
	c := m.getCall()

	return c.Return[0].(int)
}

func (m *mock) Lstat(name string) (fi os.FileInfo, err error) {
	c := m.getCall()

	if name != c.Args[0].(string) {
		panic(argError{"Lstat", c.Args[0], name})
	}

	if c.Return[1] == nil {
		return c.Return[0].(*mockFileInfo), nil
	} else {
		return c.Return[0].(*mockFileInfo), c.Return[1].(error)
	}
}

func (m *mock) NewFile(fd uintptr, name string) *os.File {
	c := m.getCall()

	if c.Args[0].(uintptr) != fd {
		panic(argError{"NewFile", c.Args[0], fd})
	}

	if c.Args[1].(string) != name {
		panic(argError{"NewFile", c.Args[1], name})
	}

	return c.Return[0].(*os.File)
}

func (m *mock) Setenv(key, val string) error {
	panic("Not implemented Setenv")
}

func (m *mock) Fstat(fd int, st *syscall.Stat_t) error {
	c := m.getCall()

	if c.Args[0].(int) != fd {
		panic(argError{"Fstat", c.Args[0], fd})
	}

	*st = c.Return[0].(syscall.Stat_t)
	if n, ok := c.Return[1].(syscall.Errno); ok && n != 0 {
		return c.Return[1].(syscall.Errno)
	} else {
		return nil
	}
}

func (m *mock) Syscall(trap, a1, a2, a3 uintptr) (r1, r2 uintptr, err syscall.Errno) {
	c := m.getCall()

	if c.Args[0].(uintptr) != trap {
		panic(argError{"Syscall", c.Args[0], trap})
	}

	if c.Args[1].(uintptr) != a1 {
		panic(argError{"Syscall", c.Args[1], a1})
	}

	if c.Args[2].(uintptr) != a2 {
		panic(argError{"Syscall", c.Args[2], a2})
	}

	if c.Args[3].(uintptr) != a3 {
		panic(argError{"Syscall", c.Args[3], a3})
	}

	return c.Return[0].(uintptr), c.Return[1].(uintptr), c.Return[2].(syscall.Errno)
}
