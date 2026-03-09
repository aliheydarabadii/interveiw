package domain

import "errors"

var (
	ErrAssetNotFound       = errors.New("asset not found")
	ErrInvalidMeasurement  = errors.New("invalid measurement")
	ErrRegisterMapNotFound = errors.New("register mapping not found")
)
