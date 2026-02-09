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

package display_test

import (
	"errors"
	"image"
	"math"
	"reflect"
	"testing"

	"github.com/ma-tf/meta1v/internal/records"
	"github.com/ma-tf/meta1v/internal/service/display"
	display_test "github.com/ma-tf/meta1v/internal/service/display/mocks"
	"go.uber.org/mock/gomock"
)

var errExample = errors.New("example error")

//nolint:exhaustruct // only partial is needed
func outOfRangeEFRMs() []records.EFRM {
	hugeList := make([]records.EFRM, math.MaxUint16+1)
	for i := range hugeList {
		hugeList[i] = records.EFRM{}
	}

	return hugeList
}

//nolint:exhaustruct // only partial is needed
func Test_DisplayableRollFactory_Create(t *testing.T) {
	t.Parallel()

	type testcase struct {
		name       string
		rootRecord records.Root
		strict     bool
		expect     func(
			mockBuilder *display_test.MockBuilder,
		)
		expectedOutput display.DisplayableRoll
		expectedError  error
	}

	testcases := []testcase{
		{
			name: "invalid film id",
			rootRecord: records.Root{
				EFDF: records.EFDF{
					CodeA: 100,
					CodeB: 1000,
				},
			},
			expectedError: display.ErrFailedToParseRollData,
		},
		{
			name: "invalid first row",
			rootRecord: records.Root{
				EFDF: records.EFDF{
					FirstRow: 10,
					PerRow:   5,
				},
			},
			expectedError: display.ErrFailedToParseRollData,
		},
		{
			name: "invalid loaded date",
			rootRecord: records.Root{
				EFDF: records.EFDF{
					Year:  2023,
					Month: 13,
				},
			},
			expectedError: display.ErrFailedToParseRollData,
		},
		{
			name: "successful creation with one frame and thumbnail",
			rootRecord: records.Root{
				EFDF: records.EFDF{
					CodeA:    1,
					CodeB:    1,
					FirstRow: 0,
					PerRow:   10,
					Year:     2023,
					Month:    5,
					Day:      15,
					Hour:     12,
					Minute:   30,
					Second:   0,
				},
				EFRMs: []records.EFRM{
					{},
				},
				EFTPs: []records.EFTP{
					{
						Index:     1,
						Filepath:  [256]byte{'p', 'a', 't', 'h'},
						Width:     1,
						Height:    1,
						Thumbnail: image.NewRGBA(image.Rect(0, 0, 1, 1)),
					},
				},
			},
			expect: func(
				mockBuilder *display_test.MockBuilder,
			) {
				mockBuilder.EXPECT().
					Build(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(
						display.DisplayableFrame{},
						nil,
					)
			},
			expectedOutput: display.DisplayableRoll{
				FilmID:         "01-001",
				FirstRow:       "10",
				PerRow:         "10",
				FilmLoadedDate: "2023-05-15 12:30:00",
				Frames:         []display.DisplayableFrame{{}},
				FrameCount:     "0",
				IsoDX:          "0",
			},
			expectedError: nil,
		},
	}

	assertExpectedError := func(t *testing.T, got, expected error) {
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
			t.Fatalf("unexpected error: %v", got)
		}
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ctx := t.Context()

			mockBuilder := display_test.NewMockBuilder(
				ctrl,
			)

			if tt.expect != nil {
				tt.expect(mockBuilder)
			}

			factory := display.NewDisplayableRollFactory(
				mockBuilder,
			)

			result, err := factory.Create(ctx, tt.rootRecord, tt.strict)

			assertExpectedError(t, err, tt.expectedError)

			if !reflect.DeepEqual(result, tt.expectedOutput) {
				t.Fatalf("expected result %v, got %v",
					tt.expectedOutput, result)
			}
		})
	}
}

//nolint:exhaustruct // only partial is needed
func Test_DisplayableRollFactory_Create_FrameAndThumbnailErrors(t *testing.T) {
	t.Parallel()

	type testcase struct {
		name       string
		rootRecord records.Root
		strict     bool
		expect     func(
			mockBuilder *display_test.MockBuilder,
		)
		expectedOutput display.DisplayableRoll
		expectedError  error
	}

	testcases := []testcase{
		{
			name: "frame with multiple thumbnails for one frame",
			rootRecord: records.Root{
				EFDF: records.EFDF{
					CodeA:    1,
					CodeB:    1,
					FirstRow: 0,
					PerRow:   10,
					Year:     2023,
					Month:    5,
					Day:      15,
					Hour:     12,
					Minute:   30,
					Second:   0,
				},
				EFRMs: []records.EFRM{
					{},
				},
				EFTPs: []records.EFTP{
					{
						Index:     1,
						Filepath:  [256]byte{'p', 'a', 't', 'h'},
						Width:     1,
						Height:    1,
						Thumbnail: image.NewRGBA(image.Rect(0, 0, 1, 1)),
					},
					{
						Index:     1,
						Filepath:  [256]byte{'p', 'a', 't', 'h'},
						Width:     1,
						Height:    1,
						Thumbnail: image.NewRGBA(image.Rect(0, 0, 1, 1)),
					},
				},
			},
			expectedError: display.ErrMultipleThumbnailsForFrame,
		},
		{
			name: "frame with out of bounds thumbnail index",
			rootRecord: records.Root{
				EFDF: records.EFDF{
					CodeA:    1,
					CodeB:    1,
					FirstRow: 0,
					PerRow:   10,
					Year:     math.MaxUint16,
					Month:    math.MaxUint8,
					Day:      math.MaxUint8,
					Hour:     math.MaxUint8,
					Minute:   math.MaxUint8,
					Second:   math.MaxUint8,
				},
				EFRMs: outOfRangeEFRMs(),
			},
			expect: func(
				mockBuilder *display_test.MockBuilder,
			) {
				mockBuilder.EXPECT().
					Build(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(
						display.DisplayableFrame{},
						nil,
					).
					Times(math.MaxUint16)
			},
			expectedError: display.ErrFrameIndexOutOfRange,
		},
		{
			name: "frame builder error",
			rootRecord: records.Root{
				EFDF: records.EFDF{
					CodeA:    1,
					CodeB:    1,
					FirstRow: 0,
					PerRow:   10,
					Year:     2023,
					Month:    5,
					Day:      15,
					Hour:     12,
					Minute:   30,
					Second:   0,
				},
				EFRMs: []records.EFRM{
					{},
				},
			},
			expect: func(
				mockBuilder *display_test.MockBuilder,
			) {
				mockBuilder.EXPECT().
					Build(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(
						display.DisplayableFrame{},
						errExample,
					)
			},
			expectedError: display.ErrFailedToBuildFrame,
		},
	}

	assertExpectedError := func(t *testing.T, got, expected error) {
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
			t.Fatalf("unexpected error: %v", got)
		}
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ctx := t.Context()

			mockBuilder := display_test.NewMockBuilder(
				ctrl,
			)

			if tt.expect != nil {
				tt.expect(mockBuilder)
			}

			factory := display.NewDisplayableRollFactory(
				mockBuilder,
			)

			result, err := factory.Create(ctx, tt.rootRecord, tt.strict)

			assertExpectedError(t, err, tt.expectedError)

			if !reflect.DeepEqual(result, tt.expectedOutput) {
				t.Fatalf("expected result %v, got %v",
					tt.expectedOutput, result)
			}
		})
	}
}
