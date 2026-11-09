// Package adb brings a mobile project's test phones back after a gap: it
// reconnects each device listed in doc/devices.yaml (at its saved wireless
// address, else at the address mDNS finds now) and sets the reverse ports.
// It never disconnects a device or restarts the adb server.
package adb

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"strings"

	"gopkg.in/yaml.v3"
)

// Device is one phone or emulator in doc/devices.yaml.
type Device struct {
	Name     string `yaml:"name"`
	Role     string `yaml:"role"`     // primary or second-user
	Serial   string `yaml:"serial"`   // adb's USB serial; also inside the wireless mDNS name
	Wireless string `yaml:"wireless"` // last known ip:port for `adb connect`
}

// Config is doc/devices.yaml.
type Config struct {
	Ports   []int    `yaml:"ports"`
	Devices []Device `yaml:"devices"`
}

// ParseDevices reads and validates doc/devices.yaml.
func ParseDevices(b []byte) (*Config, error) {
	var c Config
	dec := yaml.NewDecoder(bytes.NewReader(b))
	dec.KnownFields(true)
	if err := dec.Decode(&c); err != nil {
		return nil, fmt.Errorf("doc/devices.yaml: %v", err)
	}
	if len(c.Devices) == 0 {
		return nil, errors.New("doc/devices.yaml lists no devices")
	}
	for _, p := range c.Ports {
		if p < 1 || p > 65535 {
			return nil, fmt.Errorf("doc/devices.yaml: port %d is not 1 to 65535", p)
		}
	}
	seen := map[string]bool{}
	for i := range c.Devices {
		d := &c.Devices[i]
		switch {
		case d.Name == "":
			return nil, fmt.Errorf("doc/devices.yaml: device %d has no name", i+1)
		case seen[d.Name]:
			return nil, fmt.Errorf("doc/devices.yaml: two devices are named %q", d.Name)
		case d.Serial == "" && d.Wireless == "":
			return nil, fmt.Errorf("doc/devices.yaml: %s needs a serial or a wireless address", d.Name)
		}
		seen[d.Name] = true
		if d.Role == "" {
			d.Role = "primary"
		}
		if d.Role != "primary" && d.Role != "second-user" {
			return nil, fmt.Errorf("doc/devices.yaml: %s has role %q; want primary or second-user", d.Name, d.Role)
		}
		if d.Wireless != "" {
			if _, _, err := net.SplitHostPort(d.Wireless); err != nil {
				return nil, fmt.Errorf("doc/devices.yaml: %s wireless %q is not ip:port", d.Name, d.Wireless)
			}
		}
	}
	return &c, nil
}

// Runner runs adb with args and returns its combined output.
type Runner func(args ...string) (string, error)

// Result is one device's outcome.
type Result struct {
	Device     Device
	ID         string // the adb id used for -s
	Ready      bool
	NewAddress string // set when the wireless address differs from the saved one
	Problem    string
}

// Up connects every device and sets its reverse ports.
func Up(c *Config, run Runner) []Result {
	res := make([]Result, 0, len(c.Devices))
	for _, d := range c.Devices {
		r := up(d, c.Ports, run)
		if r.Ready && d.Wireless != "" && isAddr(r.ID) && r.ID != d.Wireless {
			r.NewAddress = r.ID
		}
		res = append(res, r)
	}
	return res
}

func up(d Device, ports []int, run Runner) Result {
	r := Result{Device: d}
	id, err := listed(d, run)
	if err != nil {
		r.Problem = "adb devices failed: " + err.Error()
		return r
	}
	if id == "" && d.Wireless != "" {
		if connect(d.Wireless, run) {
			id = d.Wireless
		} else if addr := lookup(d, run); addr != "" && connect(addr, run) {
			id = addr
		}
	}
	if id == "" {
		if d.Wireless == "" {
			r.Problem = "not connected: plug it in by USB (serial " + d.Serial + ")"
		} else {
			r.Problem = "not connected at " + d.Wireless + " or any address mDNS found"
		}
		return r
	}
	r.ID = id
	for _, p := range ports {
		port := fmt.Sprintf("tcp:%d", p)
		if out, err := run("-s", id, "reverse", port, port); err != nil {
			r.Problem = fmt.Sprintf("reverse %s failed: %v %s", port, err, strings.TrimSpace(out))
			return r
		}
	}
	r.Ready = true
	return r
}

// listed returns the adb id under which d is already attached and ready, or "".
func listed(d Device, run Runner) (string, error) {
	out, err := run("devices")
	if err != nil {
		return "", err
	}
	ip := host(d.Wireless)
	for line := range strings.SplitSeq(out, "\n") {
		f := strings.Fields(line)
		if len(f) != 2 || f[1] != "device" {
			continue
		}
		id := f[0]
		switch {
		case d.Serial != "" && (id == d.Serial || strings.HasPrefix(id, "adb-"+d.Serial+"-")):
			return id, nil
		case d.Wireless != "" && (id == d.Wireless || (isAddr(id) && host(id) == ip)):
			return id, nil
		}
	}
	return "", nil
}

// connect runs `adb connect` and reads success from its text (adb exits 0 on failure too).
func connect(addr string, run Runner) bool {
	out, err := run("connect", addr)
	out = strings.ToLower(strings.TrimSpace(out))
	return err == nil && strings.Contains(out, "connected to") &&
		!strings.Contains(out, "failed") && !strings.Contains(out, "cannot")
}

// lookup finds d's current wireless connect address in `adb mdns services`,
// matched by its serial, else by the saved address's IP.
func lookup(d Device, run Runner) string {
	out, err := run("mdns", "services")
	if err != nil {
		return ""
	}
	byIP := ""
	for line := range strings.SplitSeq(out, "\n") {
		f := strings.Fields(line)
		if len(f) != 3 || f[1] != "_adb-tls-connect._tcp" || !isAddr(f[2]) {
			continue
		}
		if d.Serial != "" && strings.HasPrefix(f[0], "adb-"+d.Serial+"-") {
			return f[2]
		}
		if byIP == "" && host(f[2]) == host(d.Wireless) {
			byIP = f[2]
		}
	}
	return byIP
}

func isAddr(s string) bool {
	_, _, err := net.SplitHostPort(s)
	return err == nil
}

func host(addr string) string {
	h, _, err := net.SplitHostPort(addr)
	if err != nil {
		return ""
	}
	return h
}

// PairingMessage is what the session asks the owner when a phone needs pairing again.
func PairingMessage(name string) string {
	return "on " + name + ", open Settings → Developer options → Wireless debugging → Pair device with pairing code, and send me the code and the port"
}
