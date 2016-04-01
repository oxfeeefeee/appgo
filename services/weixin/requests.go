package weixin

import (
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"github.com/parnurzeal/gorequest"
	"net/url"
)

type UserInfo struct {
	Id       string `json:"openid"`
	UnionId  string `json:"unionid"`
	Nickname string `json:"nickname"`
	Image    string `json:"headimgurl"`
	Sex      int    `json:"sex"`
	Province string `json:"province"`
	City     string `json:"city"`
	Country  string `json:"country"`
}

type AppInfo struct {
	AppId  string
	Secret string
}

type AccessTokenParams struct {
	AppInfo
	Code string
}

type AccessTokenResult struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	OpenId       string `json:"openid"`
	UnionId      string `json:"unionid"`
}

type apiError struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

func GetAccessToken(params *AccessTokenParams) *AccessTokenResult {
	var result struct {
		AccessTokenResult
		apiError
	}
	url := accessTokenUrl(params)
	_, body, errs := gorequest.New().Get(url).EndBytes()
	if errs != nil {
		log.WithFields(log.Fields{
			"errors": errs,
			"url":    url,
		}).Error("Failed to get access token")
		return nil
	}
	if err := json.Unmarshal(body, &result); err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"body":  string(body),
		}).Error("Failed to unmarshal access token result")
		return nil
	} else if result.ErrCode != 0 {
		log.WithFields(log.Fields{
			"errCode": result.ErrCode,
			"errMsg":  result.ErrMsg,
		}).Error("GetAccessToken api returns error")
		return nil
	}
	return &result.AccessTokenResult
}

func GetUserInfo(id, token string) *UserInfo {
	var uinfo struct {
		UserInfo
		apiError
	}
	url := userInfoUrl(id, token)
	_, body, errs := gorequest.New().Get(url).EndBytes()
	if errs != nil {
		log.WithFields(log.Fields{
			"errors": errs,
			"url":    url,
		}).Error("Failed to get user info")
		return nil
	}
	if err := json.Unmarshal(body, &uinfo); err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"body":  string(body),
		}).Error("Failed to unmarshal user info")
		return nil
	} else if uinfo.ErrCode != 0 {
		log.WithFields(log.Fields{
			"errCode": uinfo.ErrCode,
			"errMsg":  uinfo.ErrMsg,
		}).Error("GetUserInfo api returns error")
		return nil
	}
	return &uinfo.UserInfo
}

func accessTokenUrl(params *AccessTokenParams) string {
	var u url.URL
	u.Scheme = "https"
	u.Host = "api.weixin.qq.com"
	u.Path = "sns/oauth2/access_token"
	q := u.Query()
	q.Set("appid", params.AppId)
	q.Set("secret", params.Secret)
	q.Set("code", params.Code)
	q.Set("grant_type", "authorization_code")
	u.RawQuery = q.Encode()
	return u.String()
}

func userInfoUrl(id, token string) string {
	var u url.URL
	u.Scheme = "https"
	u.Host = "api.weixin.qq.com"
	u.Path = "sns/userinfo"
	q := u.Query()
	q.Set("access_token", token)
	q.Set("openid", id)
	u.RawQuery = q.Encode()
	return u.String()
}
