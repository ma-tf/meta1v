package efd

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	"image/color"
	"io"
	"os"

	"github.com/ma-tf/meta1v/pkg/records"
)

var _ RecordService = &recordservice{}

type RecordService interface {
	RecordsFromFile(*os.File) (records.Root, error)
}

type recordservice struct{}

func NewService() RecordService {
	return &recordservice{}
}

var (
	ErrFailedToReadRecord     = errors.New("failed to read record from file")
	ErrMultipleEFDFRecords    = errors.New("multiple EFDF records found")
	ErrUnknownRecordType      = errors.New("unknown record type")
	ErrFailedToParseThumbnail = errors.New("failed to parse EFTP thumbnail")
)

const bytesPerPixel = 3 // RGB

func (s *recordservice) RecordsFromFile(file *os.File) (records.Root, error) {
	var (
		efdf  *records.EFDF
		efrms []records.EFRM
		eftps []records.EFTP
	)

	for {
		record, errRaw := recordFromFile(file)
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

			r, err := efdfFromRecord(record)
			if err != nil {
				return records.Root{}, err
			}

			efdf = &r
		case "EFRM":
			r, err := efrmFromRecord(record)
			if err != nil {
				return records.Root{}, err
			}

			efrms = append(efrms, r)
		case "EFTP":
			r, err := eftpFromRecord(bytes.NewReader(record.Data))
			if err != nil {
				return records.Root{}, err
			}

			eftps = append(eftps, *r)
		default:
			return records.Root{},
				fmt.Errorf("%w: %s", ErrUnknownRecordType, magic)
		}
	}

	return records.Root{
		EFDF:  *efdf,
		EFRMs: efrms,
		EFTPs: eftps,
	}, nil
}

func recordFromFile(r *os.File) (records.Raw, error) {
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

	return records.Raw{
		Magic:  [4]byte(magic),
		Length: l,
		Data:   buf,
	}, nil
}

func efdfFromRecord(record records.Raw) (records.EFDF, error) {
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

	return r, nil
}

func efrmFromRecord(record records.Raw) (records.EFRM, error) {
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

	return r, nil
}

func eftpFromRecord(r io.Reader) (*records.EFTP, error) {
	order := binary.LittleEndian

	var header [16]byte
	if err := binary.Read(r, order, &header); err != nil {
		return nil, errors.Join(ErrFailedToParseThumbnail, err)
	}

	frameNumber := order.Uint16(header[0:2])
	unknown1 := header[2]
	unknown2 := header[3]

	var filepath [256]byte
	if err := binary.Read(r, order, &filepath); err != nil {
		return nil, errors.Join(ErrFailedToParseThumbnail, err)
	}

	data, err := io.ReadAll(r)
	if err != nil {
		return nil, errors.Join(ErrFailedToParseThumbnail, err)
	}

	w := order.Uint16(header[4:6])
	h := order.Uint16(header[6:8])
	width := int(w)
	height := int(h)

	thumbnail := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := range height {
		for x := range width {
			idx := (y*width + x) * bytesPerPixel
			if idx+2 >= len(data) {
				break
			}

			b := data[idx]
			g := data[idx+1]
			r := data[idx+2]
			thumbnail.Set(x, y, color.RGBA{r, g, b, 255}) // opaque alpha
		}
	}

	return &records.EFTP{
		Index:     frameNumber,
		Unknown1:  unknown1,
		Unknown2:  unknown2,
		Width:     w,
		Height:    h,
		Unknown3:  [8]byte(header[8:16]),
		Filepath:  filepath,
		Thumbnail: thumbnail,
	}, nil
}
