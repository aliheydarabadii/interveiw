package domain

import "time"

type Measurement struct {
	AssetID   AssetID
	Name      string
	Value     float64
	Unit      string
	Timestamp time.Time
}
