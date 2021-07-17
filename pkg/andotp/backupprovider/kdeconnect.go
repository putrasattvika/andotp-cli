package backupprovider

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/pkg/errors"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"

	"github.com/putrasattvika/andotp-cli/pkg/kdeconnect"
)

var (
	kdeconnect_ssh_username = "kdeconnect"
	kdeconnect_ssh_key      = fmt.Sprintf("%s/.config/kdeconnect/privateKey.pem", os.Getenv("HOME"))
)

// KDEConnect provides andOTP backup from a file inside a KDE Connect device.
// Implements BackupProvider.
type KDEConnect struct {
	deviceHost string
	devicePort string
	filepath   string
}

// NewKDEConnectFromDeviceHostPort creates a new KDEConnect backup provider from
// the device's name or ID. The device's SFTP host & port will be discovered
// automatically.
//
// For the device to expose the SFTP port, we'll need to manually click
// "Browse this device" on the KDE Connect desktop program. The device's IP can
// be obtained from `mount | grep kdeconnect | grep <deviceID>`, while its IP can be obtained from
// `ss -plnaut | grep "<device-ip>" | grep "ssh"`.
func NewKDEConnect(deviceNameOrID string, filepath string) (*KDEConnect, error) {
	devices, err := kdeconnect.ListAvailableDevices()
	if err != nil {
		return nil, errors.Wrap(err, "error listing KDE Connect devices")
	}

	log.Printf("Found %d KDE Connect devices", len(devices))

	var matchingDevice *kdeconnect.Device = nil

	for _, d := range devices {
		if d.Name != deviceNameOrID && d.ID != deviceNameOrID {
			continue
		}

		matchingDevice = d
		break
	}

	if matchingDevice == nil {
		return nil, fmt.Errorf("KDE Connect device with name/id '%s' not found", deviceNameOrID)
	}

	if matchingDevice.SFTPHost == "" || matchingDevice.SFTPPort == "" {
		return nil, fmt.Errorf(
			"KDE Connect device with name/id '%s' does not expose its SFTP port. "+
				"Please manually click the 'Browse this device' button on the KDE Connect "+
				"desktop program or system tray icon.",
			deviceNameOrID,
		)
	}

	log.Printf(
		"Using KDE Connect device '%s' (%s) at (%s:%s)",
		matchingDevice.Name, matchingDevice.ID,
		matchingDevice.SFTPHost, matchingDevice.SFTPPort,
	)

	return NewKDEConnectFromDeviceHostPort(matchingDevice.SFTPHost, matchingDevice.SFTPPort, filepath)
}

// NewKDEConnectFromDeviceHostPort creates a new KDEConnect backup provider from
// the device's SFTP host & port.
func NewKDEConnectFromDeviceHostPort(host string, port string, filepath string) (*KDEConnect, error) {
	return &KDEConnect{
		deviceHost: host,
		devicePort: port,
		filepath:   filepath,
	}, nil
}

// FetchBackup returns the content of the backup file according to the filepath
func (p *KDEConnect) FetchBackup() ([]byte, error) {
	// Read & parse private key file
	sshKeyBytes, err := ioutil.ReadFile(kdeconnect_ssh_key)
	if err != nil {
		return nil, errors.Wrap(err, "error reading KDE Connect SSH private key")
	}

	sshKeySigner, err := ssh.ParsePrivateKey(sshKeyBytes)
	if err != nil {
		return nil, errors.Wrap(err, "error creating signer from KDE Connect SSH private key")
	}

	// Connect via SSH
	sshConfig := &ssh.ClientConfig{
		User:            kdeconnect_ssh_username,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(sshKeySigner)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	sshClient, err := ssh.Dial("tcp", fmt.Sprintf("%s:%s", p.deviceHost, p.devicePort), sshConfig)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"error connecting to KDE Connect device at %s:%s",
			p.deviceHost, p.devicePort,
		)
	}
	defer sshClient.Close()

	// Create SFTP client
	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		return nil, errors.Wrap(err, "error creating SFTP client for KDE Connect device")
	}

	// Read the file
	backupFile, err := sftpClient.Open(p.filepath)
	if err != nil {
		return nil, errors.Wrap(err, "error opening backup file from KDE Connect device")
	}

	backupFileContents, err := ioutil.ReadAll(backupFile)
	if err != nil {
		return nil, errors.Wrap(err, "error reading backup file from KDE Connect device")
	}

	log.Print("Fetched andOTP backup file from KDE Connect device")

	return backupFileContents, nil
}
