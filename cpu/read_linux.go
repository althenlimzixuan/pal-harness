package cpu

import "os"

// Read returns the machine's processor times from /proc/stat.
func Read() (Times, error) {
	f, err := os.Open("/proc/stat")
	if err != nil {
		return Times{}, err
	}
	defer f.Close()
	return parseProcStat(f)
}
