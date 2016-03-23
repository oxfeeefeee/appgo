package auth

import (
	"errors"
	///log "github.com/Sirupsen/logrus"
	"github.com/oxfeeefeee/appgo"
	"github.com/oxfeeefeee/appgo/toolkit/crypto"
	//"github.com/oxfeeefeee/appgo/services/qq"
	///"github.com/oxfeeefeee/appgo/services/weibo"
	//"github.com/oxfeeefeee/appgo/services/weixin"
)

const (
	mobileCodeLen        = 6
	mobileTokenLen       = 16
	mobileCodeTimeout    = 2 * 60
	mobileTokenTimeout   = 10 * 60
	mobileCodeKeyPrefix  = "mobilecode:"
	mobileTokenKeyPrefix = "mobiletoken:"
)

type MobileUserInfo struct {
	Mobile   string
	Token    string
	Nickname string
	Password string
	Sex      appgo.Sex
}

type MobileSupport interface {
	GetMobileUser(mobile string) (uid appgo.Id, err error)
	AddMobileUser(info *MobileUserInfo) (uid appgo.Id, err error)
	appgo.MobileMsgSender
	appgo.KvStore
}

func SendMobileCode(mobile string) error {
	code := crypto.RandNumStr(mobileCodeLen)
	if err := mobileSupport.StoreKeyValue(
		mobileCodeKeyPrefix+mobile, code, mobileCodeTimeout); err != nil {
		return err
	}
	return mobileSupport.SendMobileCode(mobile, code)
}

func VerifyMobileCode(mobile, code string) (string, error) {
	scode, err := mobileSupport.GetValueByKey(mobileCodeKeyPrefix + mobile)
	if err != nil {
		return "", err
	}
	if scode != code {
		return "", appgo.MobileUserBadCodeErr
	}
	token := crypto.RandNumStr(mobileTokenLen)
	return token, nil
}

func RegisterMobileUser(info *MobileUserInfo, role appgo.Role) (*LoginResult, error) {
	stoken, err := mobileSupport.GetValueByKey(mobileTokenKeyPrefix + info.Mobile)
	if err != nil {
		return nil, err
	}
	if stoken != info.Token {
		return nil, appgo.MobileUserBadTokenErr
	}
	uid, err := mobileSupport.AddMobileUser(info)
	if err != nil {
		return nil, err
	}
	return checkIn(uid, role)
}

func LoginByMobile(mobile string, role appgo.Role) (*LoginResult, error) {
	if mobileSupport == nil {
		return nil, errors.New("qq mobile not supported")
	}
	uid, err := mobileSupport.GetMobileUser(mobile)
	if err != nil {
		return nil, err
	}
	if uid == 0 {

	}
	return checkIn(uid, role)
}
