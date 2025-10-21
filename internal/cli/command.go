package cli

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"

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
		// Uncomment the following line if your bare application
		// has an action associated with it:
		RunE: func(_ *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("filename must be specified")
			}

			file, err := os.Open(args[0])
			if err != nil {
				return err
			}
			defer file.Close()

			var records []Record
			for {
				data, err := readLenPrefixed(file, binary.LittleEndian)
				if err != nil {
					if errors.Is(err, io.EOF) {
						break // done
					}
					// if err == io.ErrUnexpectedEOF -> truncated -> handle
					return err
				}
				// process data
				records = append(records, data)
			}

			//fmt.Fprintln(os.Stdout, "records count:", len(records))

			var EFDF EFDFRecord
			for _, rec := range records {
				if string(rec.Magic[:]) == "EFDF" {
					err := binary.Read(
						bytes.NewReader(rec.Data),
						binary.LittleEndian,
						&EFDF,
					)
					if err != nil {
						return fmt.Errorf("failed to parse EFDF record: %w", err)
					}
					break
				}
			}

			EFRMs := make([]EFRMRecord, 0, len(records))
			for _, rec := range records {
				if string(rec.Magic[:]) == "EFRM" {
					var efrm EFRMRecord
					err := binary.Read(
						bytes.NewReader(rec.Data),
						binary.LittleEndian,
						&efrm,
					)
					if err != nil {
						return fmt.Errorf("failed to parse EFRM record: %w", err)
					}
					EFRMs = append(EFRMs, efrm)
				}
			}

			fmt.Fprintln(os.Stdout, "EFDF Record:")
			fmt.Fprintln(os.Stdout, " Film ID:", fmt.Sprintf("%d-%d", EFDF.CodeA, EFDF.CodeB))
			fmt.Fprintln(os.Stdout, " Frames in first row:", (EFDF.PerRow - EFDF.FirstRow))
			fmt.Fprintln(os.Stdout, " Frames per row:", EFDF.PerRow)
			if EFDF.Title[0] == '\x00' {
				fmt.Fprintln(os.Stdout, " Title: <empty>")
			} else {
				fmt.Fprintln(os.Stdout, " Title:", string(bytes.TrimRight(EFDF.Title[:], "\x00")))
			}
			fmt.Fprintln(os.Stdout, " Film load date:", fmt.Sprintf("%04d-%02d-%02d %02d:%02d:%02d",
				EFDF.Year, EFDF.Month, EFDF.Day,
				EFDF.Hour, EFDF.Minute, EFDF.Second))
			fmt.Fprintln(os.Stdout, " Exposures:", EFDF.Exposures)
			fmt.Fprintln(os.Stdout, " ISO DX:", EFDF.IsoDX)
			if EFDF.Remarks[0] == '\x00' {
				fmt.Fprintln(os.Stdout, " Remarks: <empty>")
			} else {
				fmt.Fprintln(os.Stdout, " Remarks:", string(bytes.TrimRight(EFDF.Remarks[:], "\x00")))
			}

			fmt.Fprintln(os.Stdout, "\nEFRM Records:")
			for i, efrm := range EFRMs {
				fmt.Fprintln(os.Stdout, "Frame", i+1)
				fmt.Fprintln(os.Stdout, " Frame Number:", efrm.FrameNumber)
				fmt.Fprintln(os.Stdout, " Focal Length:", fmt.Sprintf("%dmm", efrm.FocalLength))
				fmt.Fprintln(os.Stdout, " Max Aperture:", fmt.Sprintf("f/%.1f", float32(efrm.MaxAperture)/100.0))
				tv := (-float32(efrm.Tv) / 100.0)
				if tv >= 0 {
					fmt.Fprintln(os.Stdout, " Tv:", fmt.Sprintf("1/%.0f", tv))
				} else {
					fmt.Fprintln(os.Stdout, " Tv:", fmt.Sprintf("%.0f\"", -tv))
				}
				fmt.Fprintln(os.Stdout, " Av:", fmt.Sprintf("%.1f", float32(efrm.Av)/100.0))
				fmt.Fprintln(os.Stdout, " Iso (M):", efrm.IsoM)
				fmt.Fprintln(os.Stdout, " Capture date:", fmt.Sprintf("%04d-%02d-%02d %02d:%02d:%02d",
					efrm.Year, efrm.Month, efrm.Day,
					efrm.Hour, efrm.Minute, efrm.Second))
				fmt.Fprintln(os.Stdout, " Unknown fields:", efrm.Unknown7, efrm.Unknown8)
				fmt.Fprintln(os.Stdout, " Exposure compensation:", fmt.Sprintf("%.1f", float32(efrm.ExposureCompenation)/100.0))
				fmt.Fprintln(os.Stdout, " Flash exposure compensation:", fmt.Sprintf("%.1f", float32(efrm.FlashExposureCompenation)/100.0))
				fmt.Fprintln(os.Stdout, " Flash mode:", efrm.FlashMode)
				fmt.Fprintln(os.Stdout, " Metering mode:", efrm.MeteringMode)
				fmt.Fprintln(os.Stdout, " Shooting mode:", efrm.ShootingMode)
				fmt.Fprintln(os.Stdout, " Film advance mode:", efrm.FilmAdvanceMode)
				fmt.Fprintln(os.Stdout, " AF mode:", efrm.AFMode)
				fmt.Fprintln(os.Stdout, " Bulb exposure time:")
				fmt.Fprintln(os.Stdout, " Multiple exposure:", efrm.MultipleExposure)
				fmt.Fprintln(os.Stdout, " Battery loaded date & time:")
				fmt.Fprintln(os.Stdout, " Remarks:")
				fmt.Fprintln(os.Stdout)
			}

			return nil
		},
	}
}

