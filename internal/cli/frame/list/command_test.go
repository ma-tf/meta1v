package list_test

import (
	"bytes"
	"errors"
	"log/slog"
	"testing"

	"github.com/ma-tf/meta1v/internal/cli"
	"github.com/ma-tf/meta1v/internal/cli/frame/list"
	list_test "github.com/ma-tf/meta1v/internal/cli/frame/list/mocks"
	"go.uber.org/mock/gomock"
)

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

	type testcase struct {
		name           string
		args           []string
		registerStrict bool
		expect         func(
			uc list_test.MockUseCase,
			tt testcase,
		)
		expectedError error
	}

	tests := []testcase{
		{
			name:          "no filename provided",
			args:          []string{},
			expectedError: cli.ErrEFDFileMustBeProvided,
		},
		{
			name:           "strict flag not registered",
			args:           []string{"file.efd"},
			registerStrict: false,
			expectedError:  cli.ErrFailedToGetStrictFlag,
		},
		{
			name:           "successful execution",
			args:           []string{"file.efd"},
			registerStrict: true,
			expect: func(mockUseCase list_test.MockUseCase, tt testcase) {
				mockUseCase.EXPECT().
					List(gomock.Any(), tt.args[0], gomock.Any()).
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
			t.Fatalf("expected error %v to be in chain, got %v", want, got)
		}
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUseCase := list_test.NewMockUseCase(ctrl)

			if tt.expect != nil {
				tt.expect(*mockUseCase, tt)
			}

			cmd := list.NewCommand(logger, mockUseCase)
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
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
