package display_test

import (
	"errors"
	"image"
	"math"
	"reflect"
	"testing"

	"github.com/ma-tf/meta1v/internal/service/display"
	display_test "github.com/ma-tf/meta1v/internal/service/display/mocks"
	"github.com/ma-tf/meta1v/pkg/records"
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
		expect     func(
			mockFrameMetadataBuilder *display_test.MockFrameMetadataBuilder,
			mockExposureSettingsBuilder *display_test.MockExposureSettingsBuilder,
			mockCameraModesBuilder *display_test.MockCameraModesBuilder,
			mockCustomFunctionsBuilder *display_test.MockCustomFunctionsBuilder,
			mockThumbnailBuilder *display_test.MockThumbnailBuilder,
			mockBuilder *display_test.MockDisplayableFrameBuilder,
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
				mockFrameMetadataBuilder *display_test.MockFrameMetadataBuilder,
				mockExposureSettingsBuilder *display_test.MockExposureSettingsBuilder,
				mockCameraModesBuilder *display_test.MockCameraModesBuilder,
				mockCustomFunctionsBuilder *display_test.MockCustomFunctionsBuilder,
				mockThumbnailBuilder *display_test.MockThumbnailBuilder,
				mockBuilder *display_test.MockDisplayableFrameBuilder,
			) {
				mockFrameMetadataBuilder.EXPECT().
					WithFrameMetadata(
						gomock.Any(),
					).
					Return(mockExposureSettingsBuilder)

				mockExposureSettingsBuilder.EXPECT().
					WithExposureSettings().
					Return(mockCameraModesBuilder)

				mockCameraModesBuilder.EXPECT().
					WithCameraModesAndFlashInfo().
					Return(mockCustomFunctionsBuilder)

				mockCustomFunctionsBuilder.EXPECT().
					WithCustomFunctionsAndFocusPoints().
					Return(mockThumbnailBuilder)

				mockThumbnailBuilder.EXPECT().
					WithThumbnail(gomock.Any()).
					Return(mockBuilder)

				mockBuilder.EXPECT().
					Build().
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

			mockFrameMetadataBuilder := display_test.NewMockFrameMetadataBuilder(
				ctrl,
			)
			mockExposureSettingsBuilder := display_test.NewMockExposureSettingsBuilder(
				ctrl,
			)
			mockCameraModesBuilder := display_test.NewMockCameraModesBuilder(
				ctrl,
			)
			mockCustomFunctionsBuilder := display_test.NewMockCustomFunctionsBuilder(
				ctrl,
			)
			mockThumbnailBuilder := display_test.NewMockThumbnailBuilder(
				ctrl,
			)
			mockBuilder := display_test.NewMockDisplayableFrameBuilder(
				ctrl,
			)

			if tt.expect != nil {
				tt.expect(
					mockFrameMetadataBuilder,
					mockExposureSettingsBuilder,
					mockCameraModesBuilder,
					mockCustomFunctionsBuilder,
					mockThumbnailBuilder,
					mockBuilder,
				)
			}

			factory := display.NewDisplayableRollFactory(
				mockFrameMetadataBuilder,
			)

			result, err := factory.Create(tt.rootRecord)

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
		expect     func(
			mockFrameMetadataBuilder *display_test.MockFrameMetadataBuilder,
			mockExposureSettingsBuilder *display_test.MockExposureSettingsBuilder,
			mockCameraModesBuilder *display_test.MockCameraModesBuilder,
			mockCustomFunctionsBuilder *display_test.MockCustomFunctionsBuilder,
			mockThumbnailBuilder *display_test.MockThumbnailBuilder,
			mockBuilder *display_test.MockDisplayableFrameBuilder,
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
				mockFrameMetadataBuilder *display_test.MockFrameMetadataBuilder,
				mockExposureSettingsBuilder *display_test.MockExposureSettingsBuilder,
				mockCameraModesBuilder *display_test.MockCameraModesBuilder,
				mockCustomFunctionsBuilder *display_test.MockCustomFunctionsBuilder,
				mockThumbnailBuilder *display_test.MockThumbnailBuilder,
				mockBuilder *display_test.MockDisplayableFrameBuilder,
			) {
				mockFrameMetadataBuilder.EXPECT().
					WithFrameMetadata(
						gomock.Any(),
					).
					Return(mockExposureSettingsBuilder).
					Times(math.MaxUint16)

				mockExposureSettingsBuilder.EXPECT().
					WithExposureSettings().
					Return(mockCameraModesBuilder).
					Times(math.MaxUint16)

				mockCameraModesBuilder.EXPECT().
					WithCameraModesAndFlashInfo().
					Return(mockCustomFunctionsBuilder).
					Times(math.MaxUint16)

				mockCustomFunctionsBuilder.EXPECT().
					WithCustomFunctionsAndFocusPoints().
					Return(mockThumbnailBuilder).
					Times(math.MaxUint16)

				mockThumbnailBuilder.EXPECT().
					WithThumbnail(gomock.Any()).
					Return(mockBuilder).
					Times(math.MaxUint16)

				mockBuilder.EXPECT().
					Build().
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
				mockFrameMetadataBuilder *display_test.MockFrameMetadataBuilder,
				mockExposureSettingsBuilder *display_test.MockExposureSettingsBuilder,
				mockCameraModesBuilder *display_test.MockCameraModesBuilder,
				mockCustomFunctionsBuilder *display_test.MockCustomFunctionsBuilder,
				mockThumbnailBuilder *display_test.MockThumbnailBuilder,
				mockBuilder *display_test.MockDisplayableFrameBuilder,
			) {
				mockFrameMetadataBuilder.EXPECT().
					WithFrameMetadata(
						gomock.Any(),
					).
					Return(mockExposureSettingsBuilder)

				mockExposureSettingsBuilder.EXPECT().
					WithExposureSettings().
					Return(mockCameraModesBuilder)

				mockCameraModesBuilder.EXPECT().
					WithCameraModesAndFlashInfo().
					Return(mockCustomFunctionsBuilder)

				mockCustomFunctionsBuilder.EXPECT().
					WithCustomFunctionsAndFocusPoints().
					Return(mockThumbnailBuilder)

				mockThumbnailBuilder.EXPECT().
					WithThumbnail(gomock.Any()).
					Return(mockBuilder)

				mockBuilder.EXPECT().
					Build().
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

			mockFrameMetadataBuilder := display_test.NewMockFrameMetadataBuilder(
				ctrl,
			)
			mockExposureSettingsBuilder := display_test.NewMockExposureSettingsBuilder(
				ctrl,
			)
			mockCameraModesBuilder := display_test.NewMockCameraModesBuilder(
				ctrl,
			)
			mockCustomFunctionsBuilder := display_test.NewMockCustomFunctionsBuilder(
				ctrl,
			)
			mockThumbnailBuilder := display_test.NewMockThumbnailBuilder(
				ctrl,
			)
			mockBuilder := display_test.NewMockDisplayableFrameBuilder(
				ctrl,
			)

			if tt.expect != nil {
				tt.expect(
					mockFrameMetadataBuilder,
					mockExposureSettingsBuilder,
					mockCameraModesBuilder,
					mockCustomFunctionsBuilder,
					mockThumbnailBuilder,
					mockBuilder,
				)
			}

			factory := display.NewDisplayableRollFactory(
				mockFrameMetadataBuilder,
			)

			result, err := factory.Create(tt.rootRecord)

			assertExpectedError(t, err, tt.expectedError)

			if !reflect.DeepEqual(result, tt.expectedOutput) {
				t.Fatalf("expected result %v, got %v",
					tt.expectedOutput, result)
			}
		})
	}
}
