//go:generate mockgen -destination=./mocks/service_mock.go -package=efd_test github.com/ma-tf/meta1v/internal/service/efd Service

package efd

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	"image/color"
	"io"
	"log/slog"

	"github.com/ma-tf/meta1v/pkg/records"
)

var _ Service = &service{log: nil}

type Service interface {
	RecordsFromFile(ctx context.Context, r io.Reader) (records.Root, error)
}

type service struct {
	log *slog.Logger
}

func NewService(log *slog.Logger) Service {
	return &service{
		log: log,
	}
}

var (
	ErrFailedToReadRecord     = errors.New("failed to read record from file")
	ErrMultipleEFDFRecords    = errors.New("multiple EFDF records found")
	ErrUnknownRecordType      = errors.New("unknown record type")
	ErrFailedToParseThumbnail = errors.New("failed to parse EFTP thumbnail")
)

const bytesPerPixel = 3 // RGB

func (s *service) RecordsFromFile(
	ctx context.Context,
	r io.Reader,
) (records.Root, error) {
	var (
		efdf  *records.EFDF
		efrms []records.EFRM
		eftps []records.EFTP
	)

	for {
		record, errRaw := s.recordFromFile(ctx, r)
		if errRaw != nil {
			if errors.Is(errRaw, io.EOF) {
				break // done
			}

			return records.Root{}, errRaw
		}

		magic := string(record.Magic[:])
		switch magic {
		case "EFDF":
			if efdf != nil {
				return records.Root{}, ErrMultipleEFDFRecords
			}

			r, err := s.efdfFromRecord(ctx, record)
			if err != nil {
				return records.Root{}, err
			}

			efdf = &r
		case "EFRM":
			r, err := s.efrmFromRecord(ctx, record)
			if err != nil {
				return records.Root{}, err
			}

			efrms = append(efrms, r)
		case "EFTP":
			r, err := s.eftpFromRecord(ctx, record)
			if err != nil {
				return records.Root{}, err
			}

			eftps = append(eftps, *r)
		default:
			return records.Root{},
				fmt.Errorf("%w: %s", ErrUnknownRecordType, magic)
		}
	}

	s.log.DebugContext(ctx, "efd records parsed",
		slog.Int("efrms", len(efrms)),
		slog.Int("eftps", len(eftps)))

	return records.Root{
		EFDF:  *efdf,
		EFRMs: efrms,
		EFTPs: eftps,
	}, nil
}

func (s *service) recordFromFile(
	ctx context.Context,
	r io.Reader,
) (records.Raw, error) {
	var magicAndLength [16]byte
	if err := binary.Read(r, binary.LittleEndian, &magicAndLength); err != nil {
		return records.Raw{}, errors.Join(ErrFailedToReadRecord, err)
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

func (s *service) efdfFromRecord(
	ctx context.Context,
	record records.Raw,
) (records.EFDF, error) {
	var r records.EFDF

	err := binary.Read(
		bytes.NewReader(record.Data),
		binary.LittleEndian,
		&r,
	)
	if err != nil {
		return records.EFDF{}, fmt.Errorf(
			"failed to parse EFDF record: %w",
			err,
		)
	}

	s.log.DebugContext(ctx, "efdf parsed",
		slog.Uint64("frameCount", uint64(r.FrameCount)),
		slog.Int("length", len(record.Data)),
	)

	return r, nil
}

func (s *service) efrmFromRecord(
	ctx context.Context,
	record records.Raw,
) (records.EFRM, error) {
	var r records.EFRM

	err := binary.Read(
		bytes.NewReader(record.Data),
		binary.LittleEndian,
		&r,
	)
	if err != nil {
		return records.EFRM{}, fmt.Errorf(
			"failed to parse EFRM record: %w",
			err,
		)
	}

	s.log.DebugContext(ctx, "efrm parsed",
		slog.Int("length", len(record.Data)),
		slog.Int("frameNumber", int(r.FrameNumber)),
	)

	return r, nil
}

func (s *service) eftpFromRecord(
	ctx context.Context,
	record records.Raw,
) (*records.EFTP, error) {
	var (
		order  = binary.LittleEndian
		header [16]byte
	)

	r := bytes.NewReader(record.Data)
	if err := binary.Read(r, order, &header); err != nil {
		return nil, errors.Join(ErrFailedToParseThumbnail, err)
	}

	var filepath [256]byte
	if err := binary.Read(r, order, &filepath); err != nil {
		return nil, errors.Join(ErrFailedToParseThumbnail, err)
	}

	data, err := io.ReadAll(r)
	if err != nil {
		return nil, errors.Join(ErrFailedToParseThumbnail, err)
	}

	w, h := order.Uint16(header[4:6]), order.Uint16(header[6:8])
	width, height := int(w), int(h)

	thumbnail := image.NewRGBA(image.Rect(0, 0, width, height))
	// for y := range height {
	// 	for x := range width {
	// 		idx := (y*width + x) * bytesPerPixel
	// 		if idx+2 >= len(data) {
	// 			break
	// 		}

	// 		thumbnail.Set(x, y, color.RGBA{data[idx+2], data[idx+1], data[idx], 255}) // opaque alpha
	// 	}
	// }

	for i := 0; i+2 < len(data); i += bytesPerPixel {
		idx := i / bytesPerPixel
		thumbnail.SetRGBA(
			idx%width,
			idx/height,
			color.RGBA{
				data[idx+2], data[idx+1], data[idx], 255, // opaque alpha
			},
		)
	}

	thumbnailSize := height * width * bytesPerPixel
	frameNumber := order.Uint16(header[0:2])
	s.log.DebugContext(ctx, "eftp parsed",
		slog.Int("length", thumbnailSize),
		slog.Int("frameNumber", int(frameNumber)),
	)

	return &records.EFTP{
		Index:     frameNumber,
		Unknown1:  header[2],
		Unknown2:  header[3],
		Width:     w,
		Height:    h,
		Unknown3:  [8]byte(header[8:16]),
		Filepath:  filepath,
		Thumbnail: thumbnail,
	}, nil
}
