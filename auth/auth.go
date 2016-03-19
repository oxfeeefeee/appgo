package auth

import (
	"errors"
	log "github.com/Sirupsen/logrus"
	"github.com/oxfeeefeee/appgo"
	"github.com/oxfeeefeee/appgo/services/weibo"
	"github.com/oxfeeefeee/appgo/services/weixin"
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
	// extraInfo should be ban info if banned is true, userInfo otherwise
	CheckIn(id appgo.Id, role appgo.Role,
		newToken Token) (banned bool, extraInfo interface{}, err error)
}

type WeixinSupport interface {
	GetWeixinUser(unionId string) (uid appgo.Id, err error)
	AddWeixinUser(info *weixin.UserInfo) (uid appgo.Id, err error)
}

type WeiboSupport interface {
	GetWeiboUser(unionId string) (uid appgo.Id, err error)
	AddWeiboUser(info *weibo.UserInfo) (uid appgo.Id, err error)
}

func Init(us UserSystem, wx WeixinSupport, wb WeiboSupport) {
	userSystem = us
	weixinSupport = wx
	weiboSupport = wb
	if wx != nil {
		weixinAppInfo = &weixin.AppInfo{
			appgo.Conf.Weixin.AppId,
			appgo.Conf.Weixin.Secret,
		}
		if weixinAppInfo.AppId == "" || weixinAppInfo.Secret == "" {
			log.Panicln("Bad weixin config")
		}
	}
	if wb != nil {
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

func LoginByWeixin(openId, token, code string, role appgo.Role) (*LoginResult, error) {
	if weixinSupport == nil {
		return nil, errors.New("weixin login not supported")
	}
	if openId == "" || token == "" {
		params := &weixin.AccessTokenParams{*weixinAppInfo, code}
		at := weixin.GetAccessToken(params)
		if at == nil {
			return nil, errors.New("Failed to get access token")
		}
		openId, token = at.OpenId, at.AccessToken
	}
	winfo := weixin.GetUserInfo(openId, token)
	if winfo == nil {
		return nil, errors.New("Failed to get weixin user info")
	}
	uid, err := weixinSupport.GetWeixinUser(winfo.UnionId)
	if err != nil {
		return nil, err
	}
	if uid == 0 {
		uid, err = weixinSupport.AddWeixinUser(winfo)
		if err != nil {
			return nil, err
		}
	}
	return checkIn(uid, role)
}

func LoginByWeibo(openId, token, code string, role appgo.Role) (*LoginResult, error) {
	if weiboSupport == nil {
		return nil, errors.New("weibo login not supported")
	}
	if openId == "" || token == "" {
		params := &weibo.AccessTokenParams{*weiboAppInfo, code}
		at := weibo.GetAccessToken(params)
		if at == nil {
			return nil, errors.New("Failed to get access token")
		}
		openId, token = at.Id, at.AccessToken
	}
	uid, err := weiboSupport.GetWeiboUser(openId)
	if err != nil {
		return nil, err
	}
	if uid == 0 {
		winfo := weibo.GetUserInfo(openId, token)
		if winfo == nil {
			return nil, errors.New("Failed to get weibo user info")
		}
		uid, err = weiboSupport.AddWeiboUser(winfo)
		if err != nil {
			return nil, err
		}
	}
	return checkIn(uid, role)
}

func checkIn(uid appgo.Id, role appgo.Role) (*LoginResult, error) {
	tokenlf := tokenLifetime(role)
	token := NewToken(uid, role, tokenlf)
	banned, info, err := userSystem.CheckIn(uid, role, token)
	if err != nil {
		return nil, err
	}
	if banned {
		return &LoginResult{
			UserId:  uid,
			Banned:  true,
			BanInfo: info,
		}, nil
	}
	return &LoginResult{
		UserId:   uid,
		Token:    token,
		UserInfo: info,
	}, nil
}

func tokenLifetime(role appgo.Role) int {
	switch role {
	case appgo.RoleAppUser:
		return appgo.Conf.TokenLifetime.AppUser
	case appgo.RoleWebUser:
		return appgo.Conf.TokenLifetime.WebUser
	case appgo.RoleWebAdmin:
		return appgo.Conf.TokenLifetime.WebAdmin
	default:
		return 0
	}
}
