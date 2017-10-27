package auth

import (
	"errors"
	//log "github.com/Sirupsen/logrus"
	"github.com/oxfeeefeee/appgo"
	"github.com/oxfeeefeee/appgo/services/weixin"
)

func SetWeixin(uid appgo.Id, openId, token, code string, reset bool) error {
	if weixinSupport == nil {
		return errors.New("set weixin not supported")
	}
	if openId == "" || token == "" {
		params := &weixin.AccessTokenParams{*weixinAppInfo, code}
		at := weixin.GetAccessToken(params)
		if at == nil {
			return errors.New("Failed to get access token")
		}
		openId, token = at.OpenId, at.AccessToken
	}
	winfo := weixin.GetUserInfo(openId, token)
	if winfo == nil {
		return errors.New("Failed to get weixin user info")
	}
	wxuid, err := weixinSupport.GetWeixinUser(winfo.UnionId)
	if err != nil {
		return err
	}
	if wxuid == 0 {
		if err := weixinSupport.SetWeixinForUser(uid, winfo.UnionId); err != nil {
			return err
		}
		return nil
	} else if wxuid == uid {
		return nil
	} else if reset {
		if err := weixinSupport.EmptyAndSetWxForUser(wxuid, uid, winfo.UnionId); err != nil {
			return err
		}
		return nil
	}

	if wxMobile, err := userSystem.GetUserMobile(wxuid); err != nil {
		return err
	} else {
		if len(wxMobile) != 0 {
			return appgo.WeixinUserWithAnotherMobileErr
		}
	}

	return appgo.WeixinUserAlreadyExistsErr
}

func EmptyWeixin(uid appgo.Id) error {
	return weixinSupport.EmptyWeixinForUser(uid)
}