type Record struct {
	Magic  [4]byte
	Length uint64
	Data   []byte
}

// EFDFRecord should be 512-8 bytes.
type EFDFRecord struct {
	Unknown1  [8]byte
	Unknown2  [8]byte
	Unknown3  [4]byte
	Unknown4  [2]byte
	CodeB     uint32
	Year      uint16
	Month     uint8
	Day       uint8
	Hour      uint8
	Minute    uint8
	Second    uint8
	Unknown5  [1]byte
	Exposures uint32
	IsoDX     uint32
	CodeA     uint32
	FirstRow  uint8
	PerRow    uint8
	Unknown6  [128]byte
	Title     [64]byte
	Remarks   [256]byte
}

type EFRMRecord struct {
	Unknown1                 [4]byte
	Unknown2                 [4]byte
	FrameNumber              uint32
	FocalLength              uint32
	MaxAperture              uint32
	Tv                       int32
	Av                       uint32
	IsoM                     uint32
	ExposureCompenation      int32
	FlashExposureCompenation int32
	Year                     uint16
	Month                    uint8
	Day                      uint8
	Hour                     uint8
	Minute                   uint8
	Second                   uint8
	Unknown4                 [1]byte
	FlashMode                uint32
	FilmAdvanceMode          uint32
	MultipleExposure         uint32
	Unknown7                 uint32
	Unknown8                 uint32
	MeteringMode             uint32
	ShootingMode             uint32
	AFMode                   uint32
	Unknown12                [48]byte
	Unknown13                [16]byte
	Unknown14                [16]byte
	Unknown15                [16]byte
	Unknown16                [64]byte
	Remarks                  [256]byte
}

const maxLen = 1 * 1024 * 1024 // 1 MB
func readLenPrefixed(r io.Reader, order binary.ByteOrder) (Record, error) {
	var magicAndLength [16]byte
	if err := binary.Read(r, order, &magicAndLength); err != nil {
		return Record{}, err
	}
	magic := magicAndLength[:4]
	//fmt.Fprintln(os.Stdout, "magic: ", string(magic))
	// v := binary.LittleEndian.Uint32(magicAndLength[4:8])
	// println("version number:", v)
	l := binary.LittleEndian.Uint64(magicAndLength[8:16])
	//fmt.Fprintln(os.Stdout, "length:", l)
	if uint(l) > maxLen {
		return Record{}, fmt.Errorf("length %d exceeds max %d", l, maxLen)
	}
	bufLen := l - uint64(len(magicAndLength))
	buf := make([]byte, bufLen)
	_, err := io.ReadFull(r, buf)
	if err != nil {
		return Record{}, err
	}
	return Record{
		Magic:  [4]byte(magic),
		Length: l,
		Data:   buf,
	}, nil
}
