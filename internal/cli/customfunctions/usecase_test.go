package customfunctions_test

import (
	"errors"
	"os"
	"testing"

	"github.com/ma-tf/meta1v/internal/cli"
	"github.com/ma-tf/meta1v/internal/cli/customfunctions"
	csv_test "github.com/ma-tf/meta1v/internal/service/csv/mocks"
	"github.com/ma-tf/meta1v/internal/service/display"
	display_test "github.com/ma-tf/meta1v/internal/service/display/mocks"
	efd_test "github.com/ma-tf/meta1v/internal/service/efd/mocks"
	osfs_test "github.com/ma-tf/meta1v/internal/service/osfs/mocks"
	"github.com/ma-tf/meta1v/pkg/records"
	"go.uber.org/mock/gomock"
)

var errExample = errors.New("example error")

//nolint:exhaustruct // only partial is needed
func Test_CustomFunctionsUseCase_List(t *testing.T) {
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
			expectedError: customfunctions.ErrFailedToReadFile,
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
				EFDF: records.EFDF{
					Title: [64]byte{'t', 'i', 't', 'l', 'e'},
				},
			},
			expectedError: customfunctions.ErrFailedToParseFile,
		},
		{
			name: "failed to display custom functions",
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
					DisplayCustomFunctions(gomock.Any(), tt.roll).
					Return(customfunctions.ErrFailedToDisplay)
			},
			filename: "file.efd",
			records: records.Root{
				EFDF: records.EFDF{
					Title: [64]byte{'t', 'i', 't', 'l', 'e'},
				},
			},
			roll: display.DisplayableRoll{
				Title: "title",
			},
			expectedError: customfunctions.ErrFailedToDisplay,
		},
		{
			name: "successfully display custom functions",
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
					DisplayCustomFunctions(gomock.Any(), tt.roll).
					Return(nil)
			},
			filename: "file.efd",
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

			uc := customfunctions.NewListUseCase(
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
func Test_CustomFunctionsUseCase_Export(t *testing.T) {
	t.Parallel()

	type testcase struct {
		name            string
		efdFile         string
		records         records.Root
		displayableRoll display.DisplayableRoll
		outputFile      *string
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

	const (
		permission    os.FileMode = 0o666
		unforcedFlags             = os.O_WRONLY | os.O_CREATE | os.O_EXCL
		forceFlags                = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	)

	outputFileName := "output.csv"

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
					Return(tt.records, errExample)
			},
			expectedError: customfunctions.ErrFailedToReadFile,
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
			expectedError: customfunctions.ErrFailedToParseFile,
		},
		{
			name:       "output file exists and not forced",
			efdFile:    "file.efd",
			outputFile: &outputFileName,
			expect: func(
				mockEFDService *efd_test.MockService,
				mockDisplayableRollFactory *display_test.MockDisplayableRollFactory,
				_ *csv_test.MockService,
				mockFileSystem *osfs_test.MockFileSystem,
				_ *osfs_test.MockFile,
				tt testcase,
			) {
				mockEFDService.EXPECT().
					RecordsFromFile(gomock.Any(), tt.efdFile).
					Return(tt.records, nil)

				mockDisplayableRollFactory.EXPECT().
					Create(gomock.Any(), tt.records, tt.strict).
					Return(tt.displayableRoll, nil)

				mockFileSystem.EXPECT().
					OpenFile(*tt.outputFile, unforcedFlags, permission).
					Return(nil, os.ErrExist)
			},
			expectedError: cli.ErrOutputFileAlreadyExists,
		},
		{
			name:       "failed to create output file",
			efdFile:    "file.efd",
			outputFile: &outputFileName,
			force:      true,
			expect: func(
				mockEFDService *efd_test.MockService,
				mockDisplayableRollFactory *display_test.MockDisplayableRollFactory,
				_ *csv_test.MockService,
				mockFileSystem *osfs_test.MockFileSystem,
				_ *osfs_test.MockFile,
				tt testcase,
			) {
				mockEFDService.EXPECT().
					RecordsFromFile(gomock.Any(), tt.efdFile).
					Return(tt.records, nil)

				mockDisplayableRollFactory.EXPECT().
					Create(gomock.Any(), tt.records, tt.strict).
					Return(tt.displayableRoll, nil)

				mockFileSystem.EXPECT().
					OpenFile(*tt.outputFile, forceFlags, permission).
					Return(nil, errExample)
			},
			expectedError: customfunctions.ErrFailedToCreateOutputFile,
		},
		{
			name:       "failed to export to CSV",
			efdFile:    "file.efd",
			outputFile: &outputFileName,
			force:      true,
			expect: func(
				mockEFDService *efd_test.MockService,
				mockDisplayableRollFactory *display_test.MockDisplayableRollFactory,
				mockCSVService *csv_test.MockService,
				mockFileSystem *osfs_test.MockFileSystem,
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

				mockFileSystem.EXPECT().
					OpenFile(*tt.outputFile, forceFlags, permission).
					Return(mockFile, nil)

				mockCSVService.EXPECT().
					ExportCustomFunctions(mockFile, tt.displayableRoll).
					Return(errExample)
			},
			expectedError: customfunctions.ErrFailedToWriteCSV,
		},
		{
			name:       "successful export to CSV",
			efdFile:    "file.efd",
			outputFile: &outputFileName,
			force:      true,
			expect: func(
				mockEFDService *efd_test.MockService,
				mockDisplayableRollFactory *display_test.MockDisplayableRollFactory,
				mockCSVService *csv_test.MockService,
				mockFileSystem *osfs_test.MockFileSystem,
				mockFile *osfs_test.MockFile,
				tt testcase,
			) {
				mockEFDService.EXPECT().
					RecordsFromFile(gomock.Any(), tt.efdFile).
					Return(records.Root{}, nil)

				mockDisplayableRollFactory.EXPECT().
					Create(gomock.Any(), gomock.Any(), tt.strict).
					Return(display.DisplayableRoll{}, nil)

				mockFile.EXPECT().
					Close().
					Return(nil)

				mockFileSystem.EXPECT().
					OpenFile(*tt.outputFile, forceFlags, permission).
					Return(mockFile, nil)

				mockCSVService.EXPECT().
					ExportCustomFunctions(mockFile, gomock.Any()).
					Return(nil)
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
			mockCSVService := csv_test.NewMockService(ctrl)
			mockFileSystem := osfs_test.NewMockFileSystem(ctrl)
			mockFile := osfs_test.NewMockFile(ctrl)

			if tt.expect != nil {
				tt.expect(
					mockEFDService,
					mockDisplayableRollFactory,
					mockCSVService,
					mockFileSystem,
					mockFile,
					tt,
				)
			}

			uc := customfunctions.NewExportUseCase(
				mockEFDService,
				mockDisplayableRollFactory,
				mockCSVService,
				mockFileSystem,
			)

			err := uc.Export(
				ctx,
				tt.efdFile,
				tt.outputFile,
				tt.strict,
				tt.force,
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
