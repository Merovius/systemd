package systemd

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func TestWatchdogActive(t *testing.T) {
	// We don't use a mock, we just use the normal environment
	osm = &osPackage{}

	realpid := fmt.Sprintf("%d", os.Getpid())
	fakepid := fmt.Sprintf("%d", os.Getpid()+23)

	var testcases = []struct {
		watchdogPid  string
		watchdogUsec string
		shouldActive bool
		shouldUsec   time.Duration
		shouldError  bool
	}{
		{"", "", false, 0, false},
		{"unparseable pid", "", false, 0, true},
		{fakepid, "", false, 0, false},
		{realpid, "", false, 0, true},
		{realpid, "unparseable usec", false, 0, true},
		{realpid, "-1000", false, 0, true},
		{realpid, "42235", true, 42235*time.Microsecond, false},
	}

	for _, tc := range testcases {
		os.Setenv("WATCHDOG_PID", tc.watchdogPid)
		os.Setenv("WATCHDOG_USEC", tc.watchdogUsec)
		active, usec, err := IsWatchdogActive()
		if tc.shouldActive != active {
			t.Fail()
		}
		if tc.shouldUsec != usec {
			t.Fail()
		}
		if tc.shouldError != (err != nil) {
			t.Fail()
		}
	}
}
