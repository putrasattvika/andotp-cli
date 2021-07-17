package kdeconnect

import (
	"bytes"
	"log"
	"os/exec"
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

var (
	cliListAvailableRegex = regexp.MustCompile(`^- (.*): ([a-z0-9]+).*$`)
	mountSFTPHostRegex    = regexp.MustCompile(`^kdeconnect@(.+):.*$`)
	ssSFTPPortRegex       = regexp.MustCompile(`[^\s]+\s+[^\s]+\s+[^\s]+\s+[^\s]+\s+(?:[0-9.]+:([0-9]+))\s+[^\s]+`)
)

type Device struct {
	// Name and ID of the device
	Name string
	ID   string

	// Host & port of the device's SFTP server. Will be empty strings if the
	// device's filesystem is not yet mounted by KDE Connect via the
	// "Browse this device" button.
	SFTPHost string
	SFTPPort string
}

// ListAvailableDevices returns a list of KDE Connect devices connected to
// the host
func ListAvailableDevices() ([]*Device, error) {
	availableDevices, err := listAvailableDevices()
	if err != nil {
		return nil, err
	}

	for _, d := range availableDevices {
		if err := populateSFTPHostPort(d); err != nil {
			log.Printf("error getting KDE connect device host/port for device '%s'", d.Name)
		}
	}

	return availableDevices, nil
}

// listAvailableDevices returns a list of KDE Connect devices connected to
// the host with `kdeconnect-cli -a`
func listAvailableDevices() ([]*Device, error) {
	var out bytes.Buffer

	cmd := exec.Command("kdeconnect-cli", "-a")
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return nil, errors.Wrap(err, "error listing available KDE Connect devices")
	}

	// Parse the output, example:
	//
	// - S6 Lite: 82c27bf0c8d7fbc5 (paired and reachable)
	// - Asus Max Pro M1: cc06e6c222be2ff6 (paired and reachable)
	// 2 devices found
	devices := []*Device{}

	for _, line := range strings.Split(out.String(), "\n") {
		if !strings.HasPrefix(line, "- ") {
			continue
		}

		// Parse format like this:
		// ^- Asus Max Pro M1: cc06e6c222be2ff6 (paired and reachable)$
		matches := cliListAvailableRegex.FindAllStringSubmatch(line, -1)

		if len(matches) != 1 {
			continue
		}

		if len(matches[0]) != 3 {
			continue
		}

		devices = append(
			devices,
			&Device{Name: matches[0][1], ID: matches[0][2]},
		)
	}

	return devices, nil
}

// populateSFTPHostPort populates the device's SFTP host/port field accordingly.
// The device's host is obtained from `mount | grep <device-id>`, while the port
// is obtained from `ss -plnaut | grep <device-ip>`.
func populateSFTPHostPort(device *Device) error {
	// Get the device's host
	var mountOut bytes.Buffer

	mountCmd := exec.Command("mount")
	mountCmd.Stdout = &mountOut

	if err := mountCmd.Run(); err != nil {
		return errors.Wrap(err, "error listing KDE Connect mounted volumes")
	}

	// Get device IP from expected output format:
	// kdeconnect@192.168.0.188:/ on /run/user/1000/82c27bf0c8d7fbc5 type fuse.sshfs (rw,nosuid,nodev,relatime,user_id=1000,group_id=1000)
	for _, line := range strings.Split(mountOut.String(), "\n") {
		if !strings.Contains(line, device.ID) || !strings.Contains(line, "kdeconnect") {
			continue
		}

		matches := mountSFTPHostRegex.FindAllStringSubmatch(line, -1)

		if len(matches) != 1 {
			continue
		}

		if len(matches[0]) != 2 {
			continue
		}

		device.SFTPHost = matches[0][1]
	}

	if device.SFTPHost == "" {
		return nil
	}

	// Get device SFTP port
	var ssOut bytes.Buffer

	ssCmd := exec.Command("ss", "-plnta")
	ssCmd.Stdout = &ssOut

	if err := ssCmd.Run(); err != nil {
		return errors.Wrap(err, "error listing KDE Connect established connections")
	}

	// Get device port from expected output format:
	// ESTAB     0      0                 192.168.0.138:41718            192.168.0.188:1739                    users:(("ssh",pid=361762,fd=3))
	for _, line := range strings.Split(ssOut.String(), "\n") {
		if !strings.Contains(line, device.SFTPHost) || !strings.Contains(line, "ssh") {
			continue
		}

		matches := ssSFTPPortRegex.FindAllStringSubmatch(line, -1)

		if len(matches) != 1 {
			continue
		}

		if len(matches[0]) != 2 {
			continue
		}

		device.SFTPPort = matches[0][1]
	}

	return nil
}
