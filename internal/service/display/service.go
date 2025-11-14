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

	firstRow, perRow := uint(r.EFDF.FirstRow), uint(r.EFDF.FirstRow-r.EFDF.PerRow)
	title := NewTitle(r.EFDF.Title)

	filmLoadedDate, err := NewDateTime(r.EFDF.Year, r.EFDF.Month, r.EFDF.Day,
		r.EFDF.Hour, r.EFDF.Minute, r.EFDF.Second)
	if err != nil {
		return displayableRoll{}, err
	}

	isoDx := NewIso(r.EFDF.IsoDX)
	remarks := NewRemarks(r.EFDF.Remarks)

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
			return displayableRoll{}, fmt.Errorf("%w: index %d", ErrFrameIndexOutOfRange, i+1)
		}

		var pt *DisplayableThumbnail
		if t, ok := thumbnails[uint16(idx)]; ok {
			pt = t
		}

		framePF, errPF := newDisplayableFrame(frame, pt)
		if errPF != nil {
			return displayableRoll{}, errPF
		}

		frames = append(frames, framePF)
	}

	return displayableRoll{
		FilmID:         fid,
		FirstRow:       firstRow,
		PerRow:         perRow,
		Title:          title,
		FilmLoadedDate: filmLoadedDate,
		FrameCount:     uint(r.EFDF.FrameCount),
		IsoDX:          isoDx,
		Remarks:        remarks,
		Frames:         frames,
	}, nil
}

func (r displayableRoll) Display() error {
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
		f.ExposureCompenation,
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

	ExposureCompenation ExposureCompenation
	FlashExposureComp   ExposureCompenation
	FlashMode           FlashMode
	MeteringMode        MeteringMode
	ShootingMode        ShootingMode

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

//nolint:funlen // necessary complexity
func newDisplayableFrame(r records.EFRM, t *DisplayableThumbnail) (displayableFrame, error) {
	filmID, err := NewFilmID(r.CodeA, r.CodeB)
	if err != nil {
		return displayableFrame{}, err
	}

	filmLoadedDate, err := NewDateTime(r.Year, r.Month, r.Day, r.Hour, r.Minute, r.Second)
	if err != nil {
		return displayableFrame{}, err
	}

	isoDx := NewIso(r.IsoDX)
	focalLength := NewFocalLength(r.FocalLength)

	maxAperture, err := NewAv(r.MaxAperture)
	if err != nil {
		return displayableFrame{}, err
	}

	tv, err := NewTv(r.Tv)
	if err != nil {
		return displayableFrame{}, err
	}

	av, err := NewAv(r.Av)
	if err != nil {
		return displayableFrame{}, err
	}

	isoM := NewIso(r.IsoM)

	exposureCompentation, err := NewExposureCompensation(r.ExposureCompenation)
	if err != nil {
		return displayableFrame{}, err
	}

	flashExpostureComp, err := NewExposureCompensation(r.FlashExposureCompensation)
	if err != nil {
		return displayableFrame{}, err
	}

	flashMode, err := NewFlashMode(r.FlashMode)
	if err != nil {
		return displayableFrame{}, err
	}

	meteringMode, err := NewMeteringMode(r.MeteringMode)
	if err != nil {
		return displayableFrame{}, err
	}

	shootingMode, err := NewShootingMode(r.ShootingMode)
	if err != nil {
		return displayableFrame{}, err
	}

	filmAdvanceMode, err := NewFilmAdvanceMode(r.FilmAdvanceMode)
	if err != nil {
		return displayableFrame{}, err
	}

	afMode, err := NewAutoFocusMode(r.AFMode)
	if err != nil {
		return displayableFrame{}, err
	}

	var bulbExposureTime BulbExposureTime
	if tv == "Bulb" {
		bulbExposureTime, err = NewBulbExposureTime(r.BulbExposureTime)
		if err != nil {
			return displayableFrame{}, err
		}
	}

	takenAt, err := NewDateTime(r.Year, r.Month, r.BatteryDay, r.Hour, r.Minute, r.Second)
	if err != nil {
		return displayableFrame{}, err
	}

	multipleExposure, err := NewMultipleExposure(r.MultipleExposure)
	if err != nil {
		return displayableFrame{}, err
	}

	batteryLoadedAt, err := NewDateTime(r.BatteryYear, r.BatteryMonth, r.BatteryDay,
		r.BatteryHour, r.BatteryMinute, r.BatterySecond)
	if err != nil {
		return displayableFrame{}, err
	}

	cfs, err := NewCustomFunctions(r)
	if err != nil {
		return displayableFrame{}, err
	}

	remarks := NewRemarks(r.Remarks)
	focusPoints := DisplayableFocusPoints{
		Selection: uint(r.FocusingPoint),
		Points: [8]byte{
			r.FocusPoints1,
			r.FocusPoints2,
			r.FocusPoints3,
			r.FocusPoints4,
			r.FocusPoints5,
			r.FocusPoints6,
			r.FocusPoints7,
			r.FocusPoints8,
		},
	}

	return displayableFrame{
		FrameNumber:  uint(r.FrameNumber),
		FilmID:       filmID,
		FilmLoadedAt: filmLoadedDate,
		IsoDX:        isoDx,

		UserModifiedRecord: r.IsModifiedRecord != 0,

		FocalLength: focalLength,
		MaxAperture: maxAperture,
		Tv:          tv,
		Av:          av,
		IsoM:        isoM,

		ExposureCompenation: exposureCompentation,
		FlashExposureComp:   flashExpostureComp,
		FlashMode:           flashMode,
		MeteringMode:        meteringMode,
		ShootingMode:        shootingMode,

		FilmAdvanceMode:  filmAdvanceMode,
		AFMode:           afMode,
		BulbExposureTime: bulbExposureTime,
		TakenAt:          takenAt,

		MultipleExposure: multipleExposure,
		BatteryLoadedAt:  batteryLoadedAt,

		CustomFunctions: cfs,
		Remarks:         remarks,

		FocusingPoints: focusPoints,

		Thumbnail: t,
	}, nil
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

	ascii := convert.NewImageConverter().Image2ASCIIString(eftp.Thumbnail, &options)

	return DisplayableThumbnail{
		Thumbnail: ascii,
		Filepath:  filepath,
	}
}
