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

package roll_test

import (
	"bytes"
	"errors"
	"log/slog"
	"os"
	"testing"

	"github.com/ma-tf/meta1v/internal/cli"
	"github.com/ma-tf/meta1v/internal/cli/roll"
	"github.com/ma-tf/meta1v/internal/records"
	csvexport_test "github.com/ma-tf/meta1v/internal/service/csvexport/mocks"
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
func Test_RollListUseCase(t *testing.T) {
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
			expectedError: roll.ErrFailedToReadFile,
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
			expectedError: roll.ErrFailedToParseFile,
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
					DisplayRoll(gomock.Any(), gomock.Any(), tt.roll)
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

			uc := roll.NewListUseCase(newTestLogger(),
				mockEFDService,
				mockDisplayableRollFactory,
				mockDisplayService,
			)

			err := uc.List(
				ctx,
				tt.filename,
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

//nolint:exhaustruct // only partial is needed
func Test_RollExportUseCase(t *testing.T) {
	t.Parallel()

	type testcase struct {
		name       string
		efdFile    string
		outputFile *string
		records    records.Root
		roll       display.DisplayableRoll
		strict     bool
		force      bool
		expect     func(
			*efd_test.MockService,
			*display_test.MockDisplayableRollFactory,
			*osfs_test.MockFileSystem,
			*csvexport_test.MockService,
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
			records: records.Root{},
			expect: func(
				mockEFDService *efd_test.MockService,
				_ *display_test.MockDisplayableRollFactory,
				_ *osfs_test.MockFileSystem,
				_ *csvexport_test.MockService,
				_ *osfs_test.MockFile,
				tt testcase,
			) {
				mockEFDService.EXPECT().
					RecordsFromFile(gomock.Any(), tt.efdFile).
					Return(tt.records, errExample)
			},
			expectedError: roll.ErrFailedToReadFile,
		},
		{
			name:    "failed to parse file",
			efdFile: "file.efd",
			records: records.Root{},
			roll:    display.DisplayableRoll{},
			expect: func(
				mockEFDService *efd_test.MockService,
				mockDisplayableRollFactory *display_test.MockDisplayableRollFactory,
				_ *osfs_test.MockFileSystem,
				_ *csvexport_test.MockService,
				_ *osfs_test.MockFile,
				tt testcase,
			) {
				mockEFDService.EXPECT().
					RecordsFromFile(gomock.Any(), tt.efdFile).
					Return(tt.records, nil)

				mockDisplayableRollFactory.EXPECT().
					Create(gomock.Any(), tt.records, tt.strict).
					Return(
						display.DisplayableRoll{},
						errExample,
					)
			},
			expectedError: roll.ErrFailedToParseFile,
		},
		{
			name:       "file already exists",
			efdFile:    "file.efd",
			outputFile: &outputFile,
			records:    records.Root{},
			roll:       display.DisplayableRoll{},
			expect: func(
				mockEFDService *efd_test.MockService,
				mockDisplayableRollFactory *display_test.MockDisplayableRollFactory,
				mockFileSystem *osfs_test.MockFileSystem,
				_ *csvexport_test.MockService,
				_ *osfs_test.MockFile,
				tt testcase,
			) {
				mockEFDService.EXPECT().
					RecordsFromFile(gomock.Any(), tt.efdFile).
					Return(tt.records, nil)

				mockDisplayableRollFactory.EXPECT().
					Create(gomock.Any(), tt.records, tt.strict).
					Return(tt.roll, nil)

				mockFileSystem.EXPECT().
					OpenFile(*tt.outputFile, unforcedFlags, permissions).
					Return(nil, os.ErrExist)
			},
			expectedError: cli.ErrOutputFileAlreadyExists,
		},
		{
			name:       "failed to create output file",
			efdFile:    "file.efd",
			outputFile: &outputFile,
			records:    records.Root{},
			roll:       display.DisplayableRoll{},
			force:      true,
			expect: func(
				mockEFDService *efd_test.MockService,
				mockDisplayableRollFactory *display_test.MockDisplayableRollFactory,
				mockFileSystem *osfs_test.MockFileSystem,
				_ *csvexport_test.MockService,
				_ *osfs_test.MockFile,
				tt testcase,
			) {
				mockEFDService.EXPECT().
					RecordsFromFile(gomock.Any(), tt.efdFile).
					Return(tt.records, nil)

				mockDisplayableRollFactory.EXPECT().
					Create(gomock.Any(), tt.records, tt.strict).
					Return(tt.roll, nil)

				mockFileSystem.EXPECT().
					OpenFile(*tt.outputFile, forcedFlags, permissions).
					Return(nil, errExample)
			},
			expectedError: roll.ErrFailedToCreateOutputFile,
		},
		{
			name:       "failed to export roll to CSV",
			efdFile:    "file.efd",
			outputFile: &outputFile,
			records:    records.Root{},
			roll:       display.DisplayableRoll{},
			force:      true,
			expect: func(
				mockEFDService *efd_test.MockService,
				mockDisplayableRollFactory *display_test.MockDisplayableRollFactory,
				mockFileSystem *osfs_test.MockFileSystem,
				mockCSVService *csvexport_test.MockService,
				mockFile *osfs_test.MockFile,
				tt testcase,
			) {
				mockEFDService.EXPECT().
					RecordsFromFile(gomock.Any(), tt.efdFile).
					Return(tt.records, nil)

				mockDisplayableRollFactory.EXPECT().
					Create(gomock.Any(), tt.records, tt.strict).
					Return(tt.roll, nil)

				mockFile.EXPECT().
					Close().
					Return(nil)

				mockFileSystem.EXPECT().
					OpenFile(*tt.outputFile, forcedFlags, permissions).
					Return(mockFile, nil)

				mockCSVService.EXPECT().
					ExportRoll(gomock.Any(), mockFile, tt.roll).
					Return(errExample)
			},
			expectedError: roll.ErrFailedToExport,
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

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			ctx := t.Context()

			mockEFDService := efd_test.NewMockService(mockCtrl)
			mockDisplayableRollFactory := display_test.NewMockDisplayableRollFactory(
				mockCtrl,
			)
			mockCSVService := csvexport_test.NewMockService(mockCtrl)
			mockFileSystem := osfs_test.NewMockFileSystem(mockCtrl)
			mockFile := osfs_test.NewMockFile(mockCtrl)

			if tt.expect != nil {
				tt.expect(
					mockEFDService,
					mockDisplayableRollFactory,
					mockFileSystem,
					mockCSVService,
					mockFile,
					tt,
				)
			}

			uc := roll.NewExportUseCase(newTestLogger(),
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

			assertErrors(t, err, tt.expectedError)
		})
	}
}

//nolint:exhaustruct // only partial is needed
func Test_RollExportUseCase_Success(t *testing.T) {
	t.Parallel()

	type testcase struct {
		name       string
		efdFile    string
		outputFile *string
		records    records.Root
		roll       display.DisplayableRoll
		strict     bool
		force      bool
		expect     func(
			*efd_test.MockService,
			*display_test.MockDisplayableRollFactory,
			*osfs_test.MockFileSystem,
			*csvexport_test.MockService,
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
			name:       "successfully export roll to CSV",
			efdFile:    "file.efd",
			outputFile: &outputFile,
			records:    records.Root{},
			roll:       display.DisplayableRoll{},
			force:      true,
			expect: func(
				mockEFDService *efd_test.MockService,
				mockDisplayableRollFactory *display_test.MockDisplayableRollFactory,
				mockFileSystem *osfs_test.MockFileSystem,
				mockCSVService *csvexport_test.MockService,
				mockFile *osfs_test.MockFile,
				tt testcase,
			) {
				mockEFDService.EXPECT().
					RecordsFromFile(gomock.Any(), tt.efdFile).
					Return(tt.records, nil)

				mockDisplayableRollFactory.EXPECT().
					Create(gomock.Any(), tt.records, tt.strict).
					Return(tt.roll, nil)

				mockFile.EXPECT().
					Close().
					Return(nil)

				mockFileSystem.EXPECT().
					OpenFile(*tt.outputFile, forcedFlags, permissions).
					Return(mockFile, nil)

				mockCSVService.EXPECT().
					ExportRoll(gomock.Any(), mockFile, tt.roll).
					Return(nil)
			},
			expectedError: nil,
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

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			ctx := t.Context()

			mockEFDService := efd_test.NewMockService(mockCtrl)
			mockDisplayableRollFactory := display_test.NewMockDisplayableRollFactory(
				mockCtrl,
			)
			mockCSVService := csvexport_test.NewMockService(mockCtrl)
			mockFileSystem := osfs_test.NewMockFileSystem(mockCtrl)
			mockFile := osfs_test.NewMockFile(mockCtrl)

			if tt.expect != nil {
				tt.expect(
					mockEFDService,
					mockDisplayableRollFactory,
					mockFileSystem,
					mockCSVService,
					mockFile,
					tt,
				)
			}

			uc := roll.NewExportUseCase(newTestLogger(),
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

			assertErrors(t, err, tt.expectedError)
		})
	}
}
