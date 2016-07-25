package jssdk

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"github.com/oxfeeefeee/appgo"
	"github.com/oxfeeefeee/appgo/toolkit/crypto"
	"time"
)

var (
	appId          string
	appSecret      string
	kvStore        appgo.KvStore
	tokenStoreKey  = "wxjssdk_token"
	ticketStoreKey = "wxticket_token"
	signFormat     = "jsapi_ticket=%s&noncestr=%s&timestamp=%d&url=%s"
	accessTokenUrl = `https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s`
	ticketUrl      = `https://api.weixin.qq.com/cgi-bin/ticket/getticket?type=jsapi&access_token=%s`
)

type WxConfig struct {
	Debug     bool     `json:"debug"`
	AppId     string   `json:"appId"`
	Timestamp int64    `json:"timestamp"`
	NonceStr  string   `json:"nonceStr"`
	Signature string   `json:"signature"`
	JsApiList []string `json:"jsApiList"`
}

func Init(kvs appgo.KvStore) {
	appId = appgo.Conf.WeixinJssdk.AppId
	appSecret = appgo.Conf.WeixinJssdk.Secret
	kvStore = kvs
}

func GetConfig(url string) *WxConfig {
	config := &WxConfig{
		AppId:     appId,
		Timestamp: time.Now().Unix(),
		NonceStr:  crypto.RandNumStr(15),
	}
	ticket := getTicket()
	sign(config, ticket, url)
	return config
}

func sign(c *WxConfig, ticket, url string) {
	str := fmt.Sprintf(signFormat, ticket, c.NonceStr, c.Timestamp, url)
	sha := sha1.Sum([]byte(str))
	c.Signature = hex.EncodeToString(sha[:])
}
