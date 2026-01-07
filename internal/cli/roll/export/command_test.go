package export_test

import (
	"bytes"
	"errors"
	"log/slog"
	"testing"

	"github.com/ma-tf/meta1v/internal/cli"
	"github.com/ma-tf/meta1v/internal/cli/roll/export"
	export_test "github.com/ma-tf/meta1v/internal/cli/roll/export/mocks"
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
		name   string
		args   []string
		expect func(
			uc export_test.MockUseCase,
			tt testcase,
		)
		expectedError error
	}

	tests := []testcase{
		{
			name:          "no efd file provided",
			args:          []string{},
			expect:        func(_ export_test.MockUseCase, _ testcase) {},
			expectedError: cli.ErrEFDFileMustBeProvided,
		},
		{
			name:          "only efd file provided",
			args:          []string{"file.efd"},
			expect:        func(_ export_test.MockUseCase, _ testcase) {},
			expectedError: cli.ErrTargetFileMustBeSpecified,
		},
		{
			name:          "too many arguments provided",
			args:          []string{"file.efd", "output.csv", "extra_arg"},
			expect:        func(_ export_test.MockUseCase, _ testcase) {},
			expectedError: cli.ErrTooManyArguments,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUseCase := export_test.NewMockUseCase(ctrl)

			if tt.expect != nil {
				tt.expect(*mockUseCase, tt)
			}

			cmd := export.NewCommand(logger, mockUseCase)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()

			if tt.expectedError != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.expectedError)
				}

				if !errors.Is(err, tt.expectedError) {
					t.Fatalf(
						"expected error %v to be in chain, got %v",
						tt.expectedError,
						err,
					)
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

//nolint:exhaustruct // only partial is needed
func Test_CommandRun(t *testing.T) {
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

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUseCase := export_test.NewMockUseCase(ctrl)

	mockUseCase.EXPECT().
		Export(gomock.Any(), "file.efd", "output.csv").
		Return(nil)

	cmd := export.NewCommand(logger, mockUseCase)
	cmd.SetArgs([]string{"file.efd", "output.csv"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
