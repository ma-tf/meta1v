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
