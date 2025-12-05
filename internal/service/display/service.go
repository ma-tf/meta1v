package display

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/ma-tf/meta1v/pkg/domain"
	"github.com/ma-tf/meta1v/pkg/records"
	"github.com/qeesung/image2ascii/convert"
)

const (
	filmIDWidth       = 8
	firstRowWidth     = 9
	perRowWidth       = 14
	titleWidth        = 20
	filmLoadedAtWidth = 20
	frameCountWidth   = 11
	isoDxWidth        = 8
	remarksWidth      = 30

	frameNumberWidth       = 9
	focusingPointsWidth    = 21
	focusingPointsPadding  = filmIDWidth + frameNumberWidth + 2
	focalLengthWidth       = 12
	maxApertureWidth       = 12
	tvWidth                = 7
	avWidth                = 7
	isoMWidth              = 7
	exposureCompWidth      = 15
	flashExposureCompWidth = 20
	flashModeWidth         = 15
	meteringModeWidth      = 15
	shootingModeWidth      = 15
	filmAdvanceModeWidth   = 17
	afModeWidth            = 12
	bulbExposureTimeWidth  = 20
	takenAtWidth           = 20
	multipleExposureWidth  = 20
	batteryLoadedAtWidth   = 20
	customFunctionsWidth   = 2

	imageFileWidth   = 64
	thumbnailWidth   = 64
	thumbnailPadding = filmIDWidth + frameNumberWidth + imageFileWidth + 3
)

var (
	ErrMultipleThumbnailsForFrame = errors.New("frame has multiple thumbnails")
	ErrFrameIndexOutOfRange       = errors.New("frame index out of range")
)

type Service interface {
	DisplayRoll()
	DisplayCustomFunctions()
	DisplayFocusingPoints() error
	DisplayFrames() error
	DisplayThumbnails()
}

type displayableRoll struct {
	FilmID         domain.FilmID
	FirstRow       uint
	PerRow         uint
	Title          domain.Title
	FilmLoadedDate domain.ValidatedDatetime
	FrameCount     uint
	IsoDX          domain.Iso
	Remarks        domain.Remarks // film name, location, push/pull, etc.

	Frames []DisplayableFrame
}

func NewDisplayableRoll(r records.Root) (Service, error) {
	fid, err := domain.NewFilmID(r.EFDF.CodeA, r.EFDF.CodeB)
	if err != nil {
		return displayableRoll{},
			fmt.Errorf("failed to parse roll data: %w", err)
	}

	filmLoadedDate, err := domain.NewDateTime(
		r.EFDF.Year, r.EFDF.Month, r.EFDF.Day,
		r.EFDF.Hour, r.EFDF.Minute, r.EFDF.Second)
	if err != nil {
		return displayableRoll{},
			fmt.Errorf("failed to parse roll data: %w", err)
	}

	thumbnails, err := getThumbnails(r)
	if err != nil {
		return nil, err
	}

	// r.EFRMs != rr.Exposures ∵ multiple exposures? untested with real world frames
	frames, err := getFrames(r, thumbnails)
	if err != nil {
		return nil, err
	}

	return displayableRoll{
		FilmID:         fid,
		FirstRow:       uint(r.EFDF.PerRow - r.EFDF.FirstRow),
		PerRow:         uint(r.EFDF.PerRow),
		Title:          domain.NewTitle(r.EFDF.Title),
		FilmLoadedDate: filmLoadedDate,
		FrameCount:     uint(r.EFDF.FrameCount),
		IsoDX:          domain.NewIso(r.EFDF.IsoDX),
		Remarks:        domain.NewRemarks(r.EFDF.Remarks),
		Frames:         frames,
	}, nil
}

func getThumbnails(r records.Root) (map[uint16]*DisplayableThumbnail, error) {
	thumbnails := make(map[uint16]*DisplayableThumbnail, len(r.EFTPs))
	for _, eftp := range r.EFTPs {
		thumbnail := newDisplayableThumbnail(eftp)

		if thumbnails[eftp.Index] != nil {
			return nil, fmt.Errorf("%w: frame number %d",
				ErrMultipleThumbnailsForFrame, eftp.Index)
		}

		thumbnails[eftp.Index] = &thumbnail
	}

	return thumbnails, nil
}

