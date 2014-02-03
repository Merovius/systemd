package systemd

import (
	"fmt"
	"strconv"
	"time"
)

// IsWatchdogActive returns, whether the service manager expects watchdog
// keep-alive pings to be sent regularly via NotifyWatchdog(). The returned
// duration is the duration after which the service manager acts on the
// process. It is recommended to generate keep-alive pings every half of the
// returned time.
func IsWatchdogActive() (bool, time.Duration, error) {
	// Get PID that is watched
	e := osm.Getenv("WATCHDOG_PID")
	if e == "" {
		return false, 0, nil
	}
	l, err := strconv.Atoi(e)
	if err != nil {
		return false, 0, fmt.Errorf("Could not parse WATCHDOG_PID \"%s\"", e)
	}

	// Is it us?
	if l != osm.Getpid() {
		return false, 0, nil
	}

	// Get feedback-interval
	e = osm.Getenv("WATCHDOG_USEC")
	if e == "" {
		return false, 0, fmt.Errorf("WATCHDOG_PID set, but not WATCHDOG_USEC")
	}
	l, err = strconv.Atoi(e)
	if err != nil {
		return false, 0, fmt.Errorf("Could not parse WATCHDOG_USEC \"%s\"", e)
	}

	// The duration should be positive
	if l <= 0 {
		return false, 0, fmt.Errorf("Invalid interval %d", l)
	}

	return true, time.Duration(l) * time.Microsecond, nil
}

// AutoWatchdog tests whether the watchdog is active and if so, sends a
// keep-alive ping at the recommended interval in a seperate goroutine. It
// returns, whether a watchdog is active, the intervals at which a ping will be
// sent and whether an error occured during the initialization. Any errors
// occuring when pinging will be ignored.
func AutoWatchdog() (bool, time.Duration, error) {
	active, interval, err := IsWatchdogActive()
	if !active || err != nil {
		return active, 0, err
	}

	interval = interval / 2

	// For good measure we immediately send a ping. This will also catch
	// most connection issues.
	err = NotifyWatchdog()
	if err != nil {
		return false, 0, err
	}

	go func(interval time.Duration) {
		tick := time.Tick(interval)
		for _ = range tick {
			NotifyWatchdog()
		}
	}(interval)

	return active, interval, err
}
