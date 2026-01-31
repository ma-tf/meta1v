package exif_test

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/ma-tf/meta1v/internal/service/exif"
	exif_test "github.com/ma-tf/meta1v/internal/service/exif/mocks"
	osexec_test "github.com/ma-tf/meta1v/internal/service/osexec/mocks"
	osfs_test "github.com/ma-tf/meta1v/internal/service/osfs/mocks"
	"go.uber.org/mock/gomock"
)

//nolint:exhaustruct // only partial is needed
func Test_Run(t *testing.T) {
	t.Parallel()

	type testcase struct {
		name       string
		targetFile string
		metadata   string
		cancelFunc context.CancelFunc
		expect     func(
			mockFileSystem *osfs_test.MockFileSystem,
			mockFactory *exif_test.MockExiftoolCommandFactory,
			mockCmd *osexec_test.MockCommand,
			tc testcase,
		)
		expectedError error
	}

	tests := []testcase{
		{
			name:       "pipe creation fails",
			targetFile: "test.jpg",
			metadata:   "metadata",
			expect: func(
				mockFileSystem *osfs_test.MockFileSystem,
				_ *exif_test.MockExiftoolCommandFactory,
				_ *osexec_test.MockCommand,
				_ testcase,
			) {
				mockFileSystem.
					EXPECT().
					Pipe().
					Return(nil, nil, errExample)
			},
			expectedError: exif.ErrCreatePipe,
		},
		{
			name:       "exiftool start fails",
			targetFile: "test.jpg",
			metadata:   "metadata",
			expect: func(
				mockFileSystem *osfs_test.MockFileSystem,
				mockFactory *exif_test.MockExiftoolCommandFactory,
				mockCmd *osexec_test.MockCommand,
				_ testcase,
			) {
				rPipe, wPipe, _ := os.Pipe()

				mockFileSystem.
					EXPECT().
					Pipe().
					Return(rPipe, wPipe, nil)

				mockFactory.
					EXPECT().
					CreateCommand(
						gomock.Any(),
						"test.jpg",
						gomock.Any(),
						"metadata",
						rPipe,
					).
					Return(mockCmd)

				mockCmd.EXPECT().
					Start().
					Return(errExample)
			},
			expectedError: exif.ErrStartExifTool,
		},
		{
			name:       "exiftool run fails",
			targetFile: "test.jpg",
			metadata:   "metadata",
			expect: func(
				mockFileSystem *osfs_test.MockFileSystem,
				mockFactory *exif_test.MockExiftoolCommandFactory,
				mockCmd *osexec_test.MockCommand,
				_ testcase,
			) {
				rPipe, wPipe, _ := os.Pipe()

				mockFileSystem.
					EXPECT().
					Pipe().
					Return(rPipe, wPipe, nil)

				mockFactory.
					EXPECT().
					CreateCommand(
						gomock.Any(),
						"test.jpg",
						gomock.Any(),
						"metadata",
						rPipe,
					).
					Return(mockCmd)

				mockCmd.EXPECT().
					Start().
					Return(nil)

				mockCmd.EXPECT().
					Wait().
					Return(errExample)
			},
			expectedError: exif.ErrExifToolFailed,
		},
		{
			name:       "context done before writing config",
			targetFile: "test.jpg",
			metadata:   "metadata",
			expect: func(
				mockFileSystem *osfs_test.MockFileSystem,
				mockFactory *exif_test.MockExiftoolCommandFactory,
				mockCmd *osexec_test.MockCommand,
				tc testcase,
			) {
				rPipe, wPipe, _ := os.Pipe()

				mockFileSystem.
					EXPECT().
					Pipe().
					Return(rPipe, wPipe, nil)

				mockFactory.
					EXPECT().
					CreateCommand(
						gomock.Any(),
						"test.jpg",
						gomock.Any(),
						"metadata",
						rPipe,
					).
					Return(mockCmd)

				mockCmd.EXPECT().
					Start().
					DoAndReturn(func() error {
						tc.cancelFunc()

						return nil
					})

				mockCmd.EXPECT().
					Wait().
					Return(nil)
			},
			expectedError: exif.ErrContextDone,
		},
		{
			name:       "writing config fails",
			targetFile: "test.jpg",
			metadata:   "metadata",
			expect: func(
				mockFileSystem *osfs_test.MockFileSystem,
				mockFactory *exif_test.MockExiftoolCommandFactory,
				mockCmd *osexec_test.MockCommand,
				_ testcase,
			) {
				rPipe, wPipe, _ := os.Pipe()

				mockFileSystem.
					EXPECT().
					Pipe().
					Return(rPipe, wPipe, nil)

				mockFactory.
					EXPECT().
					CreateCommand(
						gomock.Any(),
						"test.jpg",
						gomock.Any(),
						"metadata",
						rPipe,
					).
					Return(mockCmd)

				mockCmd.EXPECT().
					Start().
					Return(nil)

				mockCmd.EXPECT().
					Wait().
					Return(nil)

				// Close the write pipe to cause a write error.
				wPipe.Close()
			},
			expectedError: exif.ErrWriteExifToolConfig,
		},
		{
			name:       "exiftool runs successfully",
			targetFile: "test.jpg",
			metadata:   "metadata",
			expect: func(
				mockFileSystem *osfs_test.MockFileSystem,
				mockFactory *exif_test.MockExiftoolCommandFactory,
				mockCmd *osexec_test.MockCommand,
				_ testcase,
			) {
				rPipe, wPipe, _ := os.Pipe()

				mockFileSystem.
					EXPECT().
					Pipe().
					Return(rPipe, wPipe, nil)

				mockFactory.
					EXPECT().
					CreateCommand(
						gomock.Any(),
						"test.jpg",
						gomock.Any(),
						"metadata",
						rPipe,
					).
					Return(mockCmd)

				mockCmd.EXPECT().
					Start().
					Return(nil)

				mockCmd.EXPECT().
					Wait().
					Return(nil)
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()

			tt.cancelFunc = cancel

			mockFileSystem := osfs_test.NewMockFileSystem(ctrl)
			mockFactory := exif_test.NewMockExiftoolCommandFactory(ctrl)
			mockCmd := osexec_test.NewMockCommand(ctrl)

			if tt.expect != nil {
				tt.expect(
					mockFileSystem,
					mockFactory,
					mockCmd,
					tt,
				)
			}

			runner := exif.NewExifToolRunner(
				mockFileSystem,
				mockFactory,
			)

			err := runner.Run(
				ctx,
				tt.targetFile,
				tt.metadata,
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
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}
