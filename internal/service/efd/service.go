//go:generate mockgen -destination=./mocks/service_mock.go -package=efd_test github.com/ma-tf/meta1v/internal/service/efd Service
package efd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"

	"github.com/ma-tf/meta1v/internal/service/osfs"
	"github.com/ma-tf/meta1v/pkg/records"
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

type Service interface {
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
	file, errFile := s.fs.Open(filename)
	if errFile != nil {
		return records.Root{}, fmt.Errorf(
			"%w %q: %w",
			ErrFailedToOpenFile,
			filename,
			errFile,
		)
	}
	defer file.Close()

	s.log.DebugContext(ctx, "opened file:",
		slog.String("filename", filename))

	for {
		record, errRaw := s.reader.ReadRaw(ctx, file)
		if errors.Is(errRaw, io.EOF) {
			break
		}

		if errRaw != nil {
			return records.Root{}, fmt.Errorf(
				"%w %q: %w",
				ErrFailedToReadRecord,
				filename,
				errRaw,
			)
		}

		if errProcess := s.processRecord(ctx, record); errProcess != nil {
			return records.Root{}, errProcess
		}
	}

	root, err := s.builder.Build()
	if err != nil {
		return records.Root{}, fmt.Errorf(
			"%w %q: %w",
			ErrFailedToBuildRoot,
			filename,
			err,
		)
	}

	s.log.DebugContext(ctx, "efd records parsed",
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

		return nil
	case "EFRM":
		efrm, errRead := s.reader.ReadEFRM(ctx, record.Data)
		if errRead != nil {
			return errors.Join(ErrFailedToAddRecord, errRead)
		}

		s.builder.AddEFRM(ctx, efrm)

		return nil
	case "EFTP":
		eftp, errRead := s.reader.ReadEFTP(ctx, record.Data)
		if errRead != nil {
			return errors.Join(ErrFailedToAddRecord, errRead)
		}

		s.builder.AddEFTP(ctx, eftp)

		return nil
	default:
		return fmt.Errorf(
			"%w: found %q, expected EFDF (file record), EFRM (frame record), or EFTP (thumbnail record)",
			ErrUnknownRecordType,
			magic,
		)
	}
}