func getFrames(
	r records.Root,
	thumbnails map[uint16]*DisplayableThumbnail,
) ([]DisplayableFrame, error) {
	frames := make([]DisplayableFrame, 0, len(r.EFRMs))
	for i, frame := range r.EFRMs {
		idx := i + 1
		if idx < 0 || idx > math.MaxUint16 {
			return nil,
				fmt.Errorf("%w: index %d", ErrFrameIndexOutOfRange, i+1)
		}

		var pt *DisplayableThumbnail
		if t, ok := thumbnails[uint16(idx)]; ok {
			pt = t
		}

		framePF, errPF := newFrameBuilder(frame, pt, false).
			WithBasicInfo().
			WithCameraModesAndFlashInfo().
			WithCustomFunctionsAndFocusPoints().
			Build()
		if errPF != nil {
			return nil, errPF
		}

		frames = append(frames, framePF)
	}

	return frames, nil
}

func (r displayableRoll) DisplayRoll() {
	s := fmt.Sprintf("%-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s",
		filmIDWidth, "FILM ID",
		firstRowWidth, "FIRST ROW",
		perRowWidth, "FRAMES PER ROW",
		titleWidth, "TITLE",
		filmLoadedAtWidth, "FILM LOADED AT",
		frameCountWidth, "FRAME COUNT",
		isoDxWidth, "ISO (DX)",
		remarksWidth, "REMARKS",
	)
	fmt.Fprintln(os.Stdout, s)
	fmt.Fprintln(os.Stdout, strings.Repeat("-", len(s)))

	row := fmt.Sprintf("%-*s %-*d %-*d %-*s %-*s %-*d %-*s %-*s",
		filmIDWidth, r.FilmID,
		firstRowWidth, r.FirstRow,
		perRowWidth, r.PerRow,
		titleWidth, truncate(r.Title, titleWidth),
		filmLoadedAtWidth, r.FilmLoadedDate,
		frameCountWidth, r.FrameCount,
		isoDxWidth, r.IsoDX,
		remarksWidth, truncate(r.Remarks, remarksWidth),
	)
	fmt.Fprintln(os.Stdout, row)
}

func (r displayableRoll) DisplayFrames() error {
	//nolint:golines // more readable this way
	s := fmt.Sprintf("%-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s",
		filmIDWidth, "FILM ID",
		frameNumberWidth, "FRAME NO.",
		filmLoadedAtWidth, "FILM LOADED AT",
		isoDxWidth, "ISO (DX)",
		focalLengthWidth, "FOCAL LENGTH",
		maxApertureWidth, "MAX APERTURE",
		tvWidth, "TV",
		avWidth, "AV",
		isoMWidth, "ISO (M)",
		exposureCompWidth, "EXPOSURE COMP.",
		flashExposureCompWidth, "FLASH EXPOSURE COMP.",
		flashModeWidth, "FLASH MODE",
		meteringModeWidth, "METERING MODE",
		shootingModeWidth, "SHOOTING MODE",
		filmAdvanceModeWidth, "FILM ADVANCE MODE",
		afModeWidth, "AF MODE",
		bulbExposureTimeWidth, "BULB EXPOSURE TIME",
		takenAtWidth, "TAKEN AT",
		multipleExposureWidth, "MULTIPLE EXPOSURE",
		batteryLoadedAtWidth, "BATTERY LOADED AT",
		remarksWidth, "REMARKS",
	)
	fmt.Fprintln(os.Stdout, s)
	fmt.Fprintln(os.Stdout, strings.Repeat("-", len(s)))

	for _, fr := range r.Frames {
		row := r.renderFrame(fr)
		fmt.Fprintln(os.Stdout, row)
	}

	return nil
}

func (r displayableRoll) renderFrame(fr DisplayableFrame) string {
	//nolint:golines // more readable this way
	row := fmt.Sprintf("%-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s",
		filmIDWidth, fr.FilmID,
		frameNumberWidth, r.renderFrameNumber(fr),
		filmLoadedAtWidth, fr.FilmLoadedAt,
		isoDxWidth, fr.IsoDX,
		focalLengthWidth, fr.FocalLength,
		maxApertureWidth, fr.MaxAperture,
		tvWidth, fr.Tv,
		avWidth, fr.Av,
		isoMWidth, fr.IsoM,
		exposureCompWidth, fr.ExposureCompensation,
		flashExposureCompWidth, fr.FlashExposureComp,
		flashModeWidth, fr.FlashMode,
		meteringModeWidth, fr.MeteringMode,
		shootingModeWidth, fr.ShootingMode,
		filmAdvanceModeWidth, fr.FilmAdvanceMode,
		afModeWidth, fr.AFMode,
		bulbExposureTimeWidth, fr.BulbExposureTime,
		takenAtWidth, fr.TakenAt,
		multipleExposureWidth, fr.MultipleExposure,
		batteryLoadedAtWidth, fr.BatteryLoadedAt,
		remarksWidth, truncate(fr.Remarks, remarksWidth),
	)

	return row
}

