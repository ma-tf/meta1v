package frame_test

import (
	"bytes"
	"errors"
	"log/slog"
	"os"
	"testing"

	"github.com/ma-tf/meta1v/internal/cli"
	"github.com/ma-tf/meta1v/internal/cli/frame"
	"github.com/ma-tf/meta1v/internal/records"
	csv_test "github.com/ma-tf/meta1v/internal/service/csv/mocks"
	"github.com/ma-tf/meta1v/internal/service/display"
	display_test "github.com/ma-tf/meta1v/internal/service/display/mocks"
	efd_test "github.com/ma-tf/meta1v/internal/service/efd/mocks"
	osfs_test "github.com/ma-tf/meta1v/internal/service/osfs/mocks"
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
func Test_FrameListUseCase(t *testing.T) {
	t.Parallel()

	type testcase struct {
		name   string
		expect func(
			efd_test.MockService,
			display_test.MockDisplayableRollFactory,
			display_test.MockService,
			testcase,
		)
		filename      string
		records       records.Root
		roll          display.DisplayableRoll
		strict        bool
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
					RecordsFromFile(gomock.Any(), tt.filename).
					Return(
						records.Root{},
						errExample,
					)
			},
			filename:      "file.efd",
			expectedError: frame.ErrFailedToReadFile,
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
					RecordsFromFile(gomock.Any(), tt.filename).
					Return(
						tt.records,
						nil,
					)

				mockDisplayableRollFactory.EXPECT().
					Create(gomock.Any(), tt.records, tt.strict).
					Return(
						display.DisplayableRoll{},
						errExample,
					)
			},
			filename: "file.efd",
			records: records.Root{
				EFRMs: []records.EFRM{
					{
						Remarks: [256]byte{'r', 'e', 'm', 'a', 'r', 'k', 's'},
					},
				},
			},
			expectedError: frame.ErrFailedToParseFile,
		},
		{
			name: "successfully display frames",
			expect: func(
				mockEFDService efd_test.MockService,
				mockDisplayableRollFactory display_test.MockDisplayableRollFactory,
				mockDisplayService display_test.MockService,
				tt testcase,
			) {
				mockEFDService.EXPECT().
					RecordsFromFile(gomock.Any(), tt.filename).
					Return(
						tt.records,
						nil,
					)

				mockDisplayableRollFactory.EXPECT().
					Create(gomock.Any(), tt.records, tt.strict).
					Return(
						tt.roll,
						nil,
					)

				mockDisplayService.EXPECT().
					DisplayFrames(gomock.Any(), gomock.Any(), tt.roll)
			},
			filename: "file.efd",
			records: records.Root{
				EFRMs: []records.EFRM{
					{
						Remarks: [256]byte{'r', 'e', 'm', 'a', 'r', 'k', 's'},
					},
				},
			},
			roll: display.DisplayableRoll{
				Frames: []display.DisplayableFrame{
					{
						Remarks: "remarks",
					},
				},
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

			uc := frame.NewListUseCase(newTestLogger(),
				mockEFDService,
				mockDisplayableRollFactory,
				mockDisplayService,
			)

			err := uc.List(ctx, tt.filename, tt.strict)

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

//nolint:exhaustruct // only partial is needed
func Test_FrameExportUseCase(t *testing.T) {
	t.Parallel()

	type testcase struct {
		name            string
		efdFile         string
		records         records.Root
		outputFile      *string
		displayableRoll display.DisplayableRoll
		strict          bool
		force           bool
		expect          func(
			*efd_test.MockService,
			*display_test.MockDisplayableRollFactory,
			*csv_test.MockService,
			*osfs_test.MockFileSystem,
			*osfs_test.MockFile,
			testcase,
		)
		expectedError error
	}

	outputFile := "output.csv"

	const (
		unforcedFlags             = os.O_WRONLY | os.O_CREATE | os.O_EXCL
		forcedFlags               = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
		permissions   os.FileMode = 0o666
	)

	tests := []testcase{
		{
			name:    "failed to read file",
			efdFile: "file.efd",
			expect: func(
				mockEFDService *efd_test.MockService,
				_ *display_test.MockDisplayableRollFactory,
				_ *csv_test.MockService,
				_ *osfs_test.MockFileSystem,
				_ *osfs_test.MockFile,
				tt testcase,
			) {
				mockEFDService.EXPECT().
					RecordsFromFile(gomock.Any(), tt.efdFile).
					Return(records.Root{}, errExample)
			},
			expectedError: frame.ErrFailedToReadFile,
		},
		{
			name:    "failed to parse file",
			efdFile: "file.efd",
			expect: func(
				mockEFDService *efd_test.MockService,
				mockDisplayableRollFactory *display_test.MockDisplayableRollFactory,
				_ *csv_test.MockService,
				_ *osfs_test.MockFileSystem,
				_ *osfs_test.MockFile,
				tt testcase,
			) {
				mockEFDService.EXPECT().
					RecordsFromFile(gomock.Any(), tt.efdFile).
					Return(tt.records, nil)

				mockDisplayableRollFactory.EXPECT().
					Create(gomock.Any(), tt.records, tt.strict).
					Return(display.DisplayableRoll{}, errExample)
			},
			expectedError: frame.ErrFailedToParseFile,
		},
		{
			name:    "file already exists without force",
			efdFile: "file.efd",
			records: records.Root{
				EFRMs: []records.EFRM{
					{
						Remarks: [256]byte{'r', 'e', 'm', 'a', 'r', 'k', 's'},
					},
				},
			},
			outputFile: &outputFile,
			expect: func(
				mockEFDService *efd_test.MockService,
				mockDisplayableRollFactory *display_test.MockDisplayableRollFactory,
				_ *csv_test.MockService,
				mockFS *osfs_test.MockFileSystem,
				_ *osfs_test.MockFile,
				tt testcase,
			) {
				mockEFDService.EXPECT().
					RecordsFromFile(gomock.Any(), tt.efdFile).
					Return(tt.records, nil)

				mockDisplayableRollFactory.EXPECT().
					Create(gomock.Any(), tt.records, tt.strict).
					Return(tt.displayableRoll, nil)

				mockFS.EXPECT().
					OpenFile(*tt.outputFile, unforcedFlags, permissions).
					Return(nil, os.ErrExist)
			},
			expectedError: cli.ErrOutputFileAlreadyExists,
		},
		{
			name:    "failed to create file",
			efdFile: "file.efd",
			records: records.Root{
				EFRMs: []records.EFRM{
					{
						Remarks: [256]byte{'r', 'e', 'm', 'a', 'r', 'k', 's'},
					},
				},
			},
			outputFile: &outputFile,
			expect: func(
				mockEFDService *efd_test.MockService,
				mockDisplayableRollFactory *display_test.MockDisplayableRollFactory,
				_ *csv_test.MockService,
				mockFS *osfs_test.MockFileSystem,
				_ *osfs_test.MockFile,
				tt testcase,
			) {
				mockEFDService.EXPECT().
					RecordsFromFile(gomock.Any(), tt.efdFile).
					Return(tt.records, nil)

				mockDisplayableRollFactory.EXPECT().
					Create(gomock.Any(), tt.records, tt.strict).
					Return(tt.displayableRoll, nil)

				mockFS.EXPECT().
					OpenFile(*tt.outputFile, unforcedFlags, permissions).
					Return(nil, errExample)
			},
			expectedError: frame.ErrFailedToCreateOutputFile,
		},
		{
			name:    "failed to export frames",
			efdFile: "file.efd",
			records: records.Root{
				EFRMs: []records.EFRM{
					{
						Remarks: [256]byte{'r', 'e', 'm', 'a', 'r', 'k', 's'},
					},
				},
			},
			outputFile: &outputFile,
			displayableRoll: display.DisplayableRoll{
				Frames: []display.DisplayableFrame{
					{
						Remarks: "remarks",
					},
				},
			},
			force: true,
			expect: func(
				mockEFDService *efd_test.MockService,
				mockDisplayableRollFactory *display_test.MockDisplayableRollFactory,
				mockCSVService *csv_test.MockService,
				mockFS *osfs_test.MockFileSystem,
				mockFile *osfs_test.MockFile,
				tt testcase,
			) {
				mockEFDService.EXPECT().
					RecordsFromFile(gomock.Any(), tt.efdFile).
					Return(tt.records, nil)

				mockDisplayableRollFactory.EXPECT().
					Create(gomock.Any(), tt.records, tt.strict).
					Return(tt.displayableRoll, nil)

				mockFile.EXPECT().
					Close().
					Return(nil)

				mockFS.EXPECT().
					OpenFile(*tt.outputFile, forcedFlags, permissions).
					Return(mockFile, nil)

				mockCSVService.EXPECT().
					ExportFrames(gomock.Any(), mockFile, tt.displayableRoll).
					Return(errExample)
			},
			expectedError: frame.ErrFailedToExport,
		},
	}

	assertErrors := func(t *testing.T, got, expected error) {
		t.Helper()

		if expected != nil {
			if got == nil {
				t.Fatalf("expected error %v, got nil", expected)
			}

			if !errors.Is(got, expected) {
				t.Fatalf(
					"expected error %v to be in chain, got %v",
					expected,
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

			ctx := t.Context()

			mockEFDService := efd_test.NewMockService(ctrl)
			mockDisplayableRollFactory := display_test.NewMockDisplayableRollFactory(
				ctrl,
			)
			mockCSVService := csv_test.NewMockService(ctrl)
			mockFS := osfs_test.NewMockFileSystem(ctrl)
			mockFile := osfs_test.NewMockFile(ctrl)

			if tt.expect != nil {
				tt.expect(
					mockEFDService,
					mockDisplayableRollFactory,
					mockCSVService,
					mockFS,
					mockFile,
					tt,
				)
			}

			uc := frame.NewExportUseCase(newTestLogger(),
				mockEFDService,
				mockDisplayableRollFactory,
				mockCSVService,
				mockFS,
			)

			err := uc.Export(
				ctx,
				tt.efdFile,
				tt.outputFile,
				tt.strict,
				tt.force,
			)

			assertErrors(t, err, tt.expectedError)
		})
	}
}

//nolint:exhaustruct // only partial is needed
func Test_FrameExportUseCase_Success(t *testing.T) {
	t.Parallel()

	type testcase struct {
		name            string
		efdFile         string
		records         records.Root
		outputFile      *string
		displayableRoll display.DisplayableRoll
		strict          bool
		force           bool
		expect          func(
			*efd_test.MockService,
			*display_test.MockDisplayableRollFactory,
			*csv_test.MockService,
			*osfs_test.MockFileSystem,
			*osfs_test.MockFile,
			testcase,
		)
		expectedError error
	}

	outputFile := "output.csv"

	const (
		unforcedFlags             = os.O_WRONLY | os.O_CREATE | os.O_EXCL
		forcedFlags               = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
		permissions   os.FileMode = 0o666
	)

	tests := []testcase{
		{
			name:    "successfully export frames",
			efdFile: "file.efd",
			records: records.Root{
				EFRMs: []records.EFRM{
					{
						Remarks: [256]byte{'r', 'e', 'm', 'a', 'r', 'k', 's'},
					},
				},
			},
			outputFile: &outputFile,
			displayableRoll: display.DisplayableRoll{
				Frames: []display.DisplayableFrame{
					{
						Remarks: "remarks",
					},
				},
			},
			force: true,
			expect: func(
				mockEFDService *efd_test.MockService,
				mockDisplayableRollFactory *display_test.MockDisplayableRollFactory,
				mockCSVService *csv_test.MockService,
				mockFS *osfs_test.MockFileSystem,
				mockFile *osfs_test.MockFile,
				tt testcase,
			) {
				mockEFDService.EXPECT().
					RecordsFromFile(gomock.Any(), tt.efdFile).
					Return(tt.records, nil)

				mockDisplayableRollFactory.EXPECT().
					Create(gomock.Any(), tt.records, tt.strict).
					Return(tt.displayableRoll, nil)

				mockFile.EXPECT().
					Close().
					Return(nil)

				mockFS.EXPECT().
					OpenFile(*tt.outputFile, forcedFlags, permissions).
					Return(mockFile, nil)

				mockCSVService.EXPECT().
					ExportFrames(gomock.Any(), mockFile, tt.displayableRoll).
					Return(nil)
			},
		},
	}

	assertErrors := func(t *testing.T, got, expected error) {
		t.Helper()

		if expected != nil {
			if got == nil {
				t.Fatalf("expected error %v, got nil", expected)
			}

			if !errors.Is(got, expected) {
				t.Fatalf(
					"expected error %v to be in chain, got %v",
					expected,
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

			ctx := t.Context()

			mockEFDService := efd_test.NewMockService(ctrl)
			mockDisplayableRollFactory := display_test.NewMockDisplayableRollFactory(
				ctrl,
			)
			mockCSVService := csv_test.NewMockService(ctrl)
			mockFS := osfs_test.NewMockFileSystem(ctrl)
			mockFile := osfs_test.NewMockFile(ctrl)

			if tt.expect != nil {
				tt.expect(
					mockEFDService,
					mockDisplayableRollFactory,
					mockCSVService,
					mockFS,
					mockFile,
					tt,
				)
			}

			uc := frame.NewExportUseCase(newTestLogger(),
				mockEFDService,
				mockDisplayableRollFactory,
				mockCSVService,
				mockFS,
			)

			err := uc.Export(
				ctx,
				tt.efdFile,
				tt.outputFile,
				tt.strict,
				tt.force,
			)

			assertErrors(t, err, tt.expectedError)
		})
	}
}
