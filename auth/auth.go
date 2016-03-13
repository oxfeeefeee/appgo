package auth

import (
	log "github.com/Sirupsen/logrus"
	"github.com/oxfeeefeee/appgo"
	"github.com/oxfeeefeee/appgo/services/weibo"
	"github.com/oxfeeefeee/appgo/services/weixin"
	"time"
)

var (
	weixinAppInfo *weixin.AppInfo
	weiboAppInfo  *weibo.AppInfo

	userSystem    UserSystem
	weixinSupport WeixinSupport
	weiboSupport  WeiboSupport
)

type LoginResult struct {
	UserId   appgo.Id
	Token    Token
	UserInfo interface{}
	Banned   bool
	BanInfo  interface{}
}

type UserSystem interface {
	GetBanInfo(id appgo.Id) (banned bool, banInfo interface{})
	TokenLifetime(role Role) (seconds int)
	UpdateToken(id appgo.Id, role Role, token Token) error
}

type WeixinSupport interface {
	GetWeixinUser(unionId string) (uid appgo.Id, uinfo interface{}, err error)
	AddWeixinUser(info *weixin.UserInfo) (uid appgo.Id, uinfo interface{}, err error)
}

type WeiboSupport interface {
	GetWeiboUser(unionId string) (uid appgo.Id, uinfo interface{}, err error)
	AddWeiboUser(info *weibo.UserInfo) (uid appgo.Id, uinfo interface{}, err error)
}

func Init(us UserSystem, weixin WeiboSupport, weibo WeiboSupport) {
	userSystem = us
	weixinSupport = weixin
	weiboSupport = weibo
	if weixin != nil {
		weixinAppInfo = &weixin.AppInfo{
			appgo.Conf.Weixin.AppId,
			appgo.Conf.Weixin.Secret,
		}
		if weixinAppInfo.AppId == "" || weixinAppInfo.Secret == "" {
			log.Panicln("Bad weixin config")
		}
	}
	if weibo != nil {
		weiboAppInfo = &weibo.AppInfo{
			appgo.Conf.Weibo.AppId,
			appgo.Conf.Weibo.Secret,
			appgo.Conf.Weibo.RedirectUrl,
		}
		if weiboAppInfo.AppId == "" || weiboAppInfo.Secret == "" {
			log.Panicln("Bad weibo config")
		}
	}
}

func login(uid appgo.Id) (*LoginResult, error) {
	banned, banInfo := userSystem.GetBanInfo(uid)
	if banned {
		return &LoginResult{
			UserId:  uid,
			Banned:  true,
			BanInfo: banInfo,
		}, nil
	}
	// todo: generate token
}
