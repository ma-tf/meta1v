package display

import (
	"bytes"
	"fmt"
	"math"
	"os"

	"github.com/ma-tf/meta1v/pkg/records"
	"github.com/qeesung/image2ascii/convert"
)

type Service interface {
	Display() error
}

type displayableRoll struct {
	FilmID         FilmID
	FirstRow       uint
	PerRow         uint
	Title          Title
	FilmLoadedDate DisplayableDatetime
	FrameCount     uint
	IsoDX          Iso
	Remarks        Remarks // film name, location, push/pull, etc.

	Frames []displayableFrame
}

func NewDisplayableRoll(r records.Root) (Service, error) {
	fid, err := NewFilmID(r.EFDF.CodeA, r.EFDF.CodeB)
	if err != nil {
		return displayableRoll{}, err
	}

	filmLoadedDate, err := NewDateTime(r.EFDF.Year, r.EFDF.Month, r.EFDF.Day,
		r.EFDF.Hour, r.EFDF.Minute, r.EFDF.Second)
	if err != nil {
		return displayableRoll{}, err
	}

	thumbnails := make(map[uint16]*DisplayableThumbnail, len(r.EFTPs))
	for _, eftp := range r.EFTPs {
		thumbnail := newDisplayableThumbnail(eftp)

		if thumbnails[eftp.Index] != nil {
			return displayableRoll{}, fmt.Errorf("%w: frame number %d",
				ErrMultipleThumbnailsForFrame, eftp.Index)
		}

		thumbnails[eftp.Index] = &thumbnail
	}

	// r.EFRMs != rr.Exposures ∵ multiple exposures? untested with real world frames
	frames := make([]displayableFrame, 0, len(r.EFRMs))
	for i, frame := range r.EFRMs {
		idx := i + 1
		if idx < 0 || idx > math.MaxUint16 {
			return displayableRoll{},
				fmt.Errorf("%w: index %d", ErrFrameIndexOutOfRange, i+1)
		}

		var pt *DisplayableThumbnail
		if t, ok := thumbnails[uint16(idx)]; ok {
			pt = t
		}

		framePF, errPF := newFrameBuilder(frame, pt).
			WithBasicInfoAndModes().
			WithCameraModesAndFlashInfo().
			WithCustomFunctionsAndFocusPoints().
			Build()
		if errPF != nil {
			return displayableRoll{}, errPF
		}

		frames = append(frames, framePF)
	}

	return displayableRoll{
		FilmID:         fid,
		FirstRow:       uint(r.EFDF.PerRow - r.EFDF.FirstRow),
		PerRow:         uint(r.EFDF.PerRow),
		Title:          NewTitle(r.EFDF.Title),
		FilmLoadedDate: filmLoadedDate,
		FrameCount:     uint(r.EFDF.FrameCount),
		IsoDX:          NewIso(r.EFDF.IsoDX),
		Remarks:        NewRemarks(r.EFDF.Remarks),
		Frames:         frames,
	}, nil
}

func (r displayableRoll) Display() error {
	r.DisplayRoll()

	return r.DisplayFrames()
}

func (r displayableRoll) DisplayRoll() {
	s := fmt.Sprintf(
		"Roll information:\n"+
			" Film ID: %v\n"+
			" Frames in first row: %v\n"+
			" Frames per row: %v\n"+
			" Title: %v\n"+
			" Film load date: %v\n"+
			" Frame count: %v\n"+
			" ISO (DX): %v\n"+
			" Remarks: %v",
		r.FilmID,
		r.FirstRow,
		r.PerRow,
		r.Title,
		r.FilmLoadedDate,
		r.FrameCount,
		r.IsoDX,
		r.Remarks,
	)

	fmt.Fprintln(os.Stdout, s)
}

