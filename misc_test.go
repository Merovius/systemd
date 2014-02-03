package systemd

import (
	"errors"
	"os"
	"testing"
)

func TestIsSystemdBootet(t *testing.T) {
	osm = &mock{
		{
			"Lstat",
			[]interface{}{ "/run/systemd/system" },
			[]interface{}{ &mockFileInfo{ mode: os.ModeDir}, error(nil) },
		},
	}

	if !IsSystemdBootet() {
		t.Fail()
	}

	osm = &mock{
		{
			"Lstat",
			[]interface{}{ "/run/systemd/system" },
			[]interface{}{ &mockFileInfo{ mode: os.FileMode(0) }, error(nil) },
		},
	}

	if IsSystemdBootet() {
		t.Fail()
	}

	osm = &mock{
		{
			"Lstat",
			[]interface{}{ "/run/systemd/system" },
			[]interface{}{ (*mockFileInfo)(nil), errors.New("mock error") },
		},
	}

	if IsSystemdBootet() {
		t.Fail()
	}
}

func TestClearEnv(t *testing.T) {
	// We don't use a mock here. We just screw with the environment
	osm = &osPackage{}

	var testcases = []struct{
		mask EnvMask
		env  [5]string
	}{
		{ 0, [5]string{ "", "", "", "", "" } },
		{ ListenPid, [5]string{ "T", "", "", "", "" } },
		{ ListenFds, [5]string{ "", "T", "", "", "" } },
		{ NotifySocket, [5]string{ "", "", "T", "", "" } },
		{ WatchdogPid, [5]string{ "", "", "", "T", "" } },
		{ WatchdogUsec, [5]string{ "", "", "", "", "T" } },
	}

	setupEnv := func() {
		os.Setenv("LISTEN_PID", "T")
		os.Setenv("LISTEN_FDS", "T")
		os.Setenv("NOTIFY_SOCKET", "T")
		os.Setenv("WATCHDOG_PID", "T")
		os.Setenv("WATCHDOG_USEC", "T")
	}
	testEnv := func(env [5]string) bool {
		if os.Getenv("LISTEN_PID") != env[0] {
			return false
		}
		if os.Getenv("LISTEN_FDS") != env[1] {
			return false
		}
		if os.Getenv("NOTIFY_SOCKET") != env[2] {
			return false
		}
		if os.Getenv("WATCHDOG_PID") != env[3] {
			return false
		}
		if os.Getenv("WATCHDOG_USEC") != env[4] {
			return false
		}
		return true
	}

	for _, tc := range testcases {
		setupEnv()
		ClearEnv(tc.mask)
		if !testEnv(tc.env) {
			t.Fail()
		}
	}
}
