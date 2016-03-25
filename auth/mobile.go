package auth

import (
	"errors"
	///log "github.com/Sirupsen/logrus"
	"github.com/oxfeeefeee/appgo"
	"github.com/oxfeeefeee/appgo/toolkit/crypto"
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
	GetMobileUser(mobile, password string) (uid appgo.Id, err error)
	AddMobileUser(info *MobileUserInfo) (uid appgo.Id, err error)
	UpdatePwByMobile(mobile, password string) error
	appgo.MobileMsgSender
	appgo.KvStore
}

func MobilePreRegister(mobile string) error {
	return sendSmsCode(mobile, appgo.SmsTemplateRegister)
}

func MobileVerifyRegister(mobile, code string) (string, error) {
	if err := verifySmsCode(mobile, code); err != nil {
		return "", err
	}
	token := crypto.RandNumStr(mobileTokenLen)
	if err := mobileSupport.Set(
		mobileTokenKeyPrefix+mobile, token, mobileTokenTimeout); err != nil {
		return "", err
	}
	return token, nil
}

func MobileRegisterUser(info *MobileUserInfo, role appgo.Role) (*LoginResult, error) {
	stoken, err := mobileSupport.Get(mobileTokenKeyPrefix + info.Mobile)
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

func MobilePwReset(mobile string) error {
	return sendSmsCode(mobile, appgo.SmsTemplatePwReset)
}

func MobileVerifyPwReset(mobile, code, password string) error {
	if err := verifySmsCode(mobile, code); err != nil {
		return err
	}
	if err := mobileSupport.UpdatePwByMobile(mobile, password); err != nil {
		return err
	}
	return nil
}

func LoginByMobile(mobile, password string, role appgo.Role) (*LoginResult, error) {
	if mobileSupport == nil {
		return nil, errors.New("mobile not supported")
	}
	uid, err := mobileSupport.GetMobileUser(mobile, password)
	if err != nil {
		return nil, err
	}
	if uid == 0 {
		return nil, appgo.MobileUserNotFoundErr
	}
	return checkIn(uid, role)
}

func sendSmsCode(mobile string, template appgo.SmsTemplate) error {
	code := crypto.RandNumStr(mobileCodeLen)
	if err := mobileSupport.Set(
		mobileCodeKeyPrefix+mobile, code, mobileCodeTimeout); err != nil {
		return err
	}
	return mobileSupport.SendMobileCode(mobile, template, code)
}

func verifySmsCode(mobile, code string) error {
	scode, err := mobileSupport.Get(mobileCodeKeyPrefix + mobile)
	if err != nil {
		return err
	}
	if scode != code {
		return appgo.MobileUserBadCodeErr
	}
	return nil
}
