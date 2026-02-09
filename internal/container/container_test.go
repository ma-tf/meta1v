package container_test

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/ma-tf/meta1v/internal/container"
	osexec_test "github.com/ma-tf/meta1v/internal/service/osexec/mocks"
	"go.uber.org/mock/gomock"
)

func TestNew(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	buf := &bytes.Buffer{}
	logger := slog.New(slog.NewTextHandler(buf, nil))

	mockLookPath := osexec_test.NewMockLookPath(ctrl)
	mockLookPath.EXPECT().
		LookPath("exiftool").
		Return("/usr/bin/exiftool", nil)

	ctr := container.New(logger, mockLookPath)

	if ctr == nil {
		t.Fatal("expected container to be non-nil")
	}
}
