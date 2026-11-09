package mem

import "os"

// Available returns the available physical memory (MemAvailable) in bytes.
func Available() (uint64, error) {
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return 0, err
	}
	defer f.Close()
	return parseMeminfo(f)
}
