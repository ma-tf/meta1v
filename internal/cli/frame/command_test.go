package frame_test

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/ma-tf/meta1v/internal/cli/frame"
	"github.com/ma-tf/meta1v/internal/container"
	osexec_test "github.com/ma-tf/meta1v/internal/service/osexec/mocks"
	"go.uber.org/mock/gomock"
)

//nolint:exhaustruct // only partial is needed
func Test_NewCommand(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	buf := &bytes.Buffer{}
	logger := slog.New(slog.NewTextHandler(buf, &slog.HandlerOptions{
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Attr{}
			}

			return a
		},
	}))

	mockLookPath := osexec_test.NewMockLookPath(ctrl)
	mockLookPath.EXPECT().
		LookPath("exiftool").
		Return("/usr/bin/exiftool", nil)

	ctr := container.New(logger, mockLookPath)
	cmd := frame.NewCommand(logger, ctr)

	const expectedSubcommands = 2
	if len(cmd.Commands()) != expectedSubcommands {
		t.Fatalf("expected %d subcommand to be registered, got %d",
			expectedSubcommands, len(cmd.Commands()))
	}
}
