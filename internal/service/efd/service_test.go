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

package efd_test

import (
	"bytes"
	"errors"
	"io"
	"log/slog"
	"reflect"
	"testing"

	"github.com/ma-tf/meta1v/internal/records"
	"github.com/ma-tf/meta1v/internal/service/efd"
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

//nolint:exhaustruct // for records
func Test_RecordsFromFile(t *testing.T) {
	t.Parallel()

	type testcase struct {
		name     string
		filename string
		expect   func(
			mockFileSystem *osfs_test.MockFileSystem,
			mockReader *efd_test.MockReader,
			mockRootBuilder *efd_test.MockRootBuilder,
			file *osfs_test.MockFile,
			tt testcase,
		)
		expectedResult *records.Root
		expectedError  error
	}

	tests := []testcase{
		{
			name:     "file does not exist",
			filename: "nonexistent.efd",
			expect: func(
				mockFileSystem *osfs_test.MockFileSystem,
				_ *efd_test.MockReader,
				_ *efd_test.MockRootBuilder,
				_ *osfs_test.MockFile,
				tt testcase,
			) {
				mockFileSystem.EXPECT().
					Open(tt.filename).
					Return(nil, errExample)
			},
			expectedError: efd.ErrFailedToOpenFile,
		},
		{
			name:     "failed to read record",
			filename: "failed_to_read.efd",
			expect: func(
				mockFileSystem *osfs_test.MockFileSystem,
				mockReader *efd_test.MockReader,
				_ *efd_test.MockRootBuilder,
				mockFile *osfs_test.MockFile,
				tt testcase,
			) {
				mockFile.EXPECT().
					Close().
					Return(nil)

				mockReader.EXPECT().
					ReadRaw(gomock.Any(), mockFile).
					Return(records.Raw{}, errExample)

				mockFileSystem.EXPECT().
					Open(tt.filename).
					Return(mockFile, nil)
			},
			expectedError: efd.ErrFailedToReadRecord,
		},
		{
			name:     "failed to build root record",
			filename: "failed_to_build.efd",
			expect: func(
				mockFileSystem *osfs_test.MockFileSystem,
				mockReader *efd_test.MockReader,
				mockRootBuilder *efd_test.MockRootBuilder,
				mockFile *osfs_test.MockFile,
				tt testcase,
			) {
				mockFile.EXPECT().
					Close().
					Return(nil)

				mockReader.EXPECT().
					ReadRaw(gomock.Any(), mockFile).
					Return(records.Raw{}, io.EOF)

				mockRootBuilder.EXPECT().
					Build().
					Return(records.Root{}, errExample)

				mockFileSystem.EXPECT().
					Open(tt.filename).
					Return(mockFile, nil)
			},
			expectedError: efd.ErrFailedToBuildRoot,
		},
		{
			name: "successfully parsed EFD file",
			expect: func(
				mockFileSystem *osfs_test.MockFileSystem,
				mockReader *efd_test.MockReader,
				mockRootBuilder *efd_test.MockRootBuilder,
				mockFile *osfs_test.MockFile,
				tt testcase,
			) {
				mockFile.EXPECT().
					Close().
					Return(nil)

				efdfRaw, efrmRaw, eftpRaw := []byte{}, []byte{}, []byte{}

				mockReader.EXPECT().
					ReadRaw(gomock.Any(), mockFile).
					Return(records.Raw{
						Magic:  [4]byte{'E', 'F', 'D', 'F'},
						Length: uint64(len(efdfRaw)),
						Data:   efdfRaw,
					}, nil)

				mockReader.EXPECT().
					ReadRaw(gomock.Any(), mockFile).
					Return(records.Raw{
						Magic:  [4]byte{'E', 'F', 'R', 'M'},
						Length: uint64(len(efrmRaw)),
						Data:   efrmRaw,
					}, nil)

				mockReader.EXPECT().
					ReadRaw(gomock.Any(), mockFile).
					Return(records.Raw{
						Magic:  [4]byte{'E', 'F', 'T', 'P'},
						Length: uint64(len(eftpRaw)),
						Data:   eftpRaw,
					}, nil)

				mockReader.EXPECT().
					ReadRaw(gomock.Any(), mockFile).
					Return(records.Raw{}, io.EOF)

				efdf := records.EFDF{Title: [64]byte{'t', 'i', 't', 'l', 'e'}}
				mockReader.EXPECT().
					ReadEFDF(gomock.Any(), efdfRaw).
					Return(efdf, nil)

				efrm := records.EFRM{FrameNumber: 1}
				mockReader.EXPECT().
					ReadEFRM(gomock.Any(), efrmRaw).
					Return(efrm, nil)

				eftp := records.EFTP{Width: 100, Height: 100}
				mockReader.EXPECT().
					ReadEFTP(gomock.Any(), eftpRaw).
					Return(eftp, nil)

				mockRootBuilder.EXPECT().
					AddEFDF(gomock.Any(), efdf).
					Return(nil)

				mockRootBuilder.EXPECT().
					AddEFRM(gomock.Any(), efrm)

				mockRootBuilder.EXPECT().
					AddEFTP(gomock.Any(), eftp)

				mockRootBuilder.EXPECT().
					Build().
					Return(records.Root{
						EFDF:  efdf,
						EFRMs: []records.EFRM{efrm},
						EFTPs: []records.EFTP{eftp},
					}, nil)

				mockFileSystem.EXPECT().
					Open(tt.filename).
					Return(mockFile, nil)
			},
			expectedResult: &records.Root{
				EFDF:  records.EFDF{Title: [64]byte{'t', 'i', 't', 'l', 'e'}},
				EFRMs: []records.EFRM{{FrameNumber: 1}},
				EFTPs: []records.EFTP{{Width: 100, Height: 100}},
			},
			expectedError: nil,
		},
	}

	assertExpectedError := func(t *testing.T, got, expected error) {
		t.Helper()

		if expected == nil {
			return
		}

		if got == nil {
			t.Fatalf("expected error %v, got nil", expected)
		}

		if !errors.Is(got, expected) {
			t.Fatalf("expected error %v, got %v", expected, got)
		}
	}

	assertExpectedResult := func(t *testing.T, got, expected records.Root) {
		t.Helper()

		if !reflect.DeepEqual(got, expected) {
			t.Fatalf("expected result %v, got %v", expected, got)
		}
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ctx := t.Context()

			mockFileSystem := osfs_test.NewMockFileSystem(ctrl)
			mockReader := efd_test.NewMockReader(ctrl)
			mockRootBuilder := efd_test.NewMockRootBuilder(ctrl)
			mockFile := osfs_test.NewMockFile(ctrl)

			if tt.expect != nil {
				tt.expect(
					mockFileSystem,
					mockReader,
					mockRootBuilder,
					mockFile,
					tt,
				)
			}

			svc := efd.NewService(
				newTestLogger(),
				mockRootBuilder,
				mockReader,
				mockFileSystem,
			)

			result, err := svc.RecordsFromFile(
				ctx,
				tt.filename,
			)

			if tt.expectedError != nil {
				assertExpectedError(t, err, tt.expectedError)

				return
			}

			if tt.expectedResult == nil {
				t.Fatalf("test case missing expectedResult")
			}

			assertExpectedResult(t, result, *tt.expectedResult)
		})
	}
}

