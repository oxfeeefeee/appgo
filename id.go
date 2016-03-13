package appgo

import (
	"encoding/base64"
	"encoding/binary"
)

var (
	base64Encoding = base64.RawURLEncoding
	binaryEndian   = binary.BigEndian
)

type Id int64

func (id Id) String() string {
	return strconv.FormatInt(int64(id), 10)
}

func IdFromStr(str string) Id {
	i, _ := strconv.Atoi(str)
	return Id(i)
}

func IdFromBase64(str string) Id {
	if len(s) == 0 {
		return 0
	} else if data, err := base64Encoding.DecodeString(s); err != nil {
		return 0
	} else {
		val := binaryEndian.Uint64(data)
		return Id(int64(val))
	}
}

func (id Uid) Base64() string {
	buf := make([]byte, 8)
	binaryEndian.PutUint64(buf, uint64(int64(id)))
	return base64Encoding.EncodeToString(buf)
}
