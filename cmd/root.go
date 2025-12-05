/*
Package cmd implements the command line interface for meta1v.

Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log/slog"
	"os"

	"github.com/ma-tf/meta1v/internal/cli/customfunctions"
	"github.com/ma-tf/meta1v/internal/cli/dng"
	"github.com/ma-tf/meta1v/internal/cli/focusingpoints"
	"github.com/ma-tf/meta1v/internal/cli/frame"
	"github.com/ma-tf/meta1v/internal/cli/roll"
	"github.com/ma-tf/meta1v/internal/cli/thumbnail"
	"github.com/spf13/cobra"
)

//nolint:gochecknoglobals // cobra boilerplate
var rootCmd = &cobra.Command{
	Use:   "meta1v",
	Short: "Provides a way to interact with Canon's EFD files.",
	Long: `meta1v is a command line tool to interact with Canon's EFD files.

You can print out information to stdout about the film roll, including focus
points, custom functions, roll information, thumbnail previews, and more.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

//nolint:gochecknoinits // cobra boilerplate
func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	//nolint:exhaustruct // slog boilerplate
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	logger := slog.New(handler)

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.meta1v.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	rootCmd.AddCommand(dng.NewCommand(logger))
	rootCmd.AddCommand(roll.NewCommand())
	rootCmd.AddCommand(customfunctions.NewCommand())
	rootCmd.AddCommand(focusingpoints.NewCommand())
	rootCmd.AddCommand(frame.NewCommand())
	rootCmd.AddCommand(thumbnail.NewCommand())
}