//nolint:exhaustruct // for records
func Test_RecordsFromFile_ProcessRecord(t *testing.T) {
	t.Parallel()

	type testcase struct {
		name     string
		filename string
		expect   func(
			mockFileSystem *osfs_test.MockFileSystem,
			mockReader *efd_test.MockReader,
			mockRootBuilder *efd_test.MockRootBuilder,
			mockFile *osfs_test.MockFile,
			tt testcase,
		)
		expectedError error
	}

	tests := []testcase{
		{
			name:     "failed to parse EFDF record",
			filename: "failed_to_parse_efdf.efd",
			expect: func(
				mockFileSystem *osfs_test.MockFileSystem,
				mockReader *efd_test.MockReader,
				_ *efd_test.MockRootBuilder,
				mockFile *osfs_test.MockFile,
				tt testcase,
			) {
				mockFile.EXPECT().
					Close().
					Return(nil)

				efdfRaw := []byte{}
				mockReader.EXPECT().
					ReadRaw(gomock.Any(), mockFile).
					Return(records.Raw{
						Magic:  [4]byte{'E', 'F', 'D', 'F'},
						Length: uint64(len(efdfRaw)),
						Data:   efdfRaw,
					}, nil)

				mockReader.EXPECT().
					ReadEFDF(gomock.Any(), efdfRaw).
					Return(records.EFDF{}, errExample)

				mockFileSystem.EXPECT().
					Open(tt.filename).
					Return(mockFile, nil)
			},
			expectedError: efd.ErrFailedToAddRecord,
		},
		{
			name:     "failed to add efdf to builder",
			filename: "failed_to_add_efdf.efd",
			expect: func(
				mockFileSystem *osfs_test.MockFileSystem,
				mockReader *efd_test.MockReader,
				mockRootBuilder *efd_test.MockRootBuilder,
				mockFile *osfs_test.MockFile,
				tt testcase,
			) {
				mockFile.EXPECT().
					Close().
					Return(nil)

				efdfRaw := []byte{}
				mockReader.EXPECT().
					ReadRaw(gomock.Any(), mockFile).
					Return(records.Raw{
						Magic:  [4]byte{'E', 'F', 'D', 'F'},
						Length: uint64(len(efdfRaw)),
						Data:   efdfRaw,
					}, nil)

				efdf := records.EFDF{Title: [64]byte{'t', 'i', 't', 'l', 'e'}}
				mockReader.EXPECT().
					ReadEFDF(gomock.Any(), efdfRaw).
					Return(efdf, nil)

				mockRootBuilder.EXPECT().
					AddEFDF(gomock.Any(), efdf).
					Return(errExample)

				mockFileSystem.EXPECT().
					Open(tt.filename).
					Return(mockFile, nil)
			},
			expectedError: efd.ErrFailedToAddRecord,
		},
		{
			name:     "failed to parse EFRM record",
			filename: "failed_to_parse_efrm.efd",
			expect: func(
				mockFileSystem *osfs_test.MockFileSystem,
				mockReader *efd_test.MockReader,
				_ *efd_test.MockRootBuilder,
				mockFile *osfs_test.MockFile,
				tt testcase,
			) {
				mockFile.EXPECT().
					Close().
					Return(nil)

				efrmRaw := []byte{}
				mockReader.EXPECT().
					ReadRaw(gomock.Any(), mockFile).
					Return(records.Raw{
						Magic:  [4]byte{'E', 'F', 'R', 'M'},
						Length: uint64(len(efrmRaw)),
						Data:   efrmRaw,
					}, nil)

				mockReader.EXPECT().
					ReadEFRM(gomock.Any(), efrmRaw).
					Return(records.EFRM{}, errExample)

				mockFileSystem.EXPECT().
					Open(tt.filename).
					Return(mockFile, nil)
			},
			expectedError: efd.ErrFailedToAddRecord,
		},
		{
			name:     "failed to parse EFTP record",
			filename: "failed_to_parse_eftp.efd",
			expect: func(
				mockFileSystem *osfs_test.MockFileSystem,
				mockReader *efd_test.MockReader,
				_ *efd_test.MockRootBuilder,
				mockFile *osfs_test.MockFile,
				tt testcase,
			) {
				mockFile.EXPECT().
					Close().
					Return(nil)

				eftpRaw := []byte{}
				mockReader.EXPECT().
					ReadRaw(gomock.Any(), mockFile).
					Return(records.Raw{
						Magic:  [4]byte{'E', 'F', 'T', 'P'},
						Length: uint64(len(eftpRaw)),
						Data:   eftpRaw,
					}, nil)

				mockReader.EXPECT().
					ReadEFTP(gomock.Any(), eftpRaw).
					Return(records.EFTP{}, errExample)

				mockFileSystem.EXPECT().
					Open(tt.filename).
					Return(mockFile, nil)
			},
			expectedError: efd.ErrFailedToAddRecord,
		},
		{
			name: "unknown record magic number",
			expect: func(
				mockFileSystem *osfs_test.MockFileSystem,
				mockReader *efd_test.MockReader,
				_ *efd_test.MockRootBuilder,
				mockFile *osfs_test.MockFile,
				tt testcase,
			) {
				mockFile.EXPECT().
					Close().
					Return(nil)

				mockReader.EXPECT().
					ReadRaw(gomock.Any(), mockFile).
					Return(records.Raw{
						Magic:  [4]byte{'X', 'Y', 'Z', 'W'},
						Length: 0,
						Data:   []byte{},
					}, nil)

				mockFileSystem.EXPECT().
					Open(tt.filename).
					Return(mockFile, nil)
			},
			expectedError: efd.ErrUnknownRecordType,
		},
	}

	assertExpectedError := func(t *testing.T, got, expected error) {
		t.Helper()

		if expected == nil {
			return
		}

		if got == nil {
			t.Fatalf("expected error %v, got nil", expected)
		}

		if !errors.Is(got, expected) {
			t.Fatalf("expected error %v, got %v", expected, got)
		}
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ctx := t.Context()

			mockFileSystem := osfs_test.NewMockFileSystem(ctrl)
			mockReader := efd_test.NewMockReader(ctrl)
			mockRootBuilder := efd_test.NewMockRootBuilder(ctrl)
			mockFile := osfs_test.NewMockFile(ctrl)

			if tt.expect != nil {
				tt.expect(
					mockFileSystem,
					mockReader,
					mockRootBuilder,
					mockFile,
					tt,
				)
			}

			svc := efd.NewService(
				newTestLogger(),
				mockRootBuilder,
				mockReader,
				mockFileSystem,
			)

			_, err := svc.RecordsFromFile(
				ctx,
				tt.filename,
			)

			if tt.expectedError != nil {
				assertExpectedError(t, err, tt.expectedError)

				return
			}

			t.Fatalf("expected error %v, got nil", tt.expectedError)
		})
	}
}
