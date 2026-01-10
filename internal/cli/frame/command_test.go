package frame_test

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/ma-tf/meta1v/internal/cli/frame"
	"github.com/ma-tf/meta1v/internal/container"
)

//nolint:exhaustruct // only partial is needed
func Test_NewCommand(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	logger := slog.New(slog.NewTextHandler(buf, &slog.HandlerOptions{
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Attr{}
			}

			return a
		},
	}))

	ctr := container.New(logger)
	cmd := frame.NewCommand(logger, ctr)

	const expectedSubcommands = 1
	if len(cmd.Commands()) != expectedSubcommands {
		t.Fatalf("expected %d subcommand to be registered, got %d",
			expectedSubcommands, len(cmd.Commands()))
	}
}
