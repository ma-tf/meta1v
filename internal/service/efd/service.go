//go:generate mockgen -destination=./mocks/service_mock.go -package=efd_test github.com/ma-tf/meta1v/internal/service/efd Service
package efd

import (
	"context"
	"encoding/binary"
	"errors"
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

var _ Service = &service{log: nil, builder: nil, fs: nil}

type Service interface {
	RecordsFromFile(ctx context.Context, filename string) (records.Root, error)
}

type service struct {
	log     *slog.Logger
	builder RootBuilder
	fs      osfs.FileSystem
}

func NewService(
	log *slog.Logger,
	builder RootBuilder,
	fs osfs.FileSystem,
) Service {
	return &service{
		log:     log,
		builder: builder,
		fs:      fs,
	}
}

func (s *service) RecordsFromFile(
	ctx context.Context,
	filename string,
) (records.Root, error) {
	file, errFile := s.fs.Open(filename)
	if errFile != nil {
		return records.Root{}, errors.Join(ErrFailedToOpenFile, errFile)
	}
	defer file.Close()

	s.log.DebugContext(ctx, "opened file:",
		slog.String("filename", filename))

	for {
		record, errRaw := s.recordFromFile(ctx, file)
		if errRaw != nil {
			if errors.Is(errRaw, io.EOF) {
				break // done
			}

			return records.Root{}, errRaw
		}

		if err := s.builder.AddRecord(ctx, record); err != nil {
			return records.Root{}, errors.Join(ErrFailedToAddRecord, err)
		}
	}

	root, err := s.builder.Build()
	if err != nil {
		return records.Root{}, errors.Join(ErrFailedToBuildRoot, err)
	}

	s.log.DebugContext(ctx, "efd records parsed",
		slog.Int("efrms", len(root.EFRMs)),
		slog.Int("eftps", len(root.EFTPs)))

	return root, nil
}

func (s *service) recordFromFile(
	ctx context.Context,
	r io.Reader,
) (records.Raw, error) {
	var magicAndLength [16]byte
	if err := binary.Read(r, binary.LittleEndian, &magicAndLength); err != nil {
		return records.Raw{}, errors.Join(ErrInvalidRecordMagicNumber, err)
	}

	magic := magicAndLength[:4]

	l := binary.LittleEndian.Uint64(magicAndLength[8:16])
	bufLen := l - uint64(len(magicAndLength))
	buf := make([]byte, bufLen)

	_, err := io.ReadFull(r, buf)
	if err != nil {
		return records.Raw{}, errors.Join(ErrFailedToReadRecord, err)
	}

	s.log.DebugContext(ctx, "record read",
		slog.String("magic", string(magic)),
		slog.Uint64("length", l),
	)

	return records.Raw{
		Magic:  [4]byte(magic),
		Length: l,
		Data:   buf,
	}, nil
}
