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

//go:generate mockgen -destination=./mocks/reader_mock.go -package=efd_test github.com/ma-tf/meta1v/internal/service/efd Reader
package efd

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"image"
	"image/color"
	"io"
	"log/slog"

	"github.com/ma-tf/meta1v/internal/records"
)

var (
	ErrFailedToReadEFDF = errors.New("failed to read EFDF record")
	ErrFailedToReadEFRM = errors.New("failed to read EFRM record")
)

// Reader provides low-level binary reading operations for EFD file records.
type Reader interface {
	// ReadRaw reads the next raw record (magic bytes + length + data) from the input stream.
	ReadRaw(ctx context.Context, r io.Reader) (records.Raw, error)

	// ReadEFDF parses EFDF (film roll metadata) from raw bytes.
	ReadEFDF(ctx context.Context, data []byte) (records.EFDF, error)

	// ReadEFRM parses EFRM (frame metadata) from raw bytes.
	ReadEFRM(ctx context.Context, data []byte) (records.EFRM, error)

	// ReadEFTP parses EFTP (thumbnail image) from raw bytes and decodes the RGB image data.
	ReadEFTP(ctx context.Context, data []byte) (records.EFTP, error)
}

type reader struct {
	log              *slog.Logger
	thumbnailFactory records.ThumbnailFactory
}

func NewReader(
	log *slog.Logger,
	thumbnailFactory records.ThumbnailFactory,
) Reader {
	return &reader{
		log:              log,
		thumbnailFactory: thumbnailFactory,
	}
}

// ReadRaw reads the next raw record (magic bytes + length + data) from the input stream.
func (b *reader) ReadRaw(
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

	b.log.DebugContext(ctx, "parsed raw record",
		slog.String("magic", string(magic)),
		slog.Uint64("length", l),
	)

	return records.Raw{
		Magic:  [4]byte(magic),
		Length: l,
		Data:   buf,
	}, nil
}

// ReadEFDF parses EFDF (film roll metadata) from raw bytes.
func (b *reader) ReadEFDF(
	ctx context.Context,
	data []byte,
) (records.EFDF, error) {
	var efdf records.EFDF
	if err := binary.Read(
		bytes.NewReader(data),
		binary.LittleEndian,
		&efdf,
	); err != nil {
		return records.EFDF{}, errors.Join(ErrFailedToReadEFDF, err)
	}

	b.log.DebugContext(ctx, "parsed EFDF record")

	return efdf, nil
}

// ReadEFRM parses EFRM (frame metadata) from raw bytes.
func (b *reader) ReadEFRM(
	ctx context.Context,
	data []byte,
) (records.EFRM, error) {
	var efrm records.EFRM
	if err := binary.Read(
		bytes.NewReader(data),
		binary.LittleEndian,
		&efrm,
	); err != nil {
		return records.EFRM{}, errors.Join(ErrFailedToReadEFRM, err)
	}

	b.log.DebugContext(ctx, "parsed EFRM record")

	return efrm, nil
}

// ReadEFTP parses EFTP (thumbnail image) from raw bytes and decodes the RGB image data.
func (b *reader) ReadEFTP(
	ctx context.Context,
	data []byte,
) (records.EFTP, error) {
	const bytesPerPixel = 3

	var (
		order    = binary.LittleEndian
		header   [16]byte
		filepath [256]byte
		r        = bytes.NewReader(data)
	)

	if err := binary.Read(r, order, &header); err != nil {
		return records.EFTP{}, errors.Join(ErrFailedToParseThumbnail, err)
	}

	if err := binary.Read(r, order, &filepath); err != nil {
		return records.EFTP{}, errors.Join(ErrFailedToParseThumbnail, err)
	}

	w, h := order.Uint16(header[4:6]), order.Uint16(header[6:8])
	width, height := int(w), int(h)

	thumbnail := b.thumbnailFactory.NewRGBA(image.Rect(0, 0, width, height))

	for i := 0; i+2 < len(data); i += bytesPerPixel {
		idx := i / bytesPerPixel
		thumbnail.SetRGBA(idx%width, idx/height,
			color.RGBA{data[idx+2], data[idx+1], data[idx], 255})
	}

	frameNumber := order.Uint16(header[0:2])

	b.log.DebugContext(ctx, "parsed EFTP record",
		slog.Uint64("frame_number", uint64(frameNumber)),
		slog.Int("width", width),
		slog.Int("height", height),
	)

	return records.EFTP{
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
