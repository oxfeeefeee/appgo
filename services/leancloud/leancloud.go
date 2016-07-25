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

type IosApnsAlert struct {
	Title        string   `json:"title,omitempty"`
	Body         string   `json:"body,omitempty"`
	ActionLocKey string   `json:"action-loc-key,omitempty"`
	LocKey       string   `json:"loc-key,omitempty"`
	LocArgs      []string `json:"loc-args,omitempty"`
	LaunchImage  string   `json:"launch-image,omitempty"`
}

type IosData struct {
	Alert            *IosApnsAlert `json:"alert,omitempty"`
	Badge            string        `json:"badge,omitempty"`
	Sound            string        `json:"sound,omitempty"`
	ContentAvailable int           `json:"content-available,omitempty"`
}

type AndroidData struct {
	Alert  string      `json:"alert,omitempty"`
	Title  string      `json:"title,omitempty"`
	Custom interface{} `json:"custom,omitempty"`
}

type IosAndroidData struct {
	Ios     *IosData     `json:"ios,omitempty"`
	Android *AndroidData `json:"android,omitempty"`
}

type LeancloudPush struct {
	ExpirationInterval string          `json:"expiration_interval,omitempty"`
	Prod               string          `json:"prod,omitempty"`
	Data               *IosAndroidData `json:"data,omitempty"`
	Cql                string          `json:"cql,omitempty"`
}

func request() {

}
