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

package domain

import (
	_ "embed"
	"encoding/json"
	"strconv"
)

//go:embed domain.json
var domainData []byte

// MapProvider provides lookup maps for converting raw Canon camera metadata values
// to human-readable strings. It loads data from an embedded JSON file containing
// the canonical mappings for shutter speeds, apertures, flash modes, and other settings.
type MapProvider struct {
	tvs  map[int32]Tv
	avs  map[uint32]Av
	ecs  map[int32]ExposureCompensation
	fms  map[uint32]FlashMode
	mms  map[uint32]MeteringMode
	sms  map[uint32]ShootingMode
	fams map[uint32]FilmAdvanceMode
	afms map[uint32]AutoFocusMode
	mes  map[uint32]MultipleExposure
	cfl  map[int]byte
}

type domainJSON struct {
	ShutterSpeeds         map[string]string `json:"shutterSpeeds"`
	ApertureValues        map[string]string `json:"apertureValues"`
	ExposureCompensations map[string]string `json:"exposureCompensations"`
	FlashModes            map[string]string `json:"flashModes"`
	MeteringModes         map[string]string `json:"meteringModes"`
	ShootingModes         map[string]string `json:"shootingModes"`
	FilmAdvanceModes      map[string]string `json:"filmAdvanceModes"`
	AutoFocusModes        map[string]string `json:"autoFocusModes"`
	MultipleExposures     map[string]string `json:"multipleExposures"`
	CustomFunctionsLimits map[string]byte   `json:"customFunctionsLimits"`
}

func NewMapProvider() *MapProvider {
	var data domainJSON

	_ = json.Unmarshal(domainData, &data)

	return &MapProvider{
		tvs:  convertMapInt32[Tv](data.ShutterSpeeds),
		avs:  convertMapUint32[Av](data.ApertureValues),
		ecs:  convertMapInt32[ExposureCompensation](data.ExposureCompensations),
		fms:  convertMapUint32[FlashMode](data.FlashModes),
		mms:  convertMapUint32[MeteringMode](data.MeteringModes),
		sms:  convertMapUint32[ShootingMode](data.ShootingModes),
		fams: convertMapUint32[FilmAdvanceMode](data.FilmAdvanceModes),
		afms: convertMapUint32[AutoFocusMode](data.AutoFocusModes),
		mes:  convertMapUint32[MultipleExposure](data.MultipleExposures),
		cfl:  convertCustomFunctionsLimits(data.CustomFunctionsLimits),
	}
}

func convertMapInt32[V ~string](src map[string]string) map[int32]V {
	result := make(map[int32]V, len(src))

	for k, v := range src {
		if i, err := strconv.ParseInt(k, 10, 32); err == nil {
			result[int32(i)] = V(v)
		}
	}

	return result
}

func convertMapUint32[V ~string](src map[string]string) map[uint32]V {
	result := make(map[uint32]V, len(src))
	for k, v := range src {
		if i, err := strconv.ParseUint(k, 10, 32); err == nil {
			result[uint32(i)] = V(v)
		}
	}

	return result
}

func convertCustomFunctionsLimits(src map[string]byte) map[int]byte {
	result := make(map[int]byte, len(src))
	for k, v := range src {
		if i, err := strconv.Atoi(k); err == nil {
			result[i] = v
		}
	}

	return result
}

// GetTv retrieves the human-readable shutter speed string for a raw Tv value.
// Returns false if the value is not found in the lookup map.
func (m *MapProvider) GetTv(tv int32) (Tv, bool) {
	r, ok := m.tvs[tv]

	return r, ok
}

// GetAv retrieves the human-readable aperture value for a raw Av value.
// Returns false if the value is not found in the lookup map.
func (m *MapProvider) GetAv(av uint32) (Av, bool) {
	r, ok := m.avs[av]

	return r, ok
}

// GetExposureCompensation retrieves the exposure compensation string for a raw value.
// Returns false if the value is not found in the lookup map.
func (m *MapProvider) GetExposureCompensation(
	ec int32,
) (ExposureCompensation, bool) {
	r, ok := m.ecs[ec]

	return r, ok
}

// GetFlashMode retrieves the flash mode string for a raw value.
// Returns false if the value is not found in the lookup map.
func (m *MapProvider) GetFlashMode(fm uint32) (FlashMode, bool) {
	r, ok := m.fms[fm]

	return r, ok
}

// GetMeteringMode retrieves the metering mode string for a raw value.
// Returns false if the value is not found in the lookup map.
func (m *MapProvider) GetMeteringMode(mm uint32) (MeteringMode, bool) {
	r, ok := m.mms[mm]

	return r, ok
}

// GetShootingMode retrieves the shooting mode string for a raw value.
// Returns false if the value is not found in the lookup map.
func (m *MapProvider) GetShootingMode(sm uint32) (ShootingMode, bool) {
	r, ok := m.sms[sm]

	return r, ok
}

// GetFilmAdvanceMode retrieves the film advance mode string for a raw value.
// Returns false if the value is not found in the lookup map.
func (m *MapProvider) GetFilmAdvanceMode(fam uint32) (FilmAdvanceMode, bool) {
	r, ok := m.fams[fam]

	return r, ok
}

// GetAutoFocusMode retrieves the autofocus mode string for a raw value.
// Returns false if the value is not found in the lookup map.
func (m *MapProvider) GetAutoFocusMode(afm uint32) (AutoFocusMode, bool) {
	r, ok := m.afms[afm]

	return r, ok
}

// GetMultipleExposure retrieves the multiple exposure setting string for a raw value.
// Returns false if the value is not found in the lookup map.
func (m *MapProvider) GetMultipleExposure(me uint32) (MultipleExposure, bool) {
	r, ok := m.mes[me]

	return r, ok
}
