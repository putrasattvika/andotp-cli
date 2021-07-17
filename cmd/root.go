package cmd

import (
	"encoding/csv"
	"log"
	"os"
	"strings"

	"github.com/awnumar/memguard"
	"github.com/spf13/cobra"

	"github.com/putrasattvika/andotp-cli/cmd/config"
	"github.com/putrasattvika/andotp-cli/pkg/interactive"
	interactiveconfig "github.com/putrasattvika/andotp-cli/pkg/interactive/config"
)

// DebugArgsEnv is the name for environment variable that contains args for
// debugging purposes
const DebugArgsEnv = "DEBUG_ARGS"

type rootCmd struct {
	config *config.Config
}

// newRootCmd creates a new "root" command group
func newRootCmd() *cobra.Command {
	rootCmdObj := &rootCmd{
		config: &config.Config{},
	}

	// The root command
	cmd := &cobra.Command{
		Use:   "andotp-cli",
		Short: "Interactive CLI TOTP generator for andOTP backup file",
		Long:  "Interactive CLI TOTP generator for andOTP backup file",

		Run: rootCmdObj.entrypoint,
	}

	// Positional
	cmd.PersistentFlags().StringVarP(
		&rootCmdObj.config.BackupFileURI,
		"backup-file-uri", "b",
		"",
		"URI to an andOTP backup file. Supports two sources: "+
			"local file (e.g. file:///home/myuser/otp_accounts.json.aes) "+
			"and KDE connect exposed device filesystem "+
			"(e.g. kdeconnect://_/device-name-or-id/path/to/otp_accounts.json.aes)",
	)

	return cmd
}

// Entrypoint for the "serve" command
func (c *rootCmd) entrypoint(cmd *cobra.Command, args []string) {
	// Start an interrupt handler that will clean up memory before exiting and
	// purge the session when returning from the main function of your program
	memguard.CatchInterrupt()
	defer memguard.Purge()

	interactiveConfig, err := interactiveconfig.ParseCmdConfig(c.config)
	if err != nil {
		log.Fatalf("error parsing/validating arguments: %v", err)
	}

	interactive_, err := interactive.NewInteractive(interactiveConfig)
	if err != nil {
		log.Fatalf("error creating interactive session: %v", err)
	}

	if err := interactive_.Start(); err != nil {
		log.Fatalf("error starting interactive session: %v", err)
	}
}

// Execute is called by main.main()
func Execute() {
	rootCmd := newRootCmd()

	// On debug mode, use arguments from environment variable DEBUG_ARGS. This is
	// useful when debugging with an IDE where we can't dynamically specify args
	// for the program.
	if os.Getenv(DebugArgsEnv) != "" {
		// Use CSV reader to parse double-quoted words properly
		// Ref: https://stackoverflow.com/a/47489825

		r := csv.NewReader(strings.NewReader(os.Getenv(DebugArgsEnv)))
		r.Comma = ' '

		record, err := r.Read()
		if err != nil {
			log.Fatalf("Error parsing DEBUG_ARGS env: %+v", err)
		}

		rootCmd.SetArgs(record)
	}

	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Error executing command: %v", err)
	}
}
