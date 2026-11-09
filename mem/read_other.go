//go:build !linux && !windows

package mem

import (
	"fmt"
	"runtime"
)

// Available is not supported on this platform.
func Available() (uint64, error) {
	return 0, fmt.Errorf("reading free memory is not supported on %s (Windows and Linux are)", runtime.GOOS)
}
