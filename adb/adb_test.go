package adb

import (
	"errors"
	"strings"
	"testing"
)

const yamlOK = `ports: [8081, 8080]
devices:
  - name: Samsung A26
    role: primary
    serial: R5CX123
    wireless: 192.168.1.23:41234
  - name: Pixel 6
    role: second-user
    wireless: 192.168.1.40:39000
`

func TestParseDevices(t *testing.T) {
	c, err := ParseDevices([]byte(yamlOK))
	if err != nil {
		t.Fatal(err)
	}
	if len(c.Ports) != 2 || c.Ports[0] != 8081 || len(c.Devices) != 2 {
		t.Fatalf("%+v", c)
	}
	if d := c.Devices[1]; d.Name != "Pixel 6" || d.Role != "second-user" || d.Wireless != "192.168.1.40:39000" {
		t.Fatalf("%+v", d)
	}
	// role defaults to primary
	c, err = ParseDevices([]byte("devices:\n  - name: A\n    serial: X\n"))
	if err != nil || c.Devices[0].Role != "primary" {
		t.Fatalf("%+v %v", c, err)
	}
}

func TestParseDevicesRejects(t *testing.T) {
	for name, y := range map[string]string{
		"not yaml":        "devices: [",
		"no devices":      "ports: [8081]\n",
		"no name":         "devices:\n  - serial: X\n",
		"no address":      "devices:\n  - name: A\n",
		"bad role":        "devices:\n  - name: A\n    serial: X\n    role: boss\n",
		"bad port":        "ports: [0]\ndevices:\n  - name: A\n    serial: X\n",
		"bad wireless":    "devices:\n  - name: A\n    wireless: 192.168.1.2\n",
		"unknown field":   "devices:\n  - name: A\n    serial: X\n    pairing: 123456\n",
		"duplicate names": "devices:\n  - name: A\n    serial: X\n  - name: A\n    serial: Y\n",
	} {
		if _, err := ParseDevices([]byte(y)); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
}

// fakeADB answers adb commands from a table and records every call.
type fakeADB struct {
	devices string            // `adb devices` output after the header
	mdns    string            // `adb mdns services` output after the header
	connect map[string]string // address → output
	reverse error             // error for every reverse call
	calls   []string
}

func (f *fakeADB) run(args ...string) (string, error) {
	call := strings.Join(args, " ")
	f.calls = append(f.calls, call)
	switch {
	case call == "devices":
		return "List of devices attached\n" + f.devices, nil
	case call == "mdns services":
		return "List of discovered mdns services\n" + f.mdns, nil
	case args[0] == "connect":
		out, ok := f.connect[args[1]]
		if !ok {
			return "failed to connect to '" + args[1] + "': Connection refused\n", nil
		}
		if strings.HasPrefix(out, "connected to") {
			f.devices += args[1] + "\tdevice\n"
		}
		return out, nil
	case len(args) > 2 && args[2] == "reverse":
		return "", f.reverse
	}
	return "", errors.New("unexpected adb call: " + call)
}

func (f *fakeADB) called(prefix string) []string {
	var got []string
	for _, c := range f.calls {
		if strings.HasPrefix(c, prefix) {
			got = append(got, c)
		}
	}
	return got
}

func cfg(t *testing.T) *Config {
	t.Helper()
	c, err := ParseDevices([]byte(yamlOK))
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestUpAlreadyConnected(t *testing.T) {
	f := &fakeADB{devices: "R5CX123\tdevice\n192.168.1.40:39000\tdevice\n"}
	res := Up(cfg(t), f.run)
	if !res[0].Ready || res[0].ID != "R5CX123" || !res[1].Ready || res[1].ID != "192.168.1.40:39000" {
		t.Fatalf("%+v", res)
	}
	if got := f.called("connect"); len(got) != 0 {
		t.Errorf("connected again: %v", got)
	}
	want := []string{"-s R5CX123 reverse tcp:8081 tcp:8081", "-s R5CX123 reverse tcp:8080 tcp:8080",
		"-s 192.168.1.40:39000 reverse tcp:8081 tcp:8081", "-s 192.168.1.40:39000 reverse tcp:8080 tcp:8080"}
	if got := f.called("-s"); strings.Join(got, "|") != strings.Join(want, "|") {
		t.Errorf("reverse calls %v, want %v", got, want)
	}
}

func TestUpListedOnAnotherPortOrByMDNSName(t *testing.T) {
	// The A26 shows under its mDNS name, the Pixel on a new port: both count as connected.
	f := &fakeADB{devices: "adb-R5CX123-AbCd._adb-tls-connect._tcp\tdevice\n192.168.1.40:41111\tdevice\n"}
	res := Up(cfg(t), f.run)
	if !res[0].Ready || res[0].ID != "adb-R5CX123-AbCd._adb-tls-connect._tcp" {
		t.Fatalf("%+v", res[0])
	}
	if !res[1].Ready || res[1].ID != "192.168.1.40:41111" || res[1].NewAddress != "192.168.1.40:41111" {
		t.Fatalf("%+v", res[1])
	}
}

func TestUpIgnoresOfflineAndUnauthorized(t *testing.T) {
	f := &fakeADB{devices: "R5CX123\toffline\n192.168.1.40:39000\tunauthorized\n",
		connect: map[string]string{"192.168.1.23:41234": "connected to 192.168.1.23:41234\n"}}
	res := Up(cfg(t), f.run)
	if !res[0].Ready || res[0].ID != "192.168.1.23:41234" {
		t.Fatalf("%+v", res[0])
	}
	if res[1].Ready {
		t.Fatalf("%+v", res[1])
	}
}

func TestUpReconnectsAtSavedAddress(t *testing.T) {
	f := &fakeADB{connect: map[string]string{
		"192.168.1.23:41234": "connected to 192.168.1.23:41234\n",
		"192.168.1.40:39000": "already connected to 192.168.1.40:39000\n",
	}}
	res := Up(cfg(t), f.run)
	for _, r := range res {
		if !r.Ready || r.NewAddress != "" {
			t.Fatalf("%+v", r)
		}
	}
	if got := f.called("mdns"); len(got) != 0 {
		t.Errorf("looked up mDNS: %v", got)
	}
}

func TestUpFindsNewPortByMDNS(t *testing.T) {
	f := &fakeADB{
		mdns: "adb-R5CX123-AbCd\t_adb-tls-connect._tcp\t192.168.1.23:37000\n" +
			"adb-ZZZ-1\t_adb-tls-pairing._tcp\t192.168.1.40:45000\n" +
			"adb-PIX-9\t_adb-tls-connect._tcp\t192.168.1.40:43210\n",
		connect: map[string]string{
			"192.168.1.23:37000": "connected to 192.168.1.23:37000\n",
			"192.168.1.40:43210": "connected to 192.168.1.40:43210\n",
		},
	}
	res := Up(cfg(t), f.run)
	if !res[0].Ready || res[0].NewAddress != "192.168.1.23:37000" {
		t.Fatalf("by serial: %+v", res[0])
	}
	if !res[1].Ready || res[1].NewAddress != "192.168.1.40:43210" {
		t.Fatalf("by IP: %+v", res[1])
	}
}

func TestUpNotReachable(t *testing.T) {
	f := &fakeADB{}
	res := Up(cfg(t), f.run)
	for _, r := range res {
		if r.Ready || !strings.Contains(r.Problem, "not connected") {
			t.Fatalf("%+v", r)
		}
	}
	if got := f.called("-s"); len(got) != 0 {
		t.Errorf("reversed on a missing device: %v", got)
	}
}

func TestUpUSBOnlyDeviceMissing(t *testing.T) {
	c, _ := ParseDevices([]byte("devices:\n  - name: A\n    serial: X\n"))
	f := &fakeADB{}
	res := Up(c, f.run)
	if res[0].Ready || !strings.Contains(res[0].Problem, "USB") {
		t.Fatalf("%+v", res[0])
	}
}

func TestUpReverseFails(t *testing.T) {
	f := &fakeADB{devices: "R5CX123\tdevice\n192.168.1.40:39000\tdevice\n", reverse: errors.New("error: closed")}
	res := Up(cfg(t), f.run)
	if res[0].Ready || !strings.Contains(res[0].Problem, "reverse tcp:8081") {
		t.Fatalf("%+v", res[0])
	}
}

func TestUpListFails(t *testing.T) {
	run := func(args ...string) (string, error) { return "", errors.New("adb server not running") }
	res := Up(cfg(t), run)
	if res[0].Ready || !strings.Contains(res[0].Problem, "adb server not running") {
		t.Fatalf("%+v", res[0])
	}
}

func TestUpNeverDisconnects(t *testing.T) {
	f := &fakeADB{mdns: "adb-R5CX123-AbCd\t_adb-tls-connect._tcp\t192.168.1.23:37000\n"}
	Up(cfg(t), f.run)
	for _, c := range f.calls {
		if strings.Contains(c, "disconnect") || strings.Contains(c, "kill-server") {
			t.Errorf("ran %q", c)
		}
	}
}

func TestPairingMessage(t *testing.T) {
	want := "on Pixel 6, open Settings → Developer options → Wireless debugging → Pair device with pairing code, and send me the code and the port"
	if got := PairingMessage("Pixel 6"); got != want {
		t.Errorf("%q", got)
	}
}
