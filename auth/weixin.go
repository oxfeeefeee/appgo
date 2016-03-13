package auth

import (
	"errors"
	"github.com/oxfeeefeee/appgo/services/weixin"
)

func LoginByWeixin(code string) (*LoginResult, error) {
	if weixinSupport == nil {
		return nil, errors.New("weixin login not supported")
	}
	params := &weixin.AccessTokenParams{weixinAppInfo, code}
	at := weixin.GetAccessToken(params)
	if at == nil {
		return nil, errors.New("Failed to get access token")
	}
	uid, uinfo, err := weixinSupport.GetWeixinUser(at.UnionId)
	if err != nil {
		return nil, err
	}
	if uid == 0 {
		winfo = weixin.GetUserInfo(at.OpenId, at.AccessToken)
		if winfo == nil {
			return nil, errors.New("Failed to get weixin user info")
		}
		uid, uinfo, err = weixinSupport.AddWeixinUser(winfo)
		if err != nil {
			return nil, err
		}
	}
	// todo: call login
}
