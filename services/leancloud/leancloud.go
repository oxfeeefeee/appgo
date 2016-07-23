package leancloud

import (
	"github.com/parnurzeal/gorequest"
)

const (
	sendMethod = "POST"
	sendUrl    = "https://leancloud.cn/1.1/push"
)

var (
	appId   string
	appKey  string
	devMode bool
)

func init() {
	appId = appgo.Conf.Leancloud.Id
	appKey = appgo.Conf.Leancloud.Key
	devMode = appgo.Conf.DevMode
}

type LeancloudPush struct {
	Channels string
}

func request() {

}
