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

// Package records defines the binary structure of Canon EFD files.
//
// EFD files store metadata recorded by Canon EOS film cameras.
// The file format consists of a series of tagged records (EFDF, EFRM, EFTP)
// that capture information about the film roll, individual frames, and (user linked) thumbnail images.
package records

import "image"

// Raw represents a raw EFD record with its magic bytes, length, and binary data payload.
type Raw struct {
	Magic  [4]byte
	Length uint64
	Data   []byte
}

// Root represents the complete parsed structure of an EFD file,
// containing film roll metadata (EFDF), frame metadata (EFRM), and thumbnail data (EFTP).
type Root struct {
	EFDF  EFDF
	EFRMs []EFRM
	EFTPs []EFTP
}

// EFDF contains metadata about the entire film roll,
// including embedded film ID, film load date, frame count, ISO, and user-entered title and remarks.
type EFDF struct {
	Unknown1   [8]byte
	Unknown2   [8]byte
	Unknown3   [4]byte
	Unknown4   [2]byte
	CodeB      uint32
	Year       uint16 // year film loaded
	Month      uint8  // month film loaded
	Day        uint8  // day film loaded
	Hour       uint8  // datetime hour film loaded
	Minute     uint8  // datetime minute film loaded
	Second     uint8  // datetime second film loaded
	Unknown5   [1]byte
	FrameCount uint32 // total number of frames in the roll
	IsoDX      uint32
	CodeA      uint32
	FirstRow   uint8 // number of frames in first row of your contact sheet (usually less than per row)
	PerRow     uint8 // number of frames per row of your contact sheet
	Unknown6   [128]byte
	Title      [64]byte // null-terminated string, title of the roll
	Remarks    [256]byte
}

// EFRM contains detailed metadata for a single frame, including exposure settings, camera modes,
// timestamps, custom functions, and focus point data. The structure is 512 bytes (0x200).
type EFRM struct {
	// Offset 0x10-0x4F (64 bytes from start of data)
	Unknown1                  [4]byte // 0x10-0x13
	Unknown2                  [4]byte // 0x14-0x17
	FrameNumber               uint32  // 0x18-0x1B
	FocalLength               uint32  // 0x1C-0x1F
	MaxAperture               uint32  // 0x20-0x23
	Tv                        int32   // 0x24-0x27
	Av                        uint32  // 0x28-0x2B
	IsoM                      uint32  // 0x2C-0x2F
	ExposureCompensation      int32   // 0x30-0x33
	FlashExposureCompensation int32   // 0x34-0x37
	Year                      uint16  // 0x38-0x39
	Month                     uint8   // 0x3A
	Day                       uint8   // 0x3B
	Hour                      uint8   // 0x3C
	Minute                    uint8   // 0x3D
	Second                    uint8   // 0x3E
	Unknown3                  [1]byte // 0x3F (padding)
	FlashMode                 uint32  // 0x40-0x43
	FilmAdvanceMode           uint32  // 0x44-0x47
	MultipleExposure          uint32  // 0x48-0x4B
	Unknown4                  uint32  // 0x4C-0x4F

	// Offset 0x50-0x8F (64 bytes)
	Unknown5         uint32  // 0x50-0x53
	MeteringMode     uint32  // 0x54-0x57
	ShootingMode     uint32  // 0x58-0x5B
	AFMode           uint32  // 0x5C-0x5F
	Unknown6         uint32  // 0x60-0x63
	CustomFunction0  uint8   // 0x64
	CustomFunction1  uint8   // 0x65
	CustomFunction2  uint8   // 0x66
	CustomFunction3  uint8   // 0x67
	CustomFunction4  uint8   // 0x68
	CustomFunction5  uint8   // 0x69
	CustomFunction6  uint8   // 0x6A
	CustomFunction7  uint8   // 0x6B
	CustomFunction8  uint8   // 0x6C
	CustomFunction9  uint8   // 0x6D
	CustomFunction10 uint8   // 0x6E
	CustomFunction11 uint8   // 0x6F
	CustomFunction12 uint8   // 0x70
	CustomFunction13 uint8   // 0x71
	CustomFunction14 uint8   // 0x72
	CustomFunction15 uint8   // 0x73
	CustomFunction16 uint8   // 0x74
	CustomFunction17 uint8   // 0x75
	CustomFunction18 uint8   // 0x76
	CustomFunction19 uint8   // 0x77
	Unknown7         [2]byte // 0x78-0x79
	FocusPoints1     uint8   // 0x7A  7 bits used
	FocusPoints2     uint8   // 0x7B  2 bits used
	FocusPoints3     uint8   // 0x7C  8 bits used
	FocusPoints4     uint8   // 0x7D  3 bits used
	FocusPoints5     uint8   // 0x7E  8 bits used
	FocusPoints6     uint8   // 0x7F  2 bits used
	FocusPoints7     uint8   // 0x80  8 bits used
	Unknown8         uint8   // 0x81  not used?
	FocusPoints8     uint8   // 0x82  7 bits used
	Unknown9         [8]byte // 0x83-0x8A
	BatteryYear      uint16  // 0x8B-0x8C
	BatteryMonth     uint8   // 0x8D
	BatteryDay       uint8   // 0x8E
	BatteryHour      uint8   // 0x8F

	// Offset 0x90-0xCF (64 bytes)
	BatteryMinute    uint8    // 0x90
	BatterySecond    uint8    // 0x91
	Unknown10        [5]byte  // 0x92-0x96 (unused/padding)
	BulbExposureTime uint32   // 0x97-0x9A
	FocusingPoint    uint32   // 0x9B-0x9E
	IsModifiedRecord uint8    // 0x9F
	Unknown11        [2]byte  // 0xA0-0xA1
	CodeB            uint32   // 0xA2-0xA5
	CodeA            uint32   // 0xA6-0xA9
	IsoDX            uint32   // 0xAA-0xAD
	RollYear         uint16   // 0xAE-0xAF
	RollMonth        uint8    // 0xB0
	RollDay          uint8    // 0xB1
	RollHour         uint8    // 0xB2
	RollMinute       uint8    // 0xB3
	RollSecond       uint8    // 0xB4
	Unknown12        [1]byte  // 0xB5
	Unknown13        [10]byte // 0xB6-0xBF
	Unknown14        [64]byte // 0xC0-0xFF

	// Offset 0x100-0x1FF (256 bytes)
	Remarks [256]byte // 0x100-0x1FF
}

// EFTP contains thumbnail image data for a frame, including dimensions, file path reference,
// and the decoded RGB image.
type EFTP struct {
	// Offset 0x10-0x4F (64 bytes from start of data)
	Index     uint16    // 0x10-0x11
	Unknown1  uint8     // 0x12
	Unknown2  uint8     // 0x13
	Width     uint16    // 0x14-0x15
	Height    uint16    // 0x16-0x17
	Unknown3  [8]byte   // 0x18-0x1F
	Filepath  [256]byte // 0x20-0x11F
	Thumbnail *image.RGBA
}
