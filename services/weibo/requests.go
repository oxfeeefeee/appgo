package weibo

import (
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"github.com/parnurzeal/gorequest"
	"net/url"
)

var (
	appId       string
	secret      string
	redirectUrl string
)

type UserInfo struct {
	Id    string `json:"idstr"`
	Name  string `json:"name"`
	Image string `json:"profile_image_url"`
	Sex   string `json:gender`
}

type AppInfo struct {
	AppId       string
	Secret      string
	RedirectUrl string
}

type AccessTokenParams struct {
	AppInfo
	Code string
}

type AccessTokenResult struct {
	AccessToken string `json:"access_token"`
	Id          string `json:"uid"`
}

type apiError struct {
	ErrCode int    `json:"error_code"`
	ErrMsg  string `json:"error"`
}

func GetAccessToken(params *AccessTokenParams) *AccessTokenResult {
	var result struct {
		AccessTokenResult
		apiError
	}
	url := accessTokenUrl(params)
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
	_, body, errs := gorequest.New().Get(url).End()
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
	u.Host = "api.weibo.com"
	u.Path = "oauth2/access_token"
	q := u.Query()
	q.Set("code", params.Code)
	q.Set("redirect_uri", params.RedirectUrl)
	q.Set("client_id", params.AppId)
	q.Set("client_secret", params.Secret)
	q.Set("grant_type", "authorization_code")
	u.RawQuery = q.Encode()
	return u.String()
}

func userInfoUrl(id, token string) string {
	var u url.URL
	u.Scheme = "https"
	u.Host = "api.weibo.com"
	u.Path = "2/users/show.json"
	q := u.Query()
	q.Set("uid", id)
	q.Set("access_token", token)
	u.RawQuery = q.Encode()
	return u.String()
}
