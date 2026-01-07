package cli_test

import (
	"testing"

	"github.com/ma-tf/meta1v/internal/cli"
)

func Test_NewCommand(t *testing.T) {
	t.Parallel()

	cmd := cli.NewCommand()

	if cmd != nil && cmd.Use != "meta1v" {
		t.Errorf("unexpected command use: got %s, want %s", cmd.Use, "meta1v")
	}
}
