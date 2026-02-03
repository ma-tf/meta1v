package export_test

import (
	"bytes"
	"errors"
	"log/slog"
	"testing"

	"github.com/ma-tf/meta1v/internal/cli"
	"github.com/ma-tf/meta1v/internal/cli/customfunctions/export"
	export_test "github.com/ma-tf/meta1v/internal/cli/customfunctions/export/mocks"
	"github.com/spf13/cobra"
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
		name          string
		strict        *bool
		force         *bool
		args          []string
		expect        func(uc export_test.MockUseCase, tt testcase)
		expectedError error
	}

	setFalse := func() *bool {
		b := false

		return &b
	}

	setTrue := func() *bool {
		b := true

		return &b
	}

	tests := []testcase{
		{
			name:          "failed to get strict flag",
			strict:        nil,
			force:         nil,
			args:          []string{"file.efd", "output.csv"},
			expect:        func(_ export_test.MockUseCase, _ testcase) {},
			expectedError: cli.ErrFailedToGetStrictFlag,
		},
		{
			name:          "failed to get force flag",
			strict:        setTrue(),
			force:         nil,
			args:          []string{"file.efd", "output.csv"},
			expect:        func(_ export_test.MockUseCase, _ testcase) {},
			expectedError: cli.ErrFailedToGetForceFlag,
		},
		{
			name:          "force flag without target file",
			strict:        setTrue(),
			force:         setTrue(),
			args:          []string{"file.efd"},
			expect:        func(_ export_test.MockUseCase, _ testcase) {},
			expectedError: cli.ErrForceFlagRequiresTargetFile,
		},
		{
			name:   "successful export to stdout (1 arg)",
			strict: setTrue(),
			force:  setFalse(),
			args:   []string{"file.efd"},
			expect: func(uc export_test.MockUseCase, tt testcase) {
				uc.EXPECT().
					Export(gomock.Any(), tt.args[0], nil, *tt.strict, *tt.force).
					Return(nil)
			},
			expectedError: nil,
		},
		{
			name:   "successful export with target file",
			strict: setTrue(),
			force:  setTrue(),
			args:   []string{"file.efd", "output.csv"},
			expect: func(uc export_test.MockUseCase, tt testcase) {
				uc.EXPECT().
					Export(gomock.Any(), tt.args[0], &tt.args[1], *tt.strict, *tt.force).
					Return(nil)
			},
			expectedError: nil,
		},
	}

	arrangeCmd := func(tt testcase, mockUseCase *export_test.MockUseCase) *cobra.Command {
		cmd := export.NewCommand(logger, mockUseCase)
		cmd.ResetFlags()

		if tt.strict != nil {
			cmd.Flags().Bool("strict", *tt.strict, "enable strict mode")
		}

		if tt.force != nil {
			cmd.Flags().Bool("force", *tt.force, "enable force mode")
		}

		cmd.SetArgs(tt.args)

		return cmd
	}

	assertError := func(t *testing.T, tt testcase, got error) {
		t.Helper()

		if tt.expectedError != nil {
			if got == nil {
				t.Fatalf("expected error %v, got nil", tt.expectedError)
			}

			if !errors.Is(got, tt.expectedError) {
				t.Fatalf(
					"expected error %v to be in chain, got %v",
					tt.expectedError,
					got,
				)
			}

			return
		}

		if got != nil {
			t.Fatalf("unexpected error: %v", got)
		}
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

			cmd := arrangeCmd(tt, mockUseCase)

			err := cmd.Execute()

			assertError(t, tt, err)
		})
	}
}
