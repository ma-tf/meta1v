//go:generate mockgen -destination=./mocks/service_mock.go -package=display_test github.com/ma-tf/meta1v/internal/service/display Service
package display

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
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

	frameNumberWidth               = 9
	focusingPointsWidth            = 21
	focusingPointsPadding          = filmIDWidth + frameNumberWidth + 2
	focalLengthWidth               = 12
	maxApertureWidth               = 12
	tvWidth                        = 7
	avWidth                        = 7
	isoMWidth                      = 7
	exposureCompWidth              = 15
	flashExposureCompensationWidth = 20
	flashModeWidth                 = 15
	meteringModeWidth              = 15
	shootingModeWidth              = 15
	filmAdvanceModeWidth           = 17
	afModeWidth                    = 12
	bulbExposureTimeWidth          = 20
	takenAtWidth                   = 20
	multipleExposureWidth          = 20
	batteryLoadedAtWidth           = 20
	customFunctionsWidth           = 2

	imageFileWidth   = 64
	thumbnailWidth   = 64
	thumbnailPadding = filmIDWidth + frameNumberWidth + imageFileWidth + 3
)

var (
	ErrMultipleThumbnailsForFrame = errors.New(
		"frame has multiple thumbnails",
	)
	ErrFrameIndexOutOfRange = errors.New("frame index out of range")
)

// Service provides formatted text output operations for displaying EFD metadata.
type Service interface {
	// DisplayRoll writes formatted roll-level metadata to the writer.
	DisplayRoll(w io.Writer, r DisplayableRoll)

	// DisplayCustomFunctions writes a table of custom function settings for all frames.
	DisplayCustomFunctions(w io.Writer, r DisplayableRoll) error

	// DisplayFocusingPoints writes focus point visualizations for all frames.
	DisplayFocusingPoints(w io.Writer, r DisplayableRoll)

	// DisplayFrames writes a detailed table of frame metadata.
	DisplayFrames(w io.Writer, r DisplayableRoll)

	// DisplayThumbnails writes ASCII art thumbnails for all frames.
	DisplayThumbnails(w io.Writer, r DisplayableRoll)
}

type service struct{}

func NewService() Service {
	return &service{}
}

func (s *service) DisplayRoll(w io.Writer, r DisplayableRoll) {
	header := fmt.Sprintf("%-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s",
		filmIDWidth, "FILM ID",
		firstRowWidth, "FIRST ROW",
		perRowWidth, "FRAMES PER ROW",
		titleWidth, "TITLE",
		filmLoadedAtWidth, "FILM LOADED AT",
		frameCountWidth, "FRAME COUNT",
		isoDxWidth, "ISO (DX)",
		remarksWidth, "REMARKS",
	)
	fmt.Fprintln(w, header)
	fmt.Fprintln(w, strings.Repeat("-", len(header)))

	row := fmt.Sprintf("%-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s",
		filmIDWidth, r.FilmID,
		firstRowWidth, r.FirstRow,
		perRowWidth, r.PerRow,
		titleWidth, truncate(r.Title, titleWidth),
		filmLoadedAtWidth, r.FilmLoadedDate,
		frameCountWidth, r.FrameCount,
		isoDxWidth, r.IsoDX,
		remarksWidth, truncate(r.Remarks, remarksWidth),
	)
	fmt.Fprintln(w, row)
}

func (s *service) DisplayFrames(w io.Writer, r DisplayableRoll) {
	//nolint:golines // more readable this way
	header := fmt.Sprintf("%-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s",
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
		flashExposureCompensationWidth, "FLASH EXPOSURE COMP.",
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
	fmt.Fprintln(w, header)
	fmt.Fprintln(w, strings.Repeat("-", len(header)))

	for _, fr := range r.Frames {
		row := s.renderFrame(fr)
		fmt.Fprintln(w, row)
	}
}

func (s *service) renderFrame(fr DisplayableFrame) string {
	//nolint:golines // more readable this way
	row := fmt.Sprintf("%-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s",
		filmIDWidth, fr.FilmID,
		frameNumberWidth, s.renderFrameNumber(fr),
		filmLoadedAtWidth, fr.FilmLoadedAt,
		isoDxWidth, fr.IsoDX,
		focalLengthWidth, fr.FocalLength,
		maxApertureWidth, fr.MaxAperture,
		tvWidth, fr.Tv,
		avWidth, fr.Av,
		isoMWidth, fr.IsoM,
		exposureCompWidth, fr.ExposureCompensation,
		flashExposureCompensationWidth, fr.FlashExposureCompensation,
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

func (s *service) DisplayCustomFunctions(w io.Writer, r DisplayableRoll) error {
	//nolint:golines // more readable this way
	header := fmt.Sprintf("%-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s",
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
	fmt.Fprintln(w, header)
	fmt.Fprintln(w, strings.Repeat("-", len(header)))

	for _, fr := range r.Frames {
		row := s.renderCustomFunctions(fr)
		fmt.Fprintln(w, row)
	}

	return nil
}

func (s *service) renderCustomFunctions(fr DisplayableFrame) string {
	//nolint:golines // more readable this way
	row := fmt.Sprintf("%-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s",
		filmIDWidth, fr.FilmID,
		frameNumberWidth, s.renderFrameNumber(fr),
		customFunctionsWidth, " ",
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

func (s *service) DisplayFocusingPoints(w io.Writer, r DisplayableRoll) {
	rows := make([]string, len(r.Frames))
	for i, fr := range r.Frames {
		rows[i] = fmt.Sprintf("%-*s %-*d %-*s",
			filmIDWidth, fr.FilmID,
			frameNumberWidth, fr.FrameNumber,
			focusingPointsWidth, pad(fr.FocusingPoints, focusingPointsPadding),
		)
	}

	header := fmt.Sprintf("%-*s %-*s %-*s",
		filmIDWidth, "FILM ID",
		frameNumberWidth, "FRAME NO.",
		focusingPointsWidth, "FOCUSING POINTS",
	)
	fmt.Fprintln(w, header)
	fmt.Fprintln(w, strings.Repeat("-", len(header)))

	for _, row := range rows {
		fmt.Fprintln(w, row)
	}
}

func (s *service) DisplayThumbnails(w io.Writer, r DisplayableRoll) {
	header := fmt.Sprintf("%-*s %-*s %-*s %-*s",
		filmIDWidth, "FILM ID",
		frameNumberWidth, "FRAME NO.",
		imageFileWidth, "IMAGE FILE",
		thumbnailWidth, "THUMBNAIL",
	)
	fmt.Fprintln(w, header)
	fmt.Fprintln(w, strings.Repeat("-", len(header)))

	for _, fr := range r.Frames {
		s.displayThumbnail(w, fr)
	}
}

func (s *service) displayThumbnail(w io.Writer, fr DisplayableFrame) {
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
		fmt.Fprint(w, s)
	}
}

func (s *service) renderFrameNumber(fr DisplayableFrame) string {
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
