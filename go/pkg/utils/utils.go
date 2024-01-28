package utils

import (
	"fmt"
	"strconv"
)

func UintToString(n uint) string {
	return strconv.FormatUint(uint64(n), 10)
}

func StringToUint(s string) (uint, error) {
	u64, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		fmt.Println(err)
	}
	wd := uint(u64)
	return wd, err
}