func (r displayableRoll) DisplayCustomFunctions() {
	//nolint:golines // more readable this way
	s := fmt.Sprintf("%-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s",
		filmIDWidth, "FILM ID",
		frameNumberWidth, "FRAME NO.",
		customFunctionsWidth, "#",
		customFunctionsWidth, "1",
		customFunctionsWidth, "2",
		customFunctionsWidth, "3",
		customFunctionsWidth, "4",
		customFunctionsWidth, "5",
		customFunctionsWidth, "6",
		customFunctionsWidth, "7",
		customFunctionsWidth, "8",
		customFunctionsWidth, "9",
		customFunctionsWidth, "10",
		customFunctionsWidth, "11",
		customFunctionsWidth, "12",
		customFunctionsWidth, "13",
		customFunctionsWidth, "14",
		customFunctionsWidth, "15",
		customFunctionsWidth, "16",
		customFunctionsWidth, "17",
		customFunctionsWidth, "18",
		customFunctionsWidth, "19",
		customFunctionsWidth, "20",
	)
	fmt.Fprintln(os.Stdout, s)
	fmt.Fprintln(os.Stdout, strings.Repeat("-", len(s)))

	for _, fr := range r.Frames {
		row := r.renderCustomFunctions(fr)
		fmt.Fprintln(os.Stdout, row)
	}
}

func (r displayableRoll) renderCustomFunctions(fr DisplayableFrame) string {
	//nolint:golines // more readable this way
	row := fmt.Sprintf("%-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s",
		filmIDWidth, fr.FilmID,
		frameNumberWidth, r.renderFrameNumber(fr),
		customFunctionsWidth, "#",
		customFunctionsWidth, fr.CustomFunctions[0],
		customFunctionsWidth, fr.CustomFunctions[1],
		customFunctionsWidth, fr.CustomFunctions[2],
		customFunctionsWidth, fr.CustomFunctions[3],
		customFunctionsWidth, fr.CustomFunctions[4],
		customFunctionsWidth, fr.CustomFunctions[5],
		customFunctionsWidth, fr.CustomFunctions[6],
		customFunctionsWidth, fr.CustomFunctions[7],
		customFunctionsWidth, fr.CustomFunctions[8],
		customFunctionsWidth, fr.CustomFunctions[9],
		customFunctionsWidth, fr.CustomFunctions[10],
		customFunctionsWidth, fr.CustomFunctions[11],
		customFunctionsWidth, fr.CustomFunctions[12],
		customFunctionsWidth, fr.CustomFunctions[13],
		customFunctionsWidth, fr.CustomFunctions[14],
		customFunctionsWidth, fr.CustomFunctions[15],
		customFunctionsWidth, fr.CustomFunctions[16],
		customFunctionsWidth, fr.CustomFunctions[17],
		customFunctionsWidth, fr.CustomFunctions[18],
		customFunctionsWidth, fr.CustomFunctions[19],
	)

	return row
}

func (r displayableRoll) DisplayFocusingPoints() error {
	rows := make([]string, len(r.Frames))
	for i, fr := range r.Frames {
		row, err := r.renderFocusPoints(fr)
		if err != nil {
			return err
		}

		rows[i] = row
	}

	s := fmt.Sprintf("%-*s %-*s %-*s",
		filmIDWidth, "FILM ID",
		frameNumberWidth, "FRAME NO.",
		focusingPointsWidth, "FOCUSING POINTS",
	)
	fmt.Fprintln(os.Stdout, s)
	fmt.Fprintln(os.Stdout, strings.Repeat("-", len(s)))

	for _, row := range rows {
		fmt.Fprintln(os.Stdout, row)
	}

	return nil
}

