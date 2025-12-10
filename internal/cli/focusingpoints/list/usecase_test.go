package list_test

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/ma-tf/meta1v/internal/cli/focusingpoints/list"
	"github.com/ma-tf/meta1v/internal/service/display"
	display_test "github.com/ma-tf/meta1v/internal/service/display/mocks"
	efd_test "github.com/ma-tf/meta1v/internal/service/efd/mocks"
	"github.com/ma-tf/meta1v/pkg/records"
	"go.uber.org/mock/gomock"
)

var errExample = errors.New("example error")

//nolint:exhaustruct // for testcase struct literals
func Test_FocusingPointsListUseCase(t *testing.T) {
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
			name: "failed to list focusing points",
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
					DisplayFocusingPoints(gomock.Any(), tt.roll).
					Return(
						errExample,
					)
			},
			file: bytes.NewReader([]byte("data")),
			records: records.Root{
				EFDF: records.EFDF{
					Title: [64]byte{'t', 'i', 't', 'l', 'e'},
				},
			},
			roll:          display.DisplayableRoll{},
			expectedError: list.ErrFailedToList,
		},
		{
			name: "successfully display focusing points",
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
					DisplayFocusingPoints(gomock.Any(), tt.roll).
					Return(
						nil,
					)
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
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ctx := t.Context()

			mockEFDService := efd_test.NewMockService(ctrl)
			mockDisplayableRollFactory := display_test.NewMockDisplayableRollFactory(
				ctrl,
			)
			mockDisplayService := display_test.NewMockService(ctrl)

			if tt.expect != nil {
				tt.expect(
					*mockEFDService,
					*mockDisplayableRollFactory,
					*mockDisplayService,
					tt,
				)
			}

			uc := list.NewFocusingPointsListUseCase(
				mockEFDService,
				mockDisplayableRollFactory,
				mockDisplayService,
			)

			err := uc.DisplayFocusingPoints(ctx, tt.file)

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
