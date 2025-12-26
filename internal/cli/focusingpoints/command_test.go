package focusingpoints_test

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/ma-tf/meta1v/internal/cli/focusingpoints"
)

//nolint:exhaustruct // for testcase struct literals
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

	cmd := focusingpoints.NewCommand(logger)

	const expectedSubcommands = 1
	if len(cmd.Commands()) != expectedSubcommands {
		t.Fatalf("expected %d subcommand to be registered, got %d",
			expectedSubcommands, len(cmd.Commands()))
	}
}
