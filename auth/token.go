package auth

import (
	"github.com/oxfeeefeee/appgo"
	"github.com/oxfeeefeee/appgo/toolkit/crypto"
	"time"
)

type Token string

func newToken(userId appgo.Id) Token {
	key := appgo.Conf.RootKey
	expires := appgo.Id(time.Now().Add(time.Second * time.Duration(exp)).Second)
	parts := []string{userId.Base64(), strutil.IntToStr(int(role)), expires.Base64()}
	data := strings.Join(parts, ",")
	return crypto.Encrypt([]byte(data), []byte(key))
}

func (t Token) Validate() {

}
