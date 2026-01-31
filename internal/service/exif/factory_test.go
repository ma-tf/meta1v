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
