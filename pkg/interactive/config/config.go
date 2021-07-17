package config

import (
	"net/url"

	"github.com/pkg/errors"
	cmdconfig "github.com/putrasattvika/andotp-cli/cmd/config"
)

// Configuration used to start the interactive CLI
type Config struct {
	// URI to an andOTP encrypted backup file
	//
	// Supports two sources:
	//   - local file (e.g. file:///home/myuser/otp_accounts.json.aes)
	//   - KDE connect exposed device filesystem (e.g. kdeconnect://_/device-name-or-id/path/to/otp_accounts.json.aes)
	BackupFileURI *url.URL
}

func ParseCmdConfig(cmdConfig *cmdconfig.Config) (*Config, error) {
	// --backup-file-uri
	if cmdConfig.BackupFileURI == "" {
		return nil, errors.New("--backup-file-uri cannot be empty")
	}

	backupFileURI, err := url.Parse(cmdConfig.BackupFileURI)
	if err != nil {
		return nil, errors.Wrap(err, "invalid --backup-file-uri")
	}

	return &Config{
		BackupFileURI: backupFileURI,
	}, nil
}
