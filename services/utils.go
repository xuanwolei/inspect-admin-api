package services

import(
	"strconv"
)

func InterfaceToInt(data interface{}) int {
	var res int
	switch v := data.(type) {
	case string:
		res, _ = strconv.Atoi(v)
	case int64:
		strInt64 := strconv.FormatInt(v, 10)
		res, _ = strconv.Atoi(strInt64)
	case int:
		res = v
	default:
		res = 0
	}

	return res
}