package backupprovider

import (
	"io/ioutil"

	"github.com/pkg/errors"
)

// LocalFile provides andOTP backup from a local file.
// Implements BackupProvider.
type LocalFile struct {
	filepath string
}

// NewLocalFile creates a new LocalFile backup provider
func NewLocalFile(filepath string) (*LocalFile, error) {
	return &LocalFile{filepath: filepath}, nil
}

// FetchBackup returns the content of the backup file according to the filepath
func (p *LocalFile) FetchBackup() ([]byte, error) {
	backupBytes, err := ioutil.ReadFile(p.filepath)
	if err != nil {
		return nil, errors.Wrap(err, "unable to read backup file from local filesystem")
	}

	return backupBytes, nil
}
