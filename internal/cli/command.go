package cli

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	"image/color"
	"io"
	"os"
	"slices"

	"github.com/qeesung/image2ascii/convert"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "meta1v",
		Short: "A brief description of your application",
		Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
		RunE: func(_ *cobra.Command, args []string) error {
			if len(args) != 1 {
				return ErrNoFilenameProvided
			}

			file, err := os.Open(args[0])
			if err != nil {
				return errors.Join(ErrFailedToOpenFile, err)
			}
			defer file.Close()

			records, err := readRecords(file)
			if err != nil {
				return err
			}

			printEFDF(records.EFDF)
			printEFRMs(records.EFRMs)
			printEFTPs(records.EFTPs)

			return nil
		},
	}
}

const zeroByte = 0x00

func printEFTPs(eftps []EFTPRecord) {
	fmt.Fprintln(os.Stdout, "EFTP Records:")

	for _, eftp := range eftps {
		fmt.Fprintln(os.Stdout, " Filepath:", string(eftp.Filepath[:bytes.IndexByte(eftp.Filepath[:], zeroByte)]))
		fmt.Fprintln(os.Stdout, " Image:")

		options := convert.DefaultOptions
		options.FixedWidth = int(eftp.Width)

		const heightRatio = 2

		options.FixedHeight = int(eftp.Height / heightRatio)
		options.Colored = true
		ascii := convert.NewImageConverter().Image2ASCIIString(eftp.Thumbnail, &options)
		fmt.Fprintln(os.Stdout, ascii)
	}
}

const (
	maxAvDenominator        = 100
	tvDenominator           = 100
	avDenominator           = 100
	expCompDenominator      = 100
	flashExpCompDenominator = 100
)

func printEFRMs(efrms []EFRMRecord) {
	fmt.Fprintln(os.Stdout, "\nEFRM Records:")

	for i, efrm := range efrms {
		fmt.Fprintln(os.Stdout, "Frame", i+1)
		fmt.Fprintln(os.Stdout, " Frame Number:", efrm.FrameNumber)
		fmt.Fprintln(os.Stdout, " Focal Length:", fmt.Sprintf("%dmm", efrm.FocalLength))
		fmt.Fprintln(os.Stdout, " Max Aperture:", fmt.Sprintf("f/%.1f", float32(efrm.MaxAperture)/maxAvDenominator))

		tv := (-float32(efrm.Tv) / tvDenominator)
		if tv >= 0 {
			fmt.Fprintln(os.Stdout, " Tv:", fmt.Sprintf("1/%.0f", tv))
		} else {
			fmt.Fprintln(os.Stdout, " Tv:", fmt.Sprintf("%.0f\"", -tv))
		}

		fmt.Fprintln(os.Stdout, " Av:", fmt.Sprintf("%.1f", float32(efrm.Av)/avDenominator))
		fmt.Fprintln(os.Stdout, " Iso (M):", efrm.IsoM)
		fmt.Fprintln(os.Stdout, " Capture date:", fmt.Sprintf("%04d-%02d-%02d %02d:%02d:%02d",
			efrm.Year, efrm.Month, efrm.Day,
			efrm.Hour, efrm.Minute, efrm.Second))
		fmt.Fprintln(os.Stdout, " Unknown fields:", efrm.Unknown4, efrm.Unknown5)
		fmt.Fprintln(os.Stdout, " Exposure compensation:", fmt.Sprintf("%.1f",
			float32(efrm.ExposureCompenation)/expCompDenominator))
		fmt.Fprintln(os.Stdout, " Flash exposure compensation:",
			fmt.Sprintf("%.1f", float32(efrm.FlashExposureCompensation)/flashExpCompDenominator))
		fmt.Fprintln(os.Stdout, " Flash mode:", efrm.FlashMode)
		fmt.Fprintln(os.Stdout, " Metering mode:", efrm.MeteringMode)
		fmt.Fprintln(os.Stdout, " Shooting mode:", efrm.ShootingMode)
		fmt.Fprintln(os.Stdout, " Film advance mode:", efrm.FilmAdvanceMode)
		fmt.Fprintln(os.Stdout, " AF mode:", efrm.AFMode)
		fmt.Fprintln(os.Stdout, " Custom functions:", efrm.CustomFunction0, efrm.CustomFunction1, efrm.CustomFunction2,
			efrm.CustomFunction3,
			efrm.CustomFunction4, efrm.CustomFunction5, efrm.CustomFunction6,
			efrm.CustomFunction7, efrm.CustomFunction8, efrm.CustomFunction9,
			efrm.CustomFunction10, efrm.CustomFunction11, efrm.CustomFunction12,
			efrm.CustomFunction13, efrm.CustomFunction14, efrm.CustomFunction15,
			efrm.CustomFunction16, efrm.CustomFunction17, efrm.CustomFunction18,
			efrm.CustomFunction19)
		fmt.Fprintln(os.Stdout, " Bulb exposure time:")
		fmt.Fprintln(os.Stdout, " Multiple exposure:", efrm.MultipleExposure)
		fmt.Fprintln(os.Stdout, " Battery loaded date & time:", fmt.Sprintf("%04d-%02d-%02d %02d:%02d:%02d",
			efrm.BatteryYear, efrm.BatteryMonth, efrm.BatteryDay,
			efrm.BatteryHour, efrm.BatteryMinute, efrm.BatterySecond))
		fmt.Fprintln(os.Stdout, " Focusing point:", efrm.FocusingPoint)

		if efrm.Remarks[0] == '\x00' {
			fmt.Fprintln(os.Stdout, " Remarks: <empty>")
		} else {
			fmt.Fprintln(os.Stdout, " Remarks:", string(bytes.TrimRight(efrm.Remarks[:], "\x00")))
		}

		fmt.Fprintln(os.Stdout, " Frame Film ID:", fmt.Sprintf("%d-%d", efrm.CodeA, efrm.CodeB))
		fmt.Fprintln(os.Stdout, " Frame Roll date:", fmt.Sprintf("%04d-%02d-%02d %02d:%02d:%02d",
			efrm.RollYear, efrm.RollMonth, efrm.RollDay,
			efrm.RollHour, efrm.RollMinute, efrm.RollSecond))
		fmt.Fprintln(os.Stdout, " Roll ISO DX:", efrm.IsoDX)

		fmt.Fprintln(os.Stdout)
	}
}

