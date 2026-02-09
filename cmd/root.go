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
	rootCmd       = &cobra.Command{
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

			level := slog.LevelInfo
			switch strings.ToLower(config.Log.Level) {
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
func Execute() {
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

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().
		StringVar(&cfgFile, "config", "", "config file (default is $HOME/.meta1v/config)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.PersistentFlags().BoolVarP(
		&config.Strict,
		"strict",
		"s",
		false,
		"enable strict mode (fail on unknown metadata values)",
	)

	ctr := container.New(logger)

	exifUseCase := exif.NewUseCase(logger, ctr.EFDService, ctr.ExifService)

	rootCmd.AddCommand(exif.NewCommand(logger, exifUseCase))
	rootCmd.AddCommand(roll.NewCommand(logger, ctr))
	rootCmd.AddCommand(customfunctions.NewCommand(logger, ctr))
	rootCmd.AddCommand(focusingpoints.NewCommand(logger, ctr))
	rootCmd.AddCommand(frame.NewCommand(logger, ctr))
	rootCmd.AddCommand(thumbnail.NewCommand(logger, ctr))
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
