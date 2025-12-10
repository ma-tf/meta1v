package list_test

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/ma-tf/meta1v/internal/cli/roll/list"
	"github.com/ma-tf/meta1v/internal/service/display"
	display_test "github.com/ma-tf/meta1v/internal/service/display/mocks"
	efd_test "github.com/ma-tf/meta1v/internal/service/efd/mocks"
	"github.com/ma-tf/meta1v/pkg/records"
	"go.uber.org/mock/gomock"
)

var errExample = errors.New("example error")

//nolint:exhaustruct // for testcase struct literals
func Test_DisplayRoll(t *testing.T) {
	t.Parallel()

	type testcase struct {
		name   string
		expect func(
			efd_test.MockService,
			display_test.MockDisplayableRollFactory,
			display_test.MockService,
			testcase,
		)
		file          io.Reader
		records       records.Root
		roll          display.DisplayableRoll
		expectedError error
	}

	tests := []testcase{
		{
			name: "failed to read file",
			expect: func(
				mockEFDService efd_test.MockService,
				_ display_test.MockDisplayableRollFactory,
				_ display_test.MockService,
				tt testcase,
			) {
				mockEFDService.EXPECT().
					RecordsFromFile(gomock.Any(), tt.file).
					Return(
						records.Root{},
						errExample,
					)
			},
			file:          bytes.NewReader([]byte("data")),
			expectedError: list.ErrFailedToReadFile,
		},
		{
			name: "failed to parse file",
			expect: func(
				mockEFDService efd_test.MockService,
				mockDisplayableRollFactory display_test.MockDisplayableRollFactory,
				_ display_test.MockService,
				tt testcase,
			) {
				mockEFDService.EXPECT().
					RecordsFromFile(gomock.Any(), tt.file).
					Return(
						tt.records,
						nil,
					)

				mockDisplayableRollFactory.EXPECT().
					Create(tt.records).
					Return(
						display.DisplayableRoll{},
						errExample,
					)
			},
			file: bytes.NewReader([]byte("data")),
			records: records.Root{
				EFDF: records.EFDF{
					Title: [64]byte{'t', 'i', 't', 'l', 'e'},
				},
			},
			expectedError: list.ErrFailedToParseFile,
		},
		{
			name: "successfully display roll",
			expect: func(
				mockEFDService efd_test.MockService,
				mockDisplayableRollFactory display_test.MockDisplayableRollFactory,
				mockDisplayService display_test.MockService,
				tt testcase,
			) {
				mockEFDService.EXPECT().
					RecordsFromFile(gomock.Any(), tt.file).
					Return(
						tt.records,
						nil,
					)

				mockDisplayableRollFactory.EXPECT().
					Create(tt.records).
					Return(
						tt.roll,
						nil,
					)

				mockDisplayService.EXPECT().
					DisplayRoll(gomock.Any(), tt.roll)
			},
			file: bytes.NewReader([]byte("data")),
			records: records.Root{
				EFDF: records.EFDF{
					Title: [64]byte{'t', 'i', 't', 'l', 'e'},
				},
			},
			roll: display.DisplayableRoll{
				Title: "title",
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			ctx := t.Context()

			mockEFDService := efd_test.NewMockService(mockCtrl)
			mockDisplayableRollFactory := display_test.NewMockDisplayableRollFactory(
				mockCtrl,
			)
			mockDisplayService := display_test.NewMockService(mockCtrl)

			tt.expect(
				*mockEFDService,
				*mockDisplayableRollFactory,
				*mockDisplayService,
				tt,
			)

			uc := list.NewRollListUseCase(
				mockEFDService,
				mockDisplayableRollFactory,
				mockDisplayService,
			)

			err := uc.DisplayRoll(
				ctx,
				tt.file,
			)

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
