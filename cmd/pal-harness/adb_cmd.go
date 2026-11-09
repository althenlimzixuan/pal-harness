package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/althenlimzixuan/pal-harness/adb"
)

// lookPath finds a binary; adbRunner gives the adb runner. Tests replace them.
var (
	lookPath  = exec.LookPath
	adbRunner = execAdb
)

// execAdb runs the adb on PATH.
func execAdb() (adb.Runner, error) {
	bin, err := lookPath("adb")
	if err != nil {
		return nil, errors.New("adb is not on PATH: install Android platform-tools or add it to PATH")
	}
	return func(args ...string) (string, error) {
		out, err := exec.Command(bin, args...).CombinedOutput()
		return string(out), err
	}, nil
}

// runAdb reconnects every device in <dir>/doc/devices.yaml and sets its
// reverse ports: 0 every device ready, 1 one not ready, 2 usage, file or adb missing.
func runAdb(dir string, stdout, stderr io.Writer) int {
	path := filepath.Join(dir, "doc", "devices.yaml")
	b, err := os.ReadFile(path)
	if err != nil {
		return fail(stderr, "cannot read %s: %v", path, err)
	}
	cfg, err := adb.ParseDevices(b)
	if err != nil {
		return fail(stderr, "%v", err)
	}
	run, err := adbRunner()
	if err != nil {
		return fail(stderr, "%v", err)
	}
	ports := make([]string, len(cfg.Ports))
	for i, p := range cfg.Ports {
		ports[i] = strconv.Itoa(p)
	}
	code := 0
	var ask []string
	for _, r := range adb.Up(cfg, run) {
		label := fmt.Sprintf("%s (%s)", r.Device.Name, r.Device.Role)
		if !r.Ready {
			fmt.Fprintf(stdout, "%s: not ready: %s\n", label, r.Problem)
			if r.Device.Wireless != "" {
				ask = append(ask, adb.PairingMessage(r.Device.Name))
			}
			code = 1
			continue
		}
		line := label + ": ready as " + r.ID
		if len(ports) > 0 {
			line += ", reverse " + strings.Join(ports, " ")
		}
		fmt.Fprintln(stdout, line)
		if r.NewAddress != "" {
			fmt.Fprintf(stdout, "%s: new address %s; save it in doc/devices.yaml\n", r.Device.Name, r.NewAddress)
		}
	}
	for _, a := range ask {
		fmt.Fprintln(stdout, "ask the owner: "+a)
	}
	return code
}
