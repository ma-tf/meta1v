package exif_test

import (
	"errors"
	"testing"

	"github.com/ma-tf/meta1v/internal/cli/exif"
	efd_test "github.com/ma-tf/meta1v/internal/service/efd/mocks"
	exif_test "github.com/ma-tf/meta1v/internal/service/exif/mocks"
	"github.com/ma-tf/meta1v/pkg/records"
	"go.uber.org/mock/gomock"
)

var errExample = errors.New("example error")

//nolint:exhaustruct // only partial is needed
func Test_ExportExif(t *testing.T) {
	t.Parallel()

	type testcase struct {
		name       string
		efdFile    string
		frame      int
		targetFile string
		strict     bool
		root       records.Root
		expect     func(
			efdTestMock efd_test.MockService,
			exifTestMock exif_test.MockService,
			tc testcase,
		)
		expectedError error
	}

	tests := []testcase{
		{
			name:    "failed to interpret EFD",
			efdFile: "file.efd",
			expect: func(
				mockEFDService efd_test.MockService,
				_ exif_test.MockService,
				tt testcase,
			) {
				mockEFDService.EXPECT().
					RecordsFromFile(gomock.Any(), tt.efdFile).
					Return(records.Root{}, errExample)
			},
			expectedError: exif.ErrFailedToInterpretEFD,
		},
		{
			name:    "duplicate frame number",
			efdFile: "file.efd",
			frame:   1,
			root: records.Root{
				EFRMs: []records.EFRM{
					{FrameNumber: 1},
					{FrameNumber: 1},
				},
			},
			expect: func(
				mockEFDService efd_test.MockService,
				_ exif_test.MockService,
				tt testcase,
			) {
				mockEFDService.EXPECT().
					RecordsFromFile(gomock.Any(), tt.efdFile).
					Return(tt.root, nil)
			},
			expectedError: exif.ErrDuplicateFrameNumber,
		},
		{
			name:    "frame number not found",
			efdFile: "file.efd",
			frame:   2,
			root: records.Root{
				EFRMs: []records.EFRM{
					{FrameNumber: 1},
					{FrameNumber: 3},
				},
			},
			expect: func(
				mockEFDService efd_test.MockService,
				_ exif_test.MockService,
				tt testcase,
			) {
				mockEFDService.EXPECT().
					RecordsFromFile(gomock.Any(), tt.efdFile).
					Return(tt.root, nil)
			},
			expectedError: exif.ErrFrameNumberNotFound,
		},
		{
			name:       "write EXIF failed",
			efdFile:    "file.efd",
			frame:      1,
			targetFile: "target.jpg",
			strict:     true,
			root: records.Root{
				EFRMs: []records.EFRM{
					{FrameNumber: 1},
				},
			},
			expect: func(
				mockEFDService efd_test.MockService,
				mockEXIFService exif_test.MockService,
				tt testcase,
			) {
				mockEFDService.EXPECT().
					RecordsFromFile(gomock.Any(), tt.efdFile).
					Return(tt.root, nil)

				mockEXIFService.EXPECT().
					WriteEXIF(
						gomock.Any(),
						tt.root.EFRMs[0],
						tt.targetFile,
						tt.strict,
					).
					Return(errExample)
			},
			expectedError: exif.ErrWriteEXIFFailed,
		},
		{
			name:       "successful EXIF export",
			efdFile:    "file.efd",
			frame:      1,
			targetFile: "target.jpg",
			strict:     true,
			root: records.Root{
				EFRMs: []records.EFRM{
					{FrameNumber: 1},
				},
			},
			expect: func(
				mockEFDService efd_test.MockService,
				mockEXIFService exif_test.MockService,
				tt testcase,
			) {
				mockEFDService.EXPECT().
					RecordsFromFile(gomock.Any(), tt.efdFile).
					Return(tt.root, nil)

				mockEXIFService.EXPECT().
					WriteEXIF(
						gomock.Any(),
						tt.root.EFRMs[0],
						tt.targetFile,
						tt.strict,
					).
					Return(nil)
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
			mockEXIFService := exif_test.NewMockService(mockCtrl)

			if tt.expect != nil {
				tt.expect(
					*mockEFDService,
					*mockEXIFService,
					tt,
				)
			}

			useCase := exif.NewUseCase(
				mockEFDService,
				mockEXIFService,
			)

			err := useCase.ExportExif(
				ctx,
				tt.efdFile,
				tt.frame,
				tt.targetFile,
				tt.strict,
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
