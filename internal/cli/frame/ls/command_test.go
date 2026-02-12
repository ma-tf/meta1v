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

package ls_test

import (
	"bytes"
	"errors"
	"log/slog"
	"testing"

	"github.com/ma-tf/meta1v/internal/cli"
	"github.com/ma-tf/meta1v/internal/cli/frame/ls"
	ls_test "github.com/ma-tf/meta1v/internal/cli/frame/ls/mocks"
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
		expect         func(uc ls_test.MockUseCase, tt testcase)
		expectedError  error
	}

	tests := []testcase{
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
			expect: func(mockUseCase ls_test.MockUseCase, tt testcase) {
				mockUseCase.EXPECT().
					List(gomock.Any(), tt.args[0], gomock.Any()).
					Return(nil)
			},
		},
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

			mockUseCase := ls_test.NewMockUseCase(ctrl)

			if tt.expect != nil {
				tt.expect(*mockUseCase, tt)
			}

			cmd := ls.NewCommand(logger, mockUseCase)
			if tt.registerStrict {
				cmd.Flags().Bool("strict", false, "enable strict mode")
			}

			cmd.SetArgs(tt.args)

			err := cmd.Execute()

			assertError(t, tt, err)
		})
	}
}
