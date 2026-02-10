// meta1v is a command-line tool for viewing and manipulating metadata for Canon EOS-1V files of the EFD format.
// Copyright (C) 2026  Matt F
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

// Package cmd provides the command-line interface for meta1v.
//
// It implements the root command and configuration management using Cobra and Viper,
// including subcommands for viewing roll data, frames, custom functions, focus points,
// thumbnails, and writing EXIF metadata.
package cmd

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/lmittmann/tint"
	"github.com/ma-tf/meta1v/internal/cli/customfunctions"
	"github.com/ma-tf/meta1v/internal/cli/exif"
	"github.com/ma-tf/meta1v/internal/cli/focusingpoints"
	"github.com/ma-tf/meta1v/internal/cli/frame"
	"github.com/ma-tf/meta1v/internal/cli/roll"
	"github.com/ma-tf/meta1v/internal/cli/thumbnail"
	"github.com/ma-tf/meta1v/internal/container"
	"github.com/ma-tf/meta1v/internal/service/osexec"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Config holds the application configuration loaded from file, environment, or flags.
type Config struct {
	Log struct {
		Level string `mapstructure:"level"`
	} `mapstructure:"log"`
	Strict  bool          `mapstructure:"strict"`
	Timeout time.Duration `mapstructure:"timeout"`
}

//nolint:gochecknoglobals // cobra boilerplate
var (
	cfgFile       string
	config        Config
	logger        *slog.Logger
	logLevel      = new(slog.LevelVar)
	cancelTimeout context.CancelFunc
	buildVersion  string
	buildCommit   string
	buildDate     string
	rootCmd       = &cobra.Command{
		Use:   "meta1v",
		Short: "Provides a way to interact with Canon's EFD files.",
		Long: `meta1v is a command line tool to interact with Canon's EFD files.

You can print out information to stdout about the film roll, including focus
points, custom functions, roll information, thumbnail previews, and more.`,
		Example: `  # View roll information
  meta1v roll list data.efd

  # Export frame data to CSV
  meta1v frame export data.efd output.csv

  # Write EXIF metadata to an image
  meta1v exif data.efd 1 image.jpg`,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			err := initialiseConfig(cmd)
			if err != nil {
				return fmt.Errorf("failed to initialise configuration: %w", err)
			}

			level := slog.LevelWarn
			switch strings.ToLower(config.Log.Level) {
			case "debug":
				level = slog.LevelDebug
			case "info":
				level = slog.LevelInfo
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

			ctx, cancel := context.WithTimeout(cmd.Context(), config.Timeout)
			cancelTimeout = cancel
			cmd.SetContext(ctx)

			return nil
		},
		PersistentPostRunE: func(_ *cobra.Command, _ []string) error {
			if cancelTimeout != nil {
				cancelTimeout()
			}

			return nil
		},
	}
)

// Execute runs the root command and handles any errors.
// This is called by main.main() and should only be called once.
func Execute(version, commit, date string) {
	buildVersion = version
	buildCommit = commit
	buildDate = date

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

//nolint:gochecknoinits,exhaustruct // cobra boilerplate, slog boilerplate
func init() {
	handler := tint.NewHandler(os.Stderr, &tint.Options{
		Level: logLevel,
	})
	logger = slog.New(handler)

	const defaultTimeout = 3 * time.Minute
	viper.SetDefault("timeout", defaultTimeout)

	rootCmd.PersistentFlags().
		StringVar(&cfgFile, "config", "", "config file (default is $HOME/.meta1v/config)")

	rootCmd.PersistentFlags().BoolVarP(
		&config.Strict,
		"strict",
		"s",
		false,
		"enable strict mode (fail on unknown metadata values)",
	)

	ctr := container.New(logger, osexec.NewLookPath())

	exifUseCase := exif.NewUseCase(logger, ctr.EFDService, ctr.ExifService)

	rootCmd.AddCommand(exif.NewCommand(logger, exifUseCase))
	rootCmd.AddCommand(roll.NewCommand(logger, ctr))
	rootCmd.AddCommand(customfunctions.NewCommand(logger, ctr))
	rootCmd.AddCommand(focusingpoints.NewCommand(logger, ctr))
	rootCmd.AddCommand(frame.NewCommand(logger, ctr))
	rootCmd.AddCommand(thumbnail.NewCommand(logger, ctr))
	rootCmd.AddCommand(newVersionCommand())
}

func initialiseConfig(cmd *cobra.Command) error {
	viper.SetEnvPrefix("META1V")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	viper.AutomaticEnv()

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

	if err = viper.Unmarshal(&config); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return nil
}

// Root exposes the root command for tools like doc generators.
func Root() *cobra.Command { return rootCmd }
