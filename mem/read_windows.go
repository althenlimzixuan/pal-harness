package mem

import (
	"fmt"
	"syscall"
	"unsafe"
)

// memoryStatusEx is Windows' MEMORYSTATUSEX.
type memoryStatusEx struct {
	length               uint32
	memoryLoad           uint32
	totalPhys            uint64
	availPhys            uint64
	totalPageFile        uint64
	availPageFile        uint64
	totalVirtual         uint64
	availVirtual         uint64
	availExtendedVirtual uint64
}

var globalMemoryStatusEx = syscall.NewLazyDLL("kernel32.dll").NewProc("GlobalMemoryStatusEx")

// Available returns the available physical memory (ullAvailPhys) in bytes.
func Available() (uint64, error) {
	var m memoryStatusEx
	m.length = uint32(unsafe.Sizeof(m))
	if r, _, err := globalMemoryStatusEx.Call(uintptr(unsafe.Pointer(&m))); r == 0 {
		return 0, fmt.Errorf("GlobalMemoryStatusEx: %v", err)
	}
	return m.availPhys, nil
}
