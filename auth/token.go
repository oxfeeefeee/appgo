package auth

import (
	"encoding/base64"
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/oxfeeefeee/appgo"
	"github.com/oxfeeefeee/appgo/redis"
	"github.com/oxfeeefeee/appgo/toolkit/crypto"
	"strconv"
	"strings"
	"time"
)

type Token string

var tokenCache *redis.Hashs

func InitTokenCache() {
	tokenCache = redis.NewHashs("userToken", redis.BaseClient)
}

func GetCacheToken(uid appgo.Id) (string, error) {
	return tokenCache.GetValue(uid.String())
}

func SetCacheToken(uid appgo.Id, newToken string) error {
	_, err := tokenCache.SetValue(uid.String(), newToken)
	return err
}

func NewToken(userId appgo.Id, role appgo.Role) Token {
	lifetime := tokenLifetime(role)
	key := appgo.Conf.RootKey
	expires := appgo.Id(time.Now().Add(time.Second * time.Duration(lifetime)).UnixNano())
	parts := []string{userId.Base64(), strconv.Itoa(int(role)), expires.Base64()}
	data := strings.Join(parts, ",")
	keybyte, err := crypto.Encrypt([]byte(data), []byte(key))
	if err != nil {
		log.Errorln("failed to encrypt token")
		return Token("")
	}
	str := base64.StdEncoding.EncodeToString(keybyte)
	return Token(str)
}

func (t Token) Validate() (appgo.Id, appgo.Role) {
	byteToken, err := base64.StdEncoding.DecodeString(string(t))
	if err != nil {
		log.Infoln("validate token failed: ", err)
		return 0, 0
	}
	key := appgo.Conf.RootKey
	decrypted, err := crypto.Decrypt([]byte(byteToken), []byte(key))
	if err != nil {
		log.Infoln("validate token failed: ", err)
		return 0, 0
	}
	subs := strings.Split(string(decrypted), ",")
	if len(subs) != 3 {
		log.Errorln("bad token format")
		return 0, 0
	}
	userId := appgo.IdFromBase64(subs[0])
	roleInt, _ := strconv.Atoi(subs[1])
	expiry := time.Unix(0, int64(appgo.IdFromBase64(subs[2])))
	ttl := expiry.Sub(time.Now())
	if ttl <= 0 {
		log.Infoln("validate token failed: expired at ", expiry)
		return 0, 0
	}
	// check if user is baned
	if userSystem.IsBanned(userId) {
		return 0, 0
	}
	return userId, appgo.Role(roleInt)
}

func (t Token) Parse() (appgo.Id, appgo.Role, *time.Time, error) {
	byteToken, err := base64.StdEncoding.DecodeString(string(t))
	if err != nil {
		return 0, 0, nil, errors.New(fmt.Sprintf("Parse(), DecodeString failed, err: %v", err))
	}
	key := appgo.Conf.RootKey
	decrypted, err := crypto.Decrypt([]byte(byteToken), []byte(key))
	if err != nil {
		return 0, 0, nil, errors.New(fmt.Sprintf("Parse(), Decrypt failed, err: %v", err))
	}
	subs := strings.Split(string(decrypted), ",")
	if len(subs) != 3 {
		return 0, 0, nil, errors.New(fmt.Sprintf("Parse(), Split failed, bad token format"))
	}
	userId := appgo.IdFromBase64(subs[0])
	roleInt, _ := strconv.Atoi(subs[1])
	expiry := time.Unix(0, int64(appgo.IdFromBase64(subs[2])))
	return userId, appgo.Role(roleInt), &expiry, nil
}

func tokenLifetime(role appgo.Role) int {
	switch role {
	case appgo.RoleAppUser:
		return appgo.Conf.TokenLifetime.AppUser
	case appgo.RoleWebUser:
		return appgo.Conf.TokenLifetime.WebUser
	case appgo.RoleWebAdmin:
		return appgo.Conf.TokenLifetime.WebAdmin
	case appgo.RoleAuthor:
		return appgo.Conf.TokenLifetime.Author
	default:
		return 0
	}
}
