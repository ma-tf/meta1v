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
func Test_NewCommand_Args(t *testing.T) {
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
		name          string
		args          []string
		expectedError error
	}

	tests := []testcase{
		{
			name:          "no arguments",
			args:          []string{},
			expectedError: cli.ErrEFDFileMustBeProvided,
		},
		{
			name:          "only first argument provided (efd file)",
			args:          []string{"file.efd"},
			expectedError: cli.ErrFrameNumberMustBeSpecified,
		},
		{
			name:          "only first and second arguments provided (efd file, frame number)",
			args:          []string{"file.efd", "1"},
			expectedError: cli.ErrTargetFileMustBeSpecified,
		},
		{
			name:          "too many arguments provided",
			args:          []string{"file.efd", "1", "target.jpg", "extra_arg"},
			expectedError: cli.ErrTooManyArguments,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUseCase := exif_test.NewMockUseCase(ctrl)

			cmd := exif.NewCommand(logger, mockUseCase)
			cmd.SilenceUsage = true
			cmd.SetArgs(tt.args)

			err := cmd.Execute()

			if tt.expectedError != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.expectedError)
				}

				if !errors.Is(err, tt.expectedError) {
					t.Fatalf("expected error %v, got %v", tt.expectedError, err)
				}

				return
			}

			if err != nil {
				t.Fatalf("expected no error, got %v", err)

				return
			}
		})
	}
}

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
		expect         func(
			mockUseCase *exif_test.MockUseCase,
			tc testcase,
		)
		expectedError error
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

	assertError := func(t *testing.T, got, want error) {
		t.Helper()

		if got == nil {
			t.Fatalf("expected error %v, got nil", want)
		}

		if !errors.Is(got, want) {
			t.Fatalf("expected error %v, got %v", want, got)
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

			if tt.expectedError != nil {
				assertError(t, err, tt.expectedError)

				return
			}

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}
