package auth

import (
	"encoding/base64"
	"github.com/oxfeeefeee/appgo"
	"github.com/oxfeeefeee/appgo/toolkit/crypto"
	"strconv"
	"strings"
	"time"
)

type Token string

func NewToken(userId appgo.Id, role appgo.Role, lifetime int) Token {
	key := appgo.Conf.RootKey
	expires := appgo.Id(time.Now().Add(time.Second * time.Duration(lifetime)).UnixNano())
	parts := []string{userId.Base64(), strconv.Itoa(int(role)), expires.Base64()}
	data := strings.Join(parts, ",")
	keybyte, _ := crypto.Encrypt([]byte(data), []byte(key))
	str := base64.StdEncoding.EncodeToString(keybyte)
	return Token(str)
}

func (t Token) Validate() (appgo.Id, appgo.Role) {
	byteToken, err := base64.StdEncoding.DecodeString(string(t))
	if err != nil {
		return 0, 0
	}
	key := appgo.Conf.RootKey
	decrypted, err := crypto.Decrypt([]byte(byteToken), []byte(key))
	if err != nil {
		return 0, 0
	}
	subs := strings.Split(string(decrypted), ",")
	if len(subs) != 3 {
		return 0, 0
	}
	userId := appgo.IdFromStr(subs[0])
	roleInt, _ := strconv.Atoi(subs[1])
	expiry := time.Unix(0, int64(appgo.IdFromStr(subs[2])))
	ttl := expiry.Sub(time.Now())
	if ttl <= 0 {
		return 0, 0
	}
	return userId, appgo.Role(roleInt)
}
