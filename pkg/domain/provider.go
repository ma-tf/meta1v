package domain

import (
	_ "embed"
	"encoding/json"
	"strconv"
)

//go:embed domain.json
var domainData []byte

type MapProvider struct {
	tvs  map[int32]Tv
	avs  map[uint32]Av
	ecs  map[int32]ExposureCompenation
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
		ecs:  convertMapInt32[ExposureCompenation](data.ExposureCompensations),
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

func (m *MapProvider) GetTv(tv int32) (Tv, bool) {
	r, ok := m.tvs[tv]

	return r, ok
}

func (m *MapProvider) GetAv(av uint32) (Av, bool) {
	r, ok := m.avs[av]

	return r, ok
}

func (m *MapProvider) GetExposureCompenation(
	ec int32,
) (ExposureCompenation, bool) {
	r, ok := m.ecs[ec]

	return r, ok
}

func (m *MapProvider) GetFlashMode(fm uint32) (FlashMode, bool) {
	r, ok := m.fms[fm]

	return r, ok
}

func (m *MapProvider) GetMeteringMode(mm uint32) (MeteringMode, bool) {
	r, ok := m.mms[mm]

	return r, ok
}

func (m *MapProvider) GetShootingMode(sm uint32) (ShootingMode, bool) {
	r, ok := m.sms[sm]

	return r, ok
}

func (m *MapProvider) GetFilmAdvanceMode(fam uint32) (FilmAdvanceMode, bool) {
	r, ok := m.fams[fam]

	return r, ok
}

func (m *MapProvider) GetAutoFocusMode(afm uint32) (AutoFocusMode, bool) {
	r, ok := m.afms[afm]

	return r, ok
}

func (m *MapProvider) GetMultipleExposure(me uint32) (MultipleExposure, bool) {
	r, ok := m.mes[me]

	return r, ok
}
