package systemd

// This file is for mocking out the operating system function for testing.
// During production it will just pass everything on to the os-package, but we
// can test by implementing the osmock-interface

import (
	"os"
	"syscall"
)

var (
	osm osmock
)

type osmock interface {
	Getenv(key string) string
	Getpid() int
	Lstat(name string) (fi os.FileInfo, err error)
	NewFile(fd uintptr, name string) *os.File
	Setenv(key, val string) error

	Fstat(fd int, st *syscall.Stat_t) error
	Syscall(trap, a1, a2, a3 uintptr) (r1, r2 uintptr, err syscall.Errno)
}

type osPackage struct{}

func init() {
	osm = osPackage{}
}

func (o osPackage) Getenv(key string) string {
	return os.Getenv(key)
}

func (o osPackage) Getpid() int {
	return os.Getpid()
}

func (o osPackage) Lstat(name string) (fi os.FileInfo, err error) {
	return os.Lstat(name)
}

func (o osPackage) NewFile(fd uintptr, name string) *os.File {
	return os.NewFile(fd, name)
}

func (o osPackage) Setenv(key, val string) error {
	return os.Setenv(key, val)
}

func (o osPackage) Fstat(fd int, st *syscall.Stat_t) error {
	return syscall.Fstat(fd, st)
}

func (o osPackage) Syscall(trap, a1, a2, a3 uintptr) (r1, r2 uintptr, err syscall.Errno) {
	return syscall.Syscall(trap, a1, a2, a3)
}
