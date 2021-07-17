package backupprovider

import (
	"fmt"
	"net/url"
	"strings"
)

var (
	// Mapping between backup file URI scheme and its provider constructor
	AvailableProviders = map[string]BackupProviderConstructor{
		"":     ConstructLocalFileProvider,
		"file": ConstructLocalFileProvider,

		"kdeconnect": ConstructKDEConnectProvider,
	}
)

// ConstructBackupProvider constructs the appropriate backup provider according
// to the backup file URI scheme
func ConstructBackupProvider(uri *url.URL) (BackupProvider, error) {
	for scheme, constructor := range AvailableProviders {
		if scheme == uri.Scheme {
			return constructor(uri)
		}
	}

	return nil, fmt.Errorf("unsupported backup file URI scheme '%s'", uri.Scheme)
}

// ConstructLocalFileProvider constructs a local file backup provider
func ConstructLocalFileProvider(uri *url.URL) (BackupProvider, error) {
	return NewLocalFile(uri.Path)
}

// ConstructKDEConnectProvider constructs a KDE Connect backup provider
func ConstructKDEConnectProvider(uri *url.URL) (BackupProvider, error) {
	// Manually-defined IP:port of the KDE Connect device
	if len(uri.Port()) > 0 {
		return NewKDEConnectFromDeviceHostPort(
			uri.Hostname(),
			uri.Port(),
			uri.Path,
		)
	}

	// Discover the device by ourselves
	uriPathSplit := strings.SplitN(uri.Path, "/", 3)
	if len(uriPathSplit) != 3 {
		return nil, fmt.Errorf("invalid KDE Connect path: %s", uri.Path)
	}

	return NewKDEConnect(uriPathSplit[1], "/"+uriPathSplit[2])
}
