// Package cli provides the root command and CLI interface for the meta1v application.
// It serves as the entry point for all CLI commands that interact with Canon EFD files.
package cli

import (
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "meta1v",
		Short: "Provides a way to interact with Canon's EFD files.",
	}
}
