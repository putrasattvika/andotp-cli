package config

// Configuration passed from the command line arguments
type Config struct {
	// URI to an andOTP encrypted backup file
	//
	// Supports two sources:
	//   - local file (e.g. file:///home/myuser/otp_accounts.json.aes)
	//   - KDE connect exposed device filesystem (e.g. kdeconnect://_/device-name-or-id/path/to/otp_accounts.json.aes)
	BackupFileURI string
}
