package main

import (
	"errors"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/althenlimzixuan/pal-harness/adb"
)

// fakeAdb replaces the adb finder with a runner answering from the given `adb devices` list.
func fakeAdb(t *testing.T, devices string) *[]string {
	t.Helper()
	old := adbRunner
	var calls []string
	adbRunner = func() (adb.Runner, error) {
		return func(args ...string) (string, error) {
			calls = append(calls, strings.Join(args, " "))
			switch args[0] {
			case "devices":
				return "List of devices attached\n" + devices, nil
			case "connect":
				return "failed to connect to " + args[1], nil
			}
			return "", nil
		}, nil
	}
	t.Cleanup(func() { adbRunner = old })
	return &calls
}

const devicesYAML = `ports: [8081]
devices:
  - name: Samsung A26
    serial: R5CX123
    wireless: 192.168.1.23:41234
  - name: Pixel 6
    role: second-user
    wireless: 192.168.1.40:39000
`

func TestAdbCommandReady(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "doc/devices.yaml", devicesYAML)
	calls := fakeAdb(t, "R5CX123\tdevice\n192.168.1.40:40000\tdevice\n")
	code, out, errs := do(t, "adb", dir)
	if code != 0 {
		t.Fatalf("%d %q %q", code, out, errs)
	}
	for _, want := range []string{
		"Samsung A26 (primary): ready as R5CX123, reverse 8081",
		"Pixel 6 (second-user): ready as 192.168.1.40:40000, reverse 8081",
		"Pixel 6: new address 192.168.1.40:40000; save it in doc/devices.yaml",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output lacks %q:\n%s", want, out)
		}
	}
	if len(*calls) == 0 {
		t.Error("adb not run")
	}
}

func TestAdbCommandNotReady(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "doc/devices.yaml", devicesYAML)
	fakeAdb(t, "R5CX123\tdevice\n")
	code, out, _ := do(t, "adb", dir)
	if code != 1 || !strings.Contains(out, "Pixel 6 (second-user): not ready: not connected") ||
		!strings.Contains(out, adb.PairingMessage("Pixel 6")) {
		t.Fatalf("%d %q", code, out)
	}
}

func TestAdbCommandErrors(t *testing.T) {
	dir := t.TempDir()
	fakeAdb(t, "")
	if code, _, errs := do(t, "adb", dir); code != 2 || !strings.Contains(errs, filepath.Join("doc", "devices.yaml")) {
		t.Errorf("missing file: %d %q", code, errs)
	}
	write(t, dir, "doc/devices.yaml", "devices: []\n")
	if code, _, errs := do(t, "adb", dir); code != 2 || !strings.Contains(errs, "no devices") {
		t.Errorf("empty file: %d %q", code, errs)
	}
	if code, _, errs := do(t, "adb", "a", "b"); code != 2 || !strings.Contains(errs, "usage") {
		t.Errorf("two dirs: %d %q", code, errs)
	}
	write(t, dir, "doc/devices.yaml", devicesYAML)
	adbRunner = func() (adb.Runner, error) { return nil, errors.New("adb is not on PATH") }
	if code, _, errs := do(t, "adb", dir); code != 2 || !strings.Contains(errs, "adb is not on PATH") {
		t.Errorf("no adb: %d %q", code, errs)
	}
}

func TestExecAdbRunner(t *testing.T) {
	old := lookPath
	t.Cleanup(func() { lookPath = old })
	lookPath = func(string) (string, error) { return "", errors.New("not found") }
	if _, err := execAdb(); err == nil || !strings.Contains(err.Error(), "adb is not on PATH") {
		t.Fatalf("err = %v", err)
	}
	// The real runner runs the binary it found: the go tool stands in for adb.
	lookPath = func(string) (string, error) { return goBinary(t), nil }
	run, err := execAdb()
	if err != nil {
		t.Fatal(err)
	}
	if out, err := run("version"); err != nil || !strings.Contains(out, "go version") {
		t.Fatalf("%q %v", out, err)
	}
}

func goBinary(t *testing.T) string {
	t.Helper()
	p, err := exec.LookPath("go")
	if err != nil {
		t.Skip("go is not on PATH")
	}
	return p
}
