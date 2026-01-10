package container_test

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/ma-tf/meta1v/internal/container"
)

func TestNew(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	logger := slog.New(slog.NewTextHandler(buf, nil))

	ctr := container.New(logger)

	if ctr == nil {
		t.Fatal("expected container to be non-nil")
	}
}
