package strutil

import (
	"encoding/json"
	"github.com/oxfeeefeee/appgo"
	"strconv"
)

func FromId(id appgo.Id) string {
	return id.String()
}

func ToId(str string) appgo.Id {
	return appgo.IdFromStr(str)
}

func FromInt64(i int64) string {
	return strconv.FormatInt(i, 10)
}

func ToInt64(str string) int64 {
	if b, err := strconv.ParseInt(str, 10, 64); err != nil {
		return 0
	} else {
		return b
	}
}

func FromInt(val int) string {
	return strconv.Itoa(val)
}

func ToInt(str string) int {
	if b, err := strconv.ParseInt(str, 10, 64); err != nil {
		return 0
	} else {
		return int(b)
	}
}

func FromByte(val byte) string {
	return strconv.Itoa(int(val))
}

func ToByte(str string) byte {
	if b, err := strconv.ParseInt(str, 10, 8); err != nil {
		return 0
	} else {
		return byte(b)
	}
}

func FromBool(b bool) string {
	if b {
		return "1"
	} else {
		return "0"
	}
}

func ToBool(s string) bool {
	if s == "0" || s == "" {
		return false
	} else {
		return true
	}
}

func ToIdList(str string) []appgo.Id {
	if len(str) <= 1 {
		return nil
	} else if str == "null" {
		return nil
	}
	var v []appgo.Id
	json.Unmarshal([]byte(str), &v)
	return v
}

func ToStrList(str string) []string {
	if len(str) <= 1 {
		return nil
	} else if str == "null" {
		return nil
	}
	var v []string
	json.Unmarshal([]byte(str), &v)
	return v
}

func FromIdList(ids []appgo.Id) string {
	if ids == nil {
		return "[]"
	}
	data, _ := json.Marshal(&ids)
	return string(data)
}

func FromStrList(strs []string) string {
	if strs == nil {
		return "[]"
	}
	data, _ := json.Marshal(&strs)
	return string(data)
}
