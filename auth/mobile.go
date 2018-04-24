package auth

import (
	"errors"
	log "github.com/Sirupsen/logrus"
	"github.com/oxfeeefeee/appgo"
	"github.com/oxfeeefeee/appgo/toolkit/crypto"
	"strings"
)

const (
	mobileCodeLen        = 6
	mobileTokenLen       = 16
	mobileCodeTimeout    = 5 * 60
	mobileTokenTimeout   = 10 * 60
	mobileCodeKeyPrefix  = "mobilecode:"
	mobileTokenKeyPrefix = "mobiletoken:"
)

type MobileUserInfo struct {
	Mobile   string
	Token    string
	Nickname string
	Portrait string
	Password string
	Sex      appgo.Sex
}

type MobileSupport interface {
	HasMobileUser(mobile string) (bool, error)
	GetMobileUser(mobile, password string) (uid appgo.Id, err error)
	GetMobileUserWithOutPwd(mobile string) (uid appgo.Id, err error)
	AddMobileUser(info *MobileUserInfo) (uid appgo.Id, err error)
	UpdatePwByMobile(mobile, password string) error
	SetMobileForUser(mobile string, userId appgo.Id) error
	appgo.MobileMsgSender
	appgo.KvStore
}

func MobilePreRegister(mobile string) (string, error) {
	if has, err := mobileSupport.HasMobileUser(mobile); err != nil {
		return "", err
	} else if has {
		return "", appgo.MobileUserAlreadyExistsErr
	}
	return sendSmsCode(mobile, 0, appgo.SmsTemplateRegister)
}

func MobileVerifyRegister(mobile, code string) (string, error) {
	token := crypto.RandNumStr(mobileTokenLen)
	if err := mobileSupport.Set(
		mobileTokenKeyPrefix+mobile, token, mobileTokenTimeout); err != nil {
		return "", err
	}
	return token, nil
}

func SMSPreLogin(mobile string) (string, error) {
	if mobileSupport == nil {
		return "", errors.New("mobile not supported")
	}

	return sendSmsCode(mobile, 0, appgo.SmsTemplateSMSLogin)
}

func SMSVerifyLogin(mobile, code string, role appgo.Role) (*LoginResult, error) {
	if mobileSupport == nil {
		return nil, errors.New("mobile not supported")
	}

	isNew := false
	uid, err := mobileSupport.GetMobileUserWithOutPwd(mobile)
	if err != nil {
		return nil, err
	}
	if uid == 0 {
		// create new user
		if uid, err = mobileSupport.AddMobileUser(&MobileUserInfo{Mobile: mobile, Nickname: mobile}); err != nil {
			return nil, err
		}
		isNew = true
	}
	loginRes, err := checkIn(uid, role)
	if err != nil {
		return nil, err
	}
	loginRes.IsNew = isNew

	return loginRes, nil
}

func MobilePreSet(mobile string, id appgo.Id) (string, error) {
	if has, err := mobileSupport.HasMobileUser(mobile); err != nil {
		return "", err
	} else if has {
		return "", appgo.MobileUserAlreadyExistsErr
	}
	return sendSmsCode(mobile, id, appgo.SmsTemplateSetMobile)
}

func MobilePreWxWebSet(mobile string, id appgo.Id) (string, error) {
	return sendSmsCode(mobile, id, appgo.SmsTemplateSetMobile)
}

func MobileVerifySet(mobile string, password string, id appgo.Id, code string) error {
	if err := mobileSupport.SetMobileForUser(mobile, id); err != nil {
		return err
	}
	// set password if provided
	if strings.TrimSpace(password) != "" {
		return mobileSupport.UpdatePwByMobile(mobile, password)
	} else {
		return nil
	}
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

func MobilePwReset(mobile string) (string, error) {
	if has, err := mobileSupport.HasMobileUser(mobile); err != nil {
		return "", err
	} else if !has {
		return "", appgo.MobileUserNotFoundErr
	}
	return sendSmsCode(mobile, 0, appgo.SmsTemplatePwReset)
}

func MobileVerifyPwReset(mobile, code, password string) error {
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

func sendSmsCode(mobile string, id appgo.Id, template appgo.SmsTemplate) (string, error) {
	code := crypto.RandNumStr(mobileCodeLen)
	if err := mobileSupport.Set(smsCodeKey(mobile, id), code, mobileCodeTimeout); err != nil {
		log.Errorln("sendSmsCode, mobileSupport.Set() failed, err: ", err)
		return "", err
	}
	return code, mobileSupport.SendMobileCode(mobile, template, code)
}

func verifySmsCode(mobile string, id appgo.Id, code string) error {
	scode, err := mobileSupport.Get(smsCodeKey(mobile, id))
	if err != nil {
		return err
	}
	if scode != code {
		return appgo.MobileUserBadCodeErr
	}
	return nil
}

func smsCodeKey(mobile string, id appgo.Id) string {
	key := mobileCodeKeyPrefix + mobile
	if id != 0 {
		key = mobileCodeKeyPrefix + mobile + "_" + id.String()
	}
	return key
}