func printEFDF(efdf EFDFRecord) {
	fmt.Fprintln(os.Stdout, "EFDF Record:")
	fmt.Fprintln(os.Stdout, " Film ID:", fmt.Sprintf("%d-%d", efdf.CodeA, efdf.CodeB))
	fmt.Fprintln(os.Stdout, " Frames in first row:", (efdf.PerRow - efdf.FirstRow))
	fmt.Fprintln(os.Stdout, " Frames per row:", efdf.PerRow)

	if efdf.Title[0] == '\x00' {
		fmt.Fprintln(os.Stdout, " Title: <empty>")
	} else {
		fmt.Fprintln(os.Stdout, " Title:", string(efdf.Title[:bytes.IndexByte(efdf.Title[:], zeroByte)]))
	}

	fmt.Fprintln(os.Stdout, " Film load date:", fmt.Sprintf("%04d-%02d-%02d %02d:%02d:%02d",
		efdf.Year, efdf.Month, efdf.Day,
		efdf.Hour, efdf.Minute, efdf.Second))
	fmt.Fprintln(os.Stdout, " Exposures:", efdf.Exposures)
	fmt.Fprintln(os.Stdout, " ISO DX:", efdf.IsoDX)

	if efdf.Remarks[0] == '\x00' {
		fmt.Fprintln(os.Stdout, " Remarks: <empty>")
	} else {
		fmt.Fprintln(os.Stdout, " Remarks:", string(efdf.Remarks[:bytes.IndexByte(efdf.Remarks[:], zeroByte)]))
	}
}

func readLenPrefixed(r io.Reader, order binary.ByteOrder) (Record, error) {
	var magicAndLength [16]byte
	if err := binary.Read(r, order, &magicAndLength); err != nil {
		return Record{}, errors.Join(ErrFailedToReadRecord, err)
	}

	magic := magicAndLength[:4]

	l := binary.LittleEndian.Uint64(magicAndLength[8:16])

	bufLen := l - uint64(len(magicAndLength))
	buf := make([]byte, bufLen)

	_, err := io.ReadFull(r, buf)
	if err != nil {
		return Record{}, errors.Join(ErrFailedToReadRecord, err)
	}

	return Record{
		Magic:  [4]byte(magic),
		Length: l,
		Data:   buf,
	}, nil
}

func readRecords(file *os.File) (Records, error) {
	var (
		efdf  *EFDFRecord
		efrms []EFRMRecord
		eftps []EFTPRecord
	)

	for {
		record, errRaw := readLenPrefixed(file, binary.LittleEndian)
		if errRaw != nil {
			if errors.Is(errRaw, io.EOF) {
				break // done
			}

			return Records{}, errRaw
		}

		magic := string(record.Magic[:])
		switch magic {
		case "EFDF":
			if efdf != nil {
				return Records{}, ErrMultipleEFDFRecords
			}

			var r EFDFRecord

			err := binary.Read(
				bytes.NewReader(record.Data),
				binary.LittleEndian,
				&r,
			)
			if err != nil {
				return Records{}, fmt.Errorf("failed to parse %s record: %w", magic, err)
			}

			efdf = &r
		case "EFRM":
			var r EFRMRecord

			err := binary.Read(
				bytes.NewReader(record.Data),
				binary.LittleEndian,
				&r,
			)
			if err != nil {
				return Records{}, fmt.Errorf("failed to parse %s record: %w", magic, err)
			}

			efrms = append(efrms, r)
		case "EFTP":
			r, err := parseThumbnail(bytes.NewReader(record.Data), binary.LittleEndian)
			if err != nil {
				return Records{}, err
			}

			eftps = append(eftps, *r)
		default:
			return Records{}, fmt.Errorf("%w: %s", ErrUnknownRecordType, magic)
		}
	}

	return Records{
		EFDF:  *efdf,
		EFRMs: efrms,
		EFTPs: eftps,
	}, nil
}

const bytesPerPixel = 3 // RGB
func parseThumbnail(r io.Reader, order binary.ByteOrder) (*EFTPRecord, error) {
	var header [16]byte
	if err := binary.Read(r, order, &header); err != nil {
		return nil, errors.Join(ErrFailedToParseThumbnail, err)
	}

	var filepath [256]byte
	if err := binary.Read(r, order, &filepath); err != nil {
		return nil, errors.Join(ErrFailedToParseThumbnail, err)
	}

	paddedData, err := io.ReadAll(r)
	if err != nil {
		return nil, errors.Join(ErrFailedToParseThumbnail, err)
	}

	var data []byte

	zi := slices.Index(paddedData, zeroByte)
	if zi != -1 {
		data = paddedData[:zi]
	} else {
		data = paddedData
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

	return &EFTPRecord{
		Unknown1:  [4]byte(header[0:4]),
		Width:     w,
		Height:    h,
		Unknown2:  [8]byte(header[8:16]),
		Filepath:  filepath,
		Thumbnail: thumbnail,
	}, nil
}
