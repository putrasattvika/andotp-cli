package interactive

import (
	"fmt"
	"log"
	"runtime"
	"strings"
	"syscall"

	"github.com/atotto/clipboard"
	prompt "github.com/c-bata/go-prompt"
	"github.com/pkg/errors"
	"golang.org/x/term"

	andotpbackup "github.com/putrasattvika/andotp-cli/pkg/andotp/backup"
	andotpbackupprovider "github.com/putrasattvika/andotp-cli/pkg/andotp/backupprovider"
	"github.com/putrasattvika/andotp-cli/pkg/andotp/otp"
	"github.com/putrasattvika/andotp-cli/pkg/interactive/config"
)

// Struct for the interactive CLI interface
type Interactive struct {
	config       *config.Config
	andOTPBackup *andotpbackup.Backup
}

// Create a new Interactive
func NewInteractive(config *config.Config) (*Interactive, error) {
	return &Interactive{config: config}, nil
}

// Start an interactive CLI session
func (i *Interactive) Start() error {
	if err := i.loadOTPKeys(); err != nil {
		return errors.Wrap(err, "unable to load OTP keys from andOTP backup file")
	}

	log.Printf(
		"andOTP backup file loaded, %d OTP keys available",
		len(i.andOTPBackup.OTPKeys),
	)

	log.Print("Starting interactive session. Press ctrl+d to exit.")

	return i.startInteractiveSession()
}

func (i *Interactive) loadOTPKeys() error {
	// Get backup file provider
	backupProvider, err := andotpbackupprovider.ConstructBackupProvider(i.config.BackupFileURI)
	if err != nil {
		return errors.Wrap(err, "unable to construct backup file provider")
	}

	// Fetch & parse backup contents
	backupContents, err := backupProvider.FetchBackup()
	if err != nil {
		return errors.Wrap(err, "unable to fetch backup file")
	}

	backup, err := andotpbackup.NewBackup(backupContents)
	if err != nil {
		return errors.Wrap(err, "unable to parse backup file")
	}

	// Decrypt the backup
	if backup.IsEncrypted() {
		fmt.Print("Enter backup password: ")
		passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
		fmt.Print("\n")

		if err != nil {
			return errors.Wrap(err, "error reading backup password from stdin")
		}

		if err := backup.Decrypt(string(passwordBytes)); err != nil {
			return errors.Wrap(err, "unable to decrypt backup")
		}
	}

	// Run GC to remove decryption password and decrypted backup from memory
	runtime.GC()

	i.andOTPBackup = backup

	return nil
}

func (i *Interactive) startInteractiveSession() error {
	// Lookup table for OTP key display name to its OTPKey struct
	otpKeyDisplayNameMap := make(map[string]*otp.OTPKey)

	// All OTP keys for suggestions
	suggestions := []prompt.Suggest{}

	for idx, otpKey := range i.andOTPBackup.OTPKeys {
		displayName := fmt.Sprintf("[%d] %s | %s", idx+1, otpKey.Issuer, otpKey.Label)

		otpKeyDisplayNameMap[displayName] = otpKey

		suggestions = append(
			suggestions,
			prompt.Suggest{Text: displayName},
		)
	}

	// Executor for the shell
	executor := func(in string) {
		in = strings.TrimSpace(in)
		if len(in) == 0 {
			return
		}

		otpKey, otpKeyExists := otpKeyDisplayNameMap[in]
		if !otpKeyExists {
			fmt.Printf("OTP key with name '%s' does not exist\n\n", in)
			return
		}

		token, err := otpKey.GenerateCode()
		if err != nil {
			fmt.Printf("Error during token generation: %v\n\n", err)
			return
		}

		if err := clipboard.WriteAll(token); err != nil {
			fmt.Printf("Token: '%s'\n", token)
			fmt.Printf("Cannot copy token to clipboard, error: %v\n", err)
		} else {
			fmt.Print("Token copied to clipboard\n\n")
		}
	}

	// Completer for the shell
	completer := func(in prompt.Document) []prompt.Suggest {
		// Do not suggest anything if a valid argument is already typed in
		if _, ok := otpKeyDisplayNameMap[strings.TrimSpace(in.Text)]; ok {
			return nil
		}

		return prompt.FilterContains(suggestions, in.GetWordBeforeCursor(), true)
	}

	// Run the interactive shell
	prompt.New(
		executor,
		completer,
		prompt.OptionPrefix(">>> "),
		prompt.OptionTitle("andOTP-cli Interactive Shell"),
	).Run()

	return nil
}
