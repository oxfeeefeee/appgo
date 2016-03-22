package qq

import (
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"github.com/parnurzeal/gorequest"
	"net/url"
)

type UserInfo struct {
	OpenId   string
	Nickname string `json:"nickname"`
	Image    string `json:"figureurl_qq_2"`
	Sex      string `json:"gender"`
	Province string `json:"province"`
	City     string `json:"city"`
}

type AppInfo struct {
	AppId string
}

type apiError struct {
	ErrCode int    `json:"ret"`
	ErrMsg  string `json:"msg"`
}

func GetUserInfo(appInfo *AppInfo, id, token string) *UserInfo {
	var uinfo struct {
		UserInfo
		apiError
	}
	url := userInfoUrl(appInfo.AppId, id, token)
	_, body, errs := gorequest.New().Get(url).End()
	if errs != nil {
		log.WithFields(log.Fields{
			"errors": errs,
			"url":    url,
		}).Error("Failed to get user info")
		return nil
	}
	if err := json.Unmarshal([]byte(body), &uinfo); err != nil {
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
	uinfo.UserInfo.OpenId = id
	return &uinfo.UserInfo
}

func userInfoUrl(appid, id, token string) string {
	var u url.URL
	u.Scheme = "https"
	u.Host = "graph.qq.com"
	u.Path = "user/get_user_info"
	q := u.Query()
	q.Set("oauth_consumer_key", appid)
	q.Set("access_token", token)
	q.Set("openid", id)
	q.Set("format", "json")
	u.RawQuery = q.Encode()
	return u.String()
}
