package systemd

// EnvMask encodes a set of environment-variables that should be cleared by
// ClearEnv.
type EnvMask uint32

const (
	ListenPid EnvMask = (1 << iota)
	ListenFds
	NotifySocket
	WatchdogPid
	WatchdogUsec

	Listen = ListenPid | ListenFds
	Watchdog = WatchdogPid | WatchdogUsec
)

// IsSystemdBootet returns, whether the running system is bootet by systemd.
func IsSystemdBootet() bool {
	fi, err := osm.Lstat("/run/systemd/system")
	if err != nil {
		return false
	}

	return fi.IsDir()
}

// ClearEnv removes the systemd-specific environment variables, leaving only
// the specified ones intact. It is recommended to call this once after startup
// is completed.
func ClearEnv(except EnvMask) {
	unset := func(str string) {
		// BUG(aw): Go does not support unsetenv yet. Instead we only set the
		// variables to the empty string.
		// see https://code.google.com/p/go/issues/detail?id=6423
		osm.Setenv(str, "")
	}

	if except & ListenPid == 0 {
		unset("LISTEN_PID")
	}
	if except & ListenFds == 0 {
		unset("LISTEN_FDS")
	}
	if except & NotifySocket == 0 {
		unset("NOTIFY_SOCKET")
	}
	if except & WatchdogPid == 0 {
		unset("WATCHDOG_PID")
	}
	if except & WatchdogUsec == 0 {
		unset("WATCHDOG_USEC")
	}
}

// TODO: We really shouldn't have this. Find a way to not need it
// CheckPath checks of the given file is the same as the given path.
//func CheckPath(f *os.File, path string) (bool, error) {
//	s, err := f.Stat()
//	if err != nil {
//		return false, err
//	}
//
//	st_fd, ok := s.Sys().(syscall.Stat_t)
//	if !ok {
//		return false, errors.New("Unknown stat implementation")
//	}
//
//	var st_path syscall.Stat_t
//	err = syscall.Stat(path, &st_path)
//	if err != nil {
//		if err == syscall.ENOENT || err == syscall.ENOTDIR {
//			return false, nil
//		}
//
//		return false, err
//	}
//


//	switch {
//	case st_fd.Mode & syscall.S_IFREG != 0 && st_path.Mode & syscall.S_IFREG != 0:
//		fallthrough
//	case st_fd.Mode & syscall.S_IFIFO != 0 && st_path.Mode & syscall.S_IFIFO != 0:
//		return st_fd.Dev == st_path.Dev && st_fd.Ino == st_path.Ino, nil
//	case st_fd.Mode & syscall.S_IFCHR != 0 && st_path.Mode & syscall.S_IFCHR != 0:
//		return st_fd.Rdev == st_path.Rdev, nil
//	default:
//		return false, errors.New("Unknown type of file")
//	}
//
//}
