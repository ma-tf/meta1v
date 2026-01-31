package exif_test

import (
	"bytes"
	"errors"
	"log/slog"
	"testing"

	"github.com/ma-tf/meta1v/internal/service/exif"
	exif_test "github.com/ma-tf/meta1v/internal/service/exif/mocks"
	"github.com/ma-tf/meta1v/pkg/records"
	"go.uber.org/mock/gomock"
)

var errExample = errors.New("example error")

//nolint:exhaustruct // only partial is needed
func newTestLogger() *slog.Logger {
	buf := &bytes.Buffer{}

	return slog.New(slog.NewTextHandler(buf, &slog.HandlerOptions{
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Attr{}
			}

			return a
		},
	}))
}

//nolint:exhaustruct // only partial is needed
func Test_WriteEXIF(t *testing.T) {
	t.Parallel()

	type testcase struct {
		name     string
		frame    records.EFRM
		filename string
		strict   bool
		expect   func(
			mockToolRunner *exif_test.MockToolRunner,
			mockBuilder *exif_test.MockBuilder,
			tc testcase,
		)
		expectedError error
	}

	tests := []testcase{
		{
			name: "failed to build exif data",
			frame: records.EFRM{
				FrameNumber: 1,
			},
			filename: "test.jpg",
			strict:   true,
			expect: func(
				_ *exif_test.MockToolRunner,
				mockBuilder *exif_test.MockBuilder,
				tc testcase,
			) {
				mockBuilder.EXPECT().Build(tc.frame, tc.strict).
					Return(nil, errExample)
			},
			expectedError: exif.ErrBuildExifData,
		},
		{
			name: "failed to run exiftool",
			frame: records.EFRM{
				FrameNumber: 2,
			},
			filename: "test2.jpg",
			strict:   false,
			expect: func(
				mockToolRunner *exif_test.MockToolRunner,
				mockBuilder *exif_test.MockBuilder,
				tc testcase,
			) {
				mockBuilder.EXPECT().Build(tc.frame, tc.strict).
					Return(map[string]string{
						"Tag1": "Value1",
						"Tag2": "Value2",
						"Tag3": "",
					}, nil)

				mockToolRunner.EXPECT().Run(
					gomock.Any(),
					tc.filename,
					"-Tag1=Value1\n-Tag2=Value2\n",
				).Return(errExample)
			},
			expectedError: exif.ErrRunExifTool,
		},
		{
			name: "successful exif write",
			frame: records.EFRM{
				FrameNumber: 3,
			},
			filename: "test3.jpg",
			strict:   true,
			expect: func(
				mockToolRunner *exif_test.MockToolRunner,
				mockBuilder *exif_test.MockBuilder,
				tc testcase,
			) {
				mockBuilder.EXPECT().Build(tc.frame, tc.strict).
					Return(map[string]string{
						"TagA": "ValueA",
						"TagB": "ValueB",
						"TagC": "",
					}, nil)

				mockToolRunner.EXPECT().Run(
					gomock.Any(),
					tc.filename,
					"-TagA=ValueA\n-TagB=ValueB\n",
				).Return(nil)
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ctx := t.Context()
			logger := newTestLogger()

			mockToolRunner := exif_test.NewMockToolRunner(ctrl)
			mockBuilder := exif_test.NewMockBuilder(ctrl)

			if tt.expect != nil {
				tt.expect(mockToolRunner, mockBuilder, tt)
			}

			svc := exif.NewService(
				logger,
				mockToolRunner,
				mockBuilder,
			)

			err := svc.WriteEXIF(
				ctx,
				tt.frame,
				tt.filename,
				tt.strict,
			)

			if tt.expectedError != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.expectedError)
				}

				if !errors.Is(err, tt.expectedError) {
					t.Fatalf("expected error %v, got %v", tt.expectedError, err)
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
