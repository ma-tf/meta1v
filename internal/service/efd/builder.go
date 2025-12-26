//go:generate mockgen -destination=./mocks/builder_mock.go -package=efd_test github.com/ma-tf/meta1v/internal/service/efd RootBuilder
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

var ErrMissingEFDFRecord = errors.New("missing EFDF record")

type RootBuilder interface {
	AddRecord(ctx context.Context, r records.Raw) error
	Build() (records.Root, error)
}

type rootBuilder struct {
	log   *slog.Logger
	efdf  *records.EFDF
	efrms []records.EFRM
	eftps []records.EFTP
}

func NewRootBuilder(log *slog.Logger) RootBuilder {
	var (
		efrms []records.EFRM
		eftps []records.EFTP
	)

	efrms = make([]records.EFRM, 0)
	eftps = make([]records.EFTP, 0)

	return &rootBuilder{
		log:   log,
		efdf:  nil,
		efrms: efrms,
		eftps: eftps,
	}
}

func (b *rootBuilder) AddRecord(ctx context.Context, r records.Raw) error {
	magic := string(r.Magic[:])
	switch magic {
	case "EFDF":
		return b.addEFDF(ctx, r)
	case "EFRM":
		return b.addEFRM(ctx, r)
	case "EFTP":
		return b.addEFTP(ctx, r)
	default:
		return fmt.Errorf("%w: %s", ErrUnknownRecordType, magic)
	}
}

func (b *rootBuilder) Build() (records.Root, error) {
	if b.efdf == nil {
		return records.Root{}, ErrMissingEFDFRecord
	}

	return records.Root{
		EFDF:  *b.efdf,
		EFRMs: b.efrms,
		EFTPs: b.eftps,
	}, nil
}

func (b *rootBuilder) addEFDF(ctx context.Context, record records.Raw) error {
	if b.efdf != nil {
		return ErrMultipleEFDFRecords
	}

	var r records.EFDF
	if err := binary.Read(bytes.NewReader(record.Data), binary.LittleEndian, &r); err != nil {
		return fmt.Errorf("failed to parse EFDF record: %w", err)
	}

	b.log.DebugContext(ctx, "efdf parsed",
		slog.Uint64("frameCount", uint64(r.FrameCount)),
		slog.Int("length", len(record.Data)))

	b.efdf = &r

	return nil
}

func (b *rootBuilder) addEFRM(ctx context.Context, record records.Raw) error {
	var r records.EFRM

	if err := binary.Read(bytes.NewReader(record.Data), binary.LittleEndian, &r); err != nil {
		return fmt.Errorf("failed to parse EFRM record: %w", err)
	}

	b.log.DebugContext(ctx, "efrm parsed",
		slog.Int("length", len(record.Data)),
		slog.Int("frameNumber", int(r.FrameNumber)))

	b.efrms = append(b.efrms, r)

	return nil
}

func (b *rootBuilder) addEFTP(ctx context.Context, record records.Raw) error {
	const bytesPerPixel = 3

	var (
		order  = binary.LittleEndian
		header [16]byte
	)

	r := bytes.NewReader(record.Data)
	if err := binary.Read(r, order, &header); err != nil {
		return errors.Join(ErrFailedToParseThumbnail, err)
	}

	var filepath [256]byte
	if err := binary.Read(r, order, &filepath); err != nil {
		return errors.Join(ErrFailedToParseThumbnail, err)
	}

	data, err := io.ReadAll(r)
	if err != nil {
		return errors.Join(ErrFailedToParseThumbnail, err)
	}

	w, h := order.Uint16(header[4:6]), order.Uint16(header[6:8])
	width, height := int(w), int(h)

	thumbnail := image.NewRGBA(image.Rect(0, 0, width, height))

	for i := 0; i+2 < len(data); i += bytesPerPixel {
		idx := i / bytesPerPixel
		thumbnail.SetRGBA(idx%width, idx/height,
			color.RGBA{data[idx+2], data[idx+1], data[idx], 255})
	}

	frameNumber := order.Uint16(header[0:2])
	b.log.DebugContext(ctx, "eftp parsed",
		slog.Int("length", height*width*bytesPerPixel),
		slog.Int("frameNumber", int(frameNumber)))

	b.eftps = append(b.eftps, records.EFTP{
		Index:     frameNumber,
		Unknown1:  header[2],
		Unknown2:  header[3],
		Width:     w,
		Height:    h,
		Unknown3:  [8]byte(header[8:16]),
		Filepath:  filepath,
		Thumbnail: thumbnail,
	})

	return nil
}
