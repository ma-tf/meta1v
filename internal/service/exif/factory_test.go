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

package exif_test

import (
	"errors"
	"testing"

	"github.com/ma-tf/meta1v/internal/service/exif"
	osexec_test "github.com/ma-tf/meta1v/internal/service/osexec/mocks"
	"go.uber.org/mock/gomock"
)

func Test_CommandFactory_CreateCommand(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLookPath := osexec_test.NewMockLookPath(ctrl)
	mockLookPath.EXPECT().
		LookPath("exiftool").
		Return("/usr/bin/exiftool", nil)

	factory := exif.NewExiftoolCommandFactory(mockLookPath)

	// expect this to not panic
	_ = factory.CreateCommand(
		t.Context(),
		"test.jpg",
		nil,
		"metadata",
		nil,
	)
}

func Test_CommandFactory_LookPathFails(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLookPath := osexec_test.NewMockLookPath(ctrl)
	mockLookPath.EXPECT().
		LookPath("exiftool").
		Return("", errExample)

	defer func() {
		r := recover()

		err, ok := r.(error)
		if !ok {
			t.Errorf("expected panic with error, got: %v", r)

			return
		}

		if !errors.Is(err, exif.ErrExifToolBinaryNotFound) {
			t.Errorf("unexpected result or panic: %v", r)
		}
	}()

	_ = exif.NewExiftoolCommandFactory(mockLookPath)
}