func (r displayableRoll) renderFocusPoints(
	f DisplayableFrame,
) (string, error) {
	pf := f.FocusingPoints

	if pf.Selection == math.MaxUint32 {
		empty := "    \033[30m\u25AF \u25AF \u25AF \u25AF \u25AF \u25AF \u25AF\n" +
			" \u25AF \u25AF \u25AF \u25AF \u25AF \u25AF \u25AF \u25AF \u25AF \u25AF\n" +
			"\u25AF \u25AF \u25AF \u25AF \u25AF \u25AF \u25AF \u25AF \u25AF \u25AF \u25AF\n" +
			" \u25AF \u25AF \u25AF \u25AF \u25AF \u25AF \u25AF \u25AF \u25AF \u25AF\n" +
			"    \u25AF \u25AF \u25AF \u25AF \u25AF \u25AF \u25AF\033[0m\n"

		s := fmt.Sprintf("%-*s %-*d %-*s",
			filmIDWidth, f.FilmID,
			frameNumberWidth, f.FrameNumber,
			focusingPointsWidth, pad(empty, focusingPointsPadding),
		)

		return s, nil
	}

	p := make([]string, len(fpBits))
	for i, fpBit := range fpBits {
		b, err := byteToBox(pf.Points[i], fpBit)
		if err != nil {
			return "", err
		}

		p[i] = b
	}

	printableFocusPoints := "    " + p[0] + "\n" +
		" " + p[2] + p[1] + "\n" +
		p[4] + p[3] + "\n" +
		" " + p[6] + p[5] + "\n" +
		"    " + p[7] + "\n"

	s := fmt.Sprintf("%-*s %-*d %-*s",
		filmIDWidth, f.FilmID,
		frameNumberWidth, f.FrameNumber,
		focusingPointsWidth, pad(printableFocusPoints, focusingPointsPadding),
	)

	return s, nil
}

func (r displayableRoll) DisplayThumbnails() {
	s := fmt.Sprintf("%-*s %-*s %-*s %-*s",
		filmIDWidth, "FILM ID",
		frameNumberWidth, "FRAME NO.",
		imageFileWidth, "IMAGE FILE",
		thumbnailWidth, "THUMBNAIL",
	)
	fmt.Fprintln(os.Stdout, s)
	fmt.Fprintln(os.Stdout, strings.Repeat("-", len(s)))

	for _, fr := range r.Frames {
		r.displayThumbnail(fr)
	}
}

func (r displayableRoll) displayThumbnail(fr DisplayableFrame) {
	t := DisplayableThumbnail{
		Filepath:  "",
		Thumbnail: "",
	}

	if fr.Thumbnail != nil {
		t = *fr.Thumbnail
	}

	if fr.Thumbnail != nil {
		s := fmt.Sprintf("%-*s %-*d %-*s %-*s",
			filmIDWidth, fr.FilmID,
			frameNumberWidth, fr.FrameNumber,
			imageFileWidth, truncate(t.Filepath, imageFileWidth),
			thumbnailWidth, pad(t.Thumbnail, thumbnailPadding),
		)
		fmt.Fprint(os.Stdout, s)
	}
}

type DisplayableFrame struct {
	FrameNumber  uint
	FilmID       domain.FilmID
	FilmLoadedAt domain.ValidatedDatetime
	IsoDX        domain.Iso

	UserModifiedRecord bool

	FocalLength domain.FocalLength
	MaxAperture domain.Av
	Tv          domain.Tv
	Av          domain.Av
	IsoM        domain.Iso

	ExposureCompensation domain.ExposureCompenation
	FlashExposureComp    domain.ExposureCompenation
	FlashMode            domain.FlashMode
	MeteringMode         domain.MeteringMode
	ShootingMode         domain.ShootingMode

	FilmAdvanceMode  domain.FilmAdvanceMode
	AFMode           domain.AutoFocusMode
	BulbExposureTime domain.BulbExposureTime
	TakenAt          domain.ValidatedDatetime

	MultipleExposure domain.MultipleExposure
	BatteryLoadedAt  domain.ValidatedDatetime

	CustomFunctions DisplayableCustomFunctions
	Remarks         domain.Remarks

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

func (r displayableRoll) renderFrameNumber(fr DisplayableFrame) string {
	if !fr.UserModifiedRecord {
		return strconv.FormatUint(uint64(fr.FrameNumber), 10)
	}

	return fmt.Sprintf("%d*", fr.FrameNumber)
}

func truncate[S ~string](s S, l int) S {
	if len(s) <= l {
		return s
	}

	return s[:l-3] + "..."
}

func pad[S ~string](s S, p int) S {
	lines := strings.Split(string(s), "\n")

	var sb strings.Builder
	sb.WriteString(lines[0] + "\n")

	for i := 1; i < len(lines); i++ {
		sb.WriteString(strings.Repeat(" ", p) + lines[i] + "\n")
	}

	return S(sb.String())
}
