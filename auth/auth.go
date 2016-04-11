package auth

import (
	"errors"
	log "github.com/Sirupsen/logrus"
	"github.com/oxfeeefeee/appgo"
	"github.com/oxfeeefeee/appgo/services/qq"
	"github.com/oxfeeefeee/appgo/services/weibo"
	"github.com/oxfeeefeee/appgo/services/weixin"
)

var (
	weixinAppInfo *weixin.AppInfo
	weiboAppInfo  *weibo.AppInfo
	qqAppInfo     *qq.AppInfo

	userSystem    UserSystem
	weixinSupport WeixinSupport
	weiboSupport  WeiboSupport
	qqSupport     QqSupport
	mobileSupport MobileSupport
	oauthSupport  OAuthSupport
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

type QqSupport interface {
	GetQqUser(openId string) (uid appgo.Id, err error)
	AddQqUser(info *qq.UserInfo) (uid appgo.Id, err error)
}

type OAuthSupport interface {
	GetOAuthUser(index int, id string) (uid appgo.Id, err error)
	GetOAuthUserInfo(index int, code string) (interface{}, string, error)
	AddOAuthUser(index int, id string, info interface{}) (uid appgo.Id, err error)
}

func Init(us UserSystem, wx WeixinSupport, wb WeiboSupport,
	qqsp QqSupport, mobile MobileSupport, oauth OAuthSupport) {
	userSystem = us
	weixinSupport = wx
	weiboSupport = wb
	qqSupport = qqsp
	mobileSupport = mobile
	oauthSupport = oauth
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
	if qqsp != nil {
		qqAppInfo = &qq.AppInfo{
			appgo.Conf.Qq.AppId,
		}
		if qqAppInfo.AppId == "" {
			log.Panicln("Bad qq config")
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

func LoginByQq(openId, token string, role appgo.Role) (*LoginResult, error) {
	if qqSupport == nil {
		return nil, errors.New("qq login not supported")
	}
	if openId == "" || token == "" {
		return nil, errors.New("LoginByQq: bad id or token")
	}
	uid, err := qqSupport.GetQqUser(openId)
	if err != nil {
		return nil, err
	}
	if uid == 0 {
		winfo := qq.GetUserInfo(qqAppInfo, openId, token)
		if winfo == nil {
			return nil, errors.New("Failed to get qq user info")
		}
		uid, err = qqSupport.AddQqUser(winfo)
		if err != nil {
			return nil, err
		}
	}
	return checkIn(uid, role)
}

func LoginByOAuth(code string, index int, role appgo.Role) (*LoginResult, error) {
	if oauthSupport == nil {
		return nil, errors.New("OAuth login not supported")
	}
	info, id, err := oauthSupport.GetOAuthUserInfo(index, code)
	if err != nil {
		return nil, err
	}
	uid, err := oauthSupport.GetOAuthUser(index, id)
	if err != nil {
		return nil, err
	}
	if uid == 0 {
		uid, err = oauthSupport.AddOAuthUser(index, id, info)
		if err != nil {
			return nil, err
		}
	}
	return checkIn(uid, role)
}

func checkIn(uid appgo.Id, role appgo.Role) (*LoginResult, error) {
	token := NewToken(uid, role)
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
