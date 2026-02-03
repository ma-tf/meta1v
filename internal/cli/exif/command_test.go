package exif_test

import (
	"bytes"
	"errors"
	"log/slog"
	"testing"

	"github.com/ma-tf/meta1v/internal/cli"
	"github.com/ma-tf/meta1v/internal/cli/exif"
	exif_test "github.com/ma-tf/meta1v/internal/cli/exif/mocks"
	"go.uber.org/mock/gomock"
)

//nolint:exhaustruct // only partial is needed
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

	type testcase struct {
		name           string
		args           []string
		registerStrict bool
		expect         func(mockUseCase *exif_test.MockUseCase, tc testcase)
		expectedError  error
	}

	tests := []testcase{
		{
			name:           "strict flag not registered",
			args:           []string{"file.efd", "1", "target.jpg"},
			registerStrict: false,
			expectedError:  cli.ErrFailedToGetStrictFlag,
		},
		{
			name: "invalid frame number argument",
			args: []string{
				"file.efd",
				"invalid_frame_number",
				"target.jpg",
			},
			registerStrict: true,
			expectedError:  exif.ErrInvalidFrameNumber,
		},
		{
			name:           "valid arguments",
			args:           []string{"file.efd", "1", "target.jpg"},
			registerStrict: true,
			expect: func(
				mockUseCase *exif_test.MockUseCase,
				tc testcase,
			) {
				mockUseCase.
					EXPECT().
					ExportExif(
						gomock.Any(),
						tc.args[0],
						1,
						tc.args[2],
						false,
					).
					Return(nil)
			},
		},
	}

	assertError := func(t *testing.T, expected, got error) {
		t.Helper()

		if expected != nil {
			if got == nil {
				t.Fatalf("expected error %v, got nil", expected)
			}

			if !errors.Is(got, expected) {
				t.Fatalf("expected error %v, got %v", expected, got)
			}

			return
		}

		if got != nil {
			t.Fatalf("expected no error, got %v", got)
		}
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUseCase := exif_test.NewMockUseCase(ctrl)

			if tt.expect != nil {
				tt.expect(mockUseCase, tt)
			}

			cmd := exif.NewCommand(logger, mockUseCase)
			cmd.SilenceUsage = true

			if tt.registerStrict {
				cmd.Flags().Bool("strict", false, "enable strict mode")
			}

			cmd.SetArgs(tt.args)

			err := cmd.Execute()

			assertError(t, tt.expectedError, err)
		})
	}
}
