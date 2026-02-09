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
