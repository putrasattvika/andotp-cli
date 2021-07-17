package main

import (
	"io/ioutil"
	"log"

	"github.com/grijul/go-andotp/andotp"
	"github.com/pkg/errors"

	"github.com/putrasattvika/andotp-cli/cmd"
)

func main() {
	cmd.Execute()

	// if err := andotpTest(); err != nil {
	// 	log.Fatal(err)
	// }
}

func andotpTest() error {
	filepath := "/home/isattvika/workspace/personal/andotp-cli/.ignored/otp_accounts_2021-07-17_10-01-07.json.aes"

	backupBytesEnc, err := ioutil.ReadFile(filepath)
	if err != nil {
		return errors.Wrap(err, "unable to read andotp backup file")
	}

	backupBytes, err := andotp.Decrypt(backupBytesEnc, "")
	if err != nil {
		return errors.Wrap(err, "unable to decrypt andotp backup file")
	}

	log.Printf("Backup content:\n%s", backupBytes)

	return nil
}
