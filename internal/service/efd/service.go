//go:generate mockgen -destination=./mocks/service_mock.go -package=efd_test github.com/ma-tf/meta1v/internal/service/efd Service

// Package efd provides services for reading and parsing Canon EFD binary files.
//
// The service reads EFD files, processes the binary records (EFDF, EFRM, EFTP),
// and constructs a structured representation of the film roll metadata.
package efd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"

	"github.com/ma-tf/meta1v/internal/records"
	"github.com/ma-tf/meta1v/internal/service/osfs"
)

var (
	ErrFailedToOpenFile         = errors.New("failed to open specified file")
	ErrInvalidRecordMagicNumber = errors.New("invalid record magic number")
	ErrFailedToReadRecord       = errors.New("failed to read record from file")
	ErrFailedToAddRecord        = errors.New("failed to add record to builder")
	ErrFailedToBuildRoot        = errors.New("failed to build root record")
	ErrMultipleEFDFRecords      = errors.New("multiple EFDF records found")
	ErrUnknownRecordType        = errors.New("unknown record type")
	ErrFailedToParseThumbnail   = errors.New("failed to parse EFTP thumbnail")
)

// Service provides operations for reading Canon EFD files and extracting structured metadata.
type Service interface {
	// RecordsFromFile reads an EFD file and returns the parsed Root structure containing
	// film roll metadata, frame records, and thumbnails.
	RecordsFromFile(ctx context.Context, filename string) (records.Root, error)
}

type service struct {
	log     *slog.Logger
	builder RootBuilder
	reader  Reader
	fs      osfs.FileSystem
}

func NewService(
	log *slog.Logger,
	builder RootBuilder,
	reader Reader,
	fs osfs.FileSystem,
) Service {
	return &service{
		log:     log,
		builder: builder,
		reader:  reader,
		fs:      fs,
	}
}

func (s *service) RecordsFromFile(
	ctx context.Context,
	filename string,
) (records.Root, error) {
	s.log.InfoContext(ctx, "parsing efd file", slog.String("file", filename))

	file, errFile := s.fs.Open(filename)
	if errFile != nil {
		return records.Root{}, fmt.Errorf("%w %q: %w",
			ErrFailedToOpenFile, filename, errFile)
	}
	defer file.Close()

	s.log.DebugContext(ctx, "opened file:", slog.String("filename", filename))

	recordCount := 0

	for {
		record, errRaw := s.reader.ReadRaw(ctx, file)
		if errors.Is(errRaw, io.EOF) {
			break
		}

		if errRaw != nil {
			return records.Root{}, fmt.Errorf("%w %q: %w",
				ErrFailedToReadRecord, filename, errRaw)
		}

		recordCount++

		if errProcess := s.processRecord(ctx, record); errProcess != nil {
			return records.Root{}, errProcess
		}
	}

	s.log.DebugContext(ctx, "all records read",
		slog.Int("total_records", recordCount))

	root, err := s.builder.Build()
	if err != nil {
		return records.Root{}, fmt.Errorf("%w %q: %w",
			ErrFailedToBuildRoot, filename, err)
	}

	s.log.InfoContext(ctx, "efd file parsed successfully",
		slog.String("file", filename),
		slog.Int("efrms", len(root.EFRMs)),
		slog.Int("eftps", len(root.EFTPs)))

	return root, nil
}

func (s *service) processRecord(ctx context.Context, record records.Raw) error {
	magic := string(record.Magic[:])
	switch magic {
	case "EFDF":
		efdf, errRead := s.reader.ReadEFDF(ctx, record.Data)
		if errRead != nil {
			return errors.Join(ErrFailedToAddRecord, errRead)
		}

		if err := s.builder.AddEFDF(ctx, efdf); err != nil {
			return errors.Join(ErrFailedToAddRecord, err)
		}

		s.log.DebugContext(ctx, "efdf record processed")

		return nil
	case "EFRM":
		efrm, errRead := s.reader.ReadEFRM(ctx, record.Data)
		if errRead != nil {
			return errors.Join(ErrFailedToAddRecord, errRead)
		}

		s.builder.AddEFRM(ctx, efrm)

		s.log.DebugContext(ctx, "efrm record processed",
			slog.Uint64("frame_number", uint64(efrm.FrameNumber)))

		return nil
	case "EFTP":
		eftp, errRead := s.reader.ReadEFTP(ctx, record.Data)
		if errRead != nil {
			return errors.Join(ErrFailedToAddRecord, errRead)
		}

		s.builder.AddEFTP(ctx, eftp)

		s.log.DebugContext(ctx, "eftp record processed",
			slog.Uint64("index", uint64(eftp.Index)))

		return nil
	default:
		return fmt.Errorf(
			"%w: found %q, expected EFDF (file record), EFRM (frame record), or EFTP (thumbnail record)",
			ErrUnknownRecordType,
			magic,
		)
	}
}
