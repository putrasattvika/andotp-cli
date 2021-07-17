package backupprovider

import "net/url"

// BackupProviderConstructor is the signature of BackupProvider constructor
type BackupProviderConstructor func(uri *url.URL) (BackupProvider, error)

// BackupProvider is an interface for obtaining andOTP backup
type BackupProvider interface {
	// FetchBackup returns the content of a backup file. The returned backup may
	// be encrypted.
	FetchBackup() ([]byte, error)
}
