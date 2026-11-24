package cpu

import (
	"fmt"
	"syscall"
	"unsafe"
)

var getSystemTimes = syscall.NewLazyDLL("kernel32.dll").NewProc("GetSystemTimes")

func ticks(f syscall.Filetime) uint64 { return uint64(f.HighDateTime)<<32 | uint64(f.LowDateTime) }

// Read returns the machine's processor times from GetSystemTimes.
func Read() (Times, error) {
	var idle, kernel, user syscall.Filetime
	if r, _, err := getSystemTimes.Call(uintptr(unsafe.Pointer(&idle)), uintptr(unsafe.Pointer(&kernel)), uintptr(unsafe.Pointer(&user))); r == 0 {
		return Times{}, fmt.Errorf("GetSystemTimes: %v", err)
	}
	return fromSystemTimes(ticks(idle), ticks(kernel), ticks(user)), nil
}
