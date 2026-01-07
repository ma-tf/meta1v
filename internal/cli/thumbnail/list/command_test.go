package list_test

import (
	"bytes"
	"errors"
	"log/slog"
	"testing"

	"github.com/ma-tf/meta1v/internal/cli"
	"github.com/ma-tf/meta1v/internal/cli/thumbnail/list"
	list_test "github.com/ma-tf/meta1v/internal/cli/thumbnail/list/mocks"
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
		name   string
		args   []string
		expect func(
			uc list_test.MockUseCase,
			tt testcase,
		)
		expectedError error
	}

	tests := []testcase{
		{
			name:          "no filename provided",
			args:          []string{},
			expect:        func(_ list_test.MockUseCase, _ testcase) {},
			expectedError: cli.ErrEFDFileMustBeProvided,
		},
		{
			name: "successful execution",
			expect: func(
				mockUseCase list_test.MockUseCase,
				tt testcase,
			) {
				mockUseCase.EXPECT().
					DisplayThumbnails(gomock.Any(), tt.args[0]).
					Return(nil)
			},
			args: []string{"file.efd"},
		},
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
