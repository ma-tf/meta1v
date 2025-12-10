/*
Package cmd implements the command line interface for meta1v.

Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/ma-tf/meta1v/internal/cli/customfunctions"
	"github.com/ma-tf/meta1v/internal/cli/exif"
	"github.com/ma-tf/meta1v/internal/cli/focusingpoints"
	"github.com/ma-tf/meta1v/internal/cli/frame"
	"github.com/ma-tf/meta1v/internal/cli/roll"
	"github.com/ma-tf/meta1v/internal/cli/thumbnail"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

//nolint:gochecknoglobals // cobra boilerplate
var (
	cfgFile  string
	logger   *slog.Logger
	logLevel = new(slog.LevelVar)
	rootCmd  = &cobra.Command{
		Use:   "meta1v",
		Short: "Provides a way to interact with Canon's EFD files.",
		Long: `meta1v is a command line tool to interact with Canon's EFD files.

You can print out information to stdout about the film roll, including focus
points, custom functions, roll information, thumbnail previews, and more.`,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			err := initialiseConfig(cmd)
			if err != nil {
				return fmt.Errorf("failed to initialise configuration: %w", err)
			}

			cfgLogLevel := viper.GetString("log.level")
			level := slog.LevelInfo
			switch strings.ToLower(cfgLogLevel) {
			case "debug":
				level = slog.LevelDebug
			case "warn", "warning":
				level = slog.LevelWarn
			case "error":
				level = slog.LevelError
			}

			logLevel.Set(level)

			//nolint:sloglint // global logger is fine here
			logger.DebugContext(
				cmd.Context(),
				"Configuration initialised. Using config file:",
				slog.String("cfgFile", viper.ConfigFileUsed()),
			)

			return nil
		},
	}
)

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
	//nolint:exhaustruct // slog boilerplate
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: logLevel,
	})
	logger = slog.New(handler)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().
		StringVar(&cfgFile, "config", "", "config file (default is $HOME/.meta1v/config)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	rootCmd.AddCommand(exif.NewCommand(logger))
	rootCmd.AddCommand(roll.NewCommand(logger))
	rootCmd.AddCommand(customfunctions.NewCommand(logger))
	rootCmd.AddCommand(focusingpoints.NewCommand(logger))
	rootCmd.AddCommand(frame.NewCommand(logger))
	rootCmd.AddCommand(thumbnail.NewCommand(logger))
}

func initialiseConfig(cmd *cobra.Command) error {
	viper.SetEnvPrefix("META1V")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "*", "-", "*"))
	viper.AutomaticEnv()

	if err := viper.BindEnv("log.level", "META1V_LOG_LEVEL"); err != nil {
		return fmt.Errorf("failed to bind env variable: %w", err)
	}

	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Search for a config file in default locations.
		home, err := os.UserHomeDir()
		// Only panic if we can't get the home directory.
		cobra.CheckErr(err)

		// Search config in home directory with name "config" (without extension).
		viper.AddConfigPath(".")
		viper.AddConfigPath(home + "/.meta1v")
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			return fmt.Errorf("failed to initialise config: %w", err)
		}
	}

	err := viper.BindPFlags(cmd.Flags())
	if err != nil {
		return fmt.Errorf("failed to bind config flags: %w", err)
	}

	return nil
}