func (r displayableRoll) DisplayFrames() error {
	for _, f := range r.Frames {
		err := r.displayFrame(f)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r displayableRoll) displayFrame(f displayableFrame) error {
	fmt.Fprintln(os.Stdout, "\nFrame information:")

	err := r.displayFocusPoints(f.FocusingPoints)
	if err != nil {
		return err
	}

	s := fmt.Sprintf(
		" Is user modified record: %v\n"+
			" Frame Number: %v\n"+
			" Film ID: %v\n"+
			" Film loaded at: %v\n"+
			" ISO (DX): %v\n"+
			" Focal Length: %v\n"+
			" Max Aperture: %v\n"+
			" Tv: %v\n"+
			" Av: %v\n"+
			" ISO (M): %v\n"+
			" Exposure Compensation: %v\n"+
			" Flash Exposure Compensation: %v\n"+
			" Flash Mode: %v\n"+
			" Metering Mode: %v\n"+
			" Shooting Mode: %v\n"+
			" Film Advance Mode: %v\n"+
			" AF Mode: %v\n"+
			" Bulb Exposure Time: %v\n"+
			" Taken At: %v\n"+
			" Multiple Exposure: %v\n"+
			" Battery Loaded At: %v",
		f.UserModifiedRecord,
		f.FrameNumber,
		f.FilmID,
		f.FilmLoadedAt,
		f.IsoDX,
		f.FocalLength,
		f.MaxAperture,
		f.Tv,
		f.Av,
		f.IsoM,
		f.ExposureCompensation,
		f.FlashExposureComp,
		f.FlashMode,
		f.MeteringMode,
		f.ShootingMode,
		f.FilmAdvanceMode,
		f.AFMode,
		f.BulbExposureTime,
		f.TakenAt,
		f.MultipleExposure,
		f.BatteryLoadedAt,
	)
	fmt.Fprintln(os.Stdout, s)
	r.displayCustomFunctions(f.CustomFunctions)
	fmt.Fprintln(os.Stdout, " Remarks:", f.Remarks)

	if f.Thumbnail != nil {
		r.displayThumbnail(f.Thumbnail)
	}

	return nil
}

func (r displayableRoll) displayCustomFunctions(cf DisplayableCustomFunctions) {
	table := " Custom Functions:\n" +
		"  #  1  2  3  4  5  6  7  8  9 10 11 12 13 14 15 16 17 18 19 20\n" +
		"     %s  %s  %s  %s  %s  %s  %s  %s  %s  %s  %s  %s  %s  %s  %s  %s  %s  %s  %s  %s\n"

	values := make([]any, len(cf))
	for i, v := range cf {
		values[i] = v
	}

	fmt.Fprintf(os.Stdout, table, values...)
}

func (r displayableRoll) displayFocusPoints(pf DisplayableFocusPoints) error {
	if pf.Selection == math.MaxUint32 {
		empty := " Focusing Points:\n" +
			"      \033[30m\u25AF \u25AF \u25AF \u25AF \u25AF \u25AF \u25AF\n" +
			"   \u25AF \u25AF \u25AF \u25AF \u25AF \u25AF \u25AF \u25AF \u25AF \u25AF\n" +
			"  \u25AF \u25AF \u25AF \u25AF \u25AF \u25AF \u25AF \u25AF \u25AF \u25AF \u25AF\n" +
			"   \u25AF \u25AF \u25AF \u25AF \u25AF \u25AF \u25AF \u25AF \u25AF \u25AF\n" +
			"      \u25AF \u25AF \u25AF \u25AF \u25AF \u25AF \u25AF\033[0m\n"
		fmt.Fprintln(os.Stdout, empty)

		return nil
	}

	p := make([]string, len(fpBits))
	for i, fpBit := range fpBits {
		b, err := byteToBox(pf.Points[i], fpBit)
		if err != nil {
			return err
		}

		p[i] = b
	}

	printableFocusPoints := " Focusing Points:\n" +
		"      " + p[0] + "\n" +
		"   " + p[2] + p[1] + "\n" +
		"  " + p[4] + p[3] + "\n" +
		"   " + p[6] + p[5] + "\n" +
		"      " + p[7] + "\n"

	fmt.Fprintln(os.Stdout, printableFocusPoints)

	return nil
}

func (r displayableRoll) displayThumbnail(t *DisplayableThumbnail) {
	s := "\n Path:" + t.Filepath +
		"\n Thumbnail:\n" + t.Thumbnail
	fmt.Fprintln(os.Stdout, s)
}

type displayableFrame struct {
	FrameNumber  uint
	FilmID       FilmID
	FilmLoadedAt DisplayableDatetime
	IsoDX        Iso

	UserModifiedRecord bool

	FocalLength FocalLength
	MaxAperture Av
	Tv          Tv
	Av          Av
	IsoM        Iso

	ExposureCompensation ExposureCompenation
	FlashExposureComp    ExposureCompenation
	FlashMode            FlashMode
	MeteringMode         MeteringMode
	ShootingMode         ShootingMode

	FilmAdvanceMode  FilmAdvanceMode
	AFMode           AutoFocusMode
	BulbExposureTime BulbExposureTime
	TakenAt          DisplayableDatetime

	MultipleExposure MultipleExposure
	BatteryLoadedAt  DisplayableDatetime

	CustomFunctions DisplayableCustomFunctions
	Remarks         Remarks

	FocusingPoints DisplayableFocusPoints

	Thumbnail *DisplayableThumbnail
}

type DisplayableThumbnail struct {
	Thumbnail string
	Filepath  string
}

func newDisplayableThumbnail(eftp records.EFTP) DisplayableThumbnail {
	filepath := string(eftp.Filepath[:bytes.IndexByte(eftp.Filepath[:], 0)])

	options := convert.DefaultOptions
	options.FixedWidth = int(eftp.Width)

	const heightRatio = 2

	options.FixedHeight = int(eftp.Height / heightRatio)

	ascii := convert.NewImageConverter().
		Image2ASCIIString(eftp.Thumbnail, &options)

	return DisplayableThumbnail{
		Thumbnail: ascii,
		Filepath:  filepath,
	}
}
