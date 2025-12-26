package thumbnail_test

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/ma-tf/meta1v/internal/cli/thumbnail"
)

//nolint:exhaustruct // for testcase struct literals
func Test_CommandRun(t *testing.T) {
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

	cmd := thumbnail.NewCommand(logger)

	const expectedSubcommands = 1
	if len(cmd.Commands()) != expectedSubcommands {
		t.Fatalf("expected %d subcommand to be registered, got %d",
			expectedSubcommands, len(cmd.Commands()))
	}
}
