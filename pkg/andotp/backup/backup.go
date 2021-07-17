package backup

import (
	"encoding/json"

	memguardcore "github.com/awnumar/memguard/core"
	"github.com/grijul/go-andotp/andotp"
	"github.com/pkg/errors"

	"github.com/putrasattvika/andotp-cli/pkg/andotp/otp"
)

// Backup is a struct for an andOTP backup
type Backup struct {
	OTPKeys []*otp.OTPKey

	encrypted []byte
}

func NewBackup(content []byte) (*Backup, error) {
	backup := &Backup{}

	if json.Valid(content) {
		if err := backup.parsePlaintext(content); err != nil {
			return nil, err
		}

		memguardcore.Wipe(content)
	} else {
		backup.encrypted = content
	}

	return backup, nil
}

// IsEncrypted returns true if the backup is currently encrypted
func (b *Backup) IsEncrypted() bool {
	return b.encrypted != nil
}

// Decrypt decrypts the backup
func (b *Backup) Decrypt(password string) error {
	if !b.IsEncrypted() {
		return nil
	}

	plaintext, err := andotp.Decrypt(b.encrypted, password)
	if err != nil {
		return errors.Wrap(err, "unable to decrypt andOTP backup file")
	}

	if err := b.parsePlaintext(plaintext); err != nil {
		return errors.Wrap(err, "error parsing decrypted andOTP backup")
	}

	memguardcore.Wipe(b.encrypted)
	memguardcore.Wipe(plaintext)

	b.encrypted = nil

	return nil
}

func (b *Backup) parsePlaintext(backupContents []byte) error {
	otpKeys, err := otp.OTPKeysFromJSON(backupContents)
	if err != nil {
		return errors.Wrap(err, "error parsing andOTP plaintext JSON backup")
	}

	b.OTPKeys = otpKeys

	return nil
}
