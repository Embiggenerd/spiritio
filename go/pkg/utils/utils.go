package utils

import (
	"encoding/json"
	"fmt"
	"strconv"
)

func PrintableStruct(s interface{}) string {
	res, _ := json.Marshal(s)
	return string(res)
}

func PrintStruct(i interface{}) {
	s, _ := json.MarshalIndent(i, "", "\t")
	fmt.Println(string(s))
}

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
