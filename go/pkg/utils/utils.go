package utils

import (
	"strconv"
)

func UintToString(n uint) string {
	return strconv.FormatUint(uint64(n), 10)
}

func StringToUint(s string) (uint, error) {
	u64, err := strconv.ParseUint(s, 10, 32)
	wd := uint(u64)
	return wd, err
}

func Float64ToUint(f float64) (uint, error) {
	str := strconv.FormatFloat(f, 'f', -1, 64)
	return StringToUint(str)
}
