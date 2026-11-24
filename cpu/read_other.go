//go:build !linux && !windows

package cpu

import (
	"fmt"
	"runtime"
)

// Read is not supported on this platform.
func Read() (Times, error) {
	return Times{}, fmt.Errorf("reading CPU use is not supported on %s (Windows and Linux are)", runtime.GOOS)
}
