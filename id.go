package appgo

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"math"
	"strconv"
)

var (
	base64Encoding = base64.RawURLEncoding
	binaryEndian   = binary.BigEndian
)

type Id int64

func (id Id) String() string {
	return strconv.FormatInt(int64(id), 10)
}

func (id Id) MarshalJSON() ([]byte, error) {
	var buffer bytes.Buffer
	buffer.WriteByte('"')
	buffer.WriteString(id.String())
	buffer.WriteByte('"')
	return buffer.Bytes(), nil
}

func (id *Id) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	if val, err := strconv.ParseInt(s, 10, 64); err != nil {
		return err
	} else {
		*id = Id(val)
	}
	return nil
}

// for github.com/gorilla/schema
func (id *Id) UnmarshalText(text []byte) (err error) {
	*id = IdFromStr(string(text))
	return nil
}

func IdFromStr(str string) Id {
	i, _ := strconv.ParseInt(str, 10, 64)
	return Id(i)
}

func IdFromBase64(str string) Id {
	if len(str) == 0 {
		return 0
	} else if data, err := base64Encoding.DecodeString(str); err != nil {
		return 0
	} else {
		val := binaryEndian.Uint64(data)
		return Id(int64(val))
	}
}

func IdMax() Id {
	return Id(math.MaxInt64)
}

func (id Id) Base64() string {
	buf := make([]byte, 8)
	binaryEndian.PutUint64(buf, uint64(int64(id)))
	return base64Encoding.EncodeToString(buf)
}
