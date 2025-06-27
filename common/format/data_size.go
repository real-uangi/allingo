package format

import (
	"fmt"
	"slices"
)

type DataSizeUnit string

const (
	SizeUnitB  DataSizeUnit = "B"
	SizeUnitKB DataSizeUnit = "KB"
	SizeUnitMB DataSizeUnit = "MB"
	SizeUnitGB DataSizeUnit = "GB"
	SizeUnitTB DataSizeUnit = "TB"
	SizeUnitPB DataSizeUnit = "PB"
	SizeUnitEB DataSizeUnit = "EB"
	SizeUnitZB DataSizeUnit = "ZB"
	SizeUnitYB DataSizeUnit = "YB"
)

func getUnitSlice() []DataSizeUnit {
	units := make([]DataSizeUnit, 9)
	units[0] = SizeUnitB
	units[1] = SizeUnitKB
	units[2] = SizeUnitMB
	units[3] = SizeUnitGB
	units[4] = SizeUnitTB
	units[5] = SizeUnitPB
	units[6] = SizeUnitEB
	units[7] = SizeUnitZB
	units[8] = SizeUnitYB
	return units
}

func DataSize(size uint64) string {
	amount, unit := DataSizeAuto(size)
	if unit == SizeUnitB {
		return fmt.Sprintf("%.0f%s", amount, unit)
	}
	return fmt.Sprintf("%.2f%s", amount, unit)
}

func DataSizeAuto(size uint64) (float64, DataSizeUnit) {
	const threshold float64 = 1024
	const maxLevel = 8
	var amount = float64(size)
	var level int
	for amount >= threshold && level < maxLevel {
		amount = amount / threshold
		level++
	}
	return amount, getUnitSlice()[level]
}

func DataSizeAsString(size uint64, unit DataSizeUnit) string {
	amount, unit := DataSizeAs(size, unit)
	if unit == SizeUnitB || unit == SizeUnitKB {
		return fmt.Sprintf("%.0f%s", amount, unit)
	}
	return fmt.Sprintf("%.2f%s", amount, unit)
}

func DataSizeAs(size uint64, unit DataSizeUnit) (float64, DataSizeUnit) {
	const threshold float64 = 1024
	var amount = float64(size)
	units := getUnitSlice()
	i := slices.Index(units, unit)
	if i == 0 {
		return amount, units[0]
	}
	for j := 0; j < i; j++ {
		amount = amount / threshold
	}
	return amount, units[i]
}
