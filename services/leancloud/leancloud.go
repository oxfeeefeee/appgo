package leancloud

import (
	log "github.com/Sirupsen/logrus"
	"github.com/oxfeeefeee/appgo"
	"github.com/parnurzeal/gorequest"
	"strings"
)

const (
	sendMethod      = "POST"
	sendUrl         = "https://leancloud.cn/1.1/push"
	push_batch_size = 5000
)

var (
	appId         string
	appKey        string
	androidAction string
	expiration    int
	prod          string
)

func init() {
	appId = appgo.Conf.Leancloud.AppId
	appKey = appgo.Conf.Leancloud.AppKey
	androidAction = appgo.Conf.Leancloud.AndroidAction
	expiration = appgo.Conf.Leancloud.ExpirationInterval
	prod = "prod"
	if appgo.Conf.DevMode {
		prod = "dev"
	}
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
	Badge            int           `json:"badge,omitempty"`
	Sound            string        `json:"sound,omitempty"`
	ContentAvailable int           `json:"content-available,omitempty"`
}

type AndroidData struct {
	Alert     string      `json:"alert,omitempty"`
	Title     string      `json:"title,omitempty"`
	Action    string      `json:"action,omitempty"`
	PlaySound bool        `json:"play_sound,omitempty"`
	Custom    interface{} `json:"custom,omitempty"`
}

type IosAndroidData struct {
	Ios     *IosData     `json:"ios,omitempty"`
	Android *AndroidData `json:"android,omitempty"`
}

type LeancloudPush struct {
	ExpirationInterval int             `json:"expiration_interval,omitempty"`
	Prod               string          `json:"prod,omitempty"`
	Data               *IosAndroidData `json:"data,omitempty"`
	Cql                string          `json:"cql,omitempty"`
}

type Leancloud struct{}

func (_ Leancloud) Name() string {
	return "leancloud"
}

func (l Leancloud) PushNotif(pushInfo map[appgo.Id]*appgo.PushInfo, content *appgo.PushData) {
	userIds := make([]string, 0, len(pushInfo))
	for uid, _ := range pushInfo {
		userIds = append(userIds, uid.String())
	}
	for i := 0; i < len(userIds); i += push_batch_size {
		end := i + push_batch_size
		if end > len(userIds) {
			end = len(userIds)
		}
		go l.doPushNotif(userIds[i:end], content)
	}
}

func (_ Leancloud) doPushNotif(ids []string, content *appgo.PushData) {
	defer func() {
		if r := recover(); r != nil {
			log.Errorln("doPushNotif paniced: ", r)
		}
	}()
	idstr := strings.Join(ids, ",")
	cql := "select * from _Installation where userId in ( " + idstr + " )"
	pl := buildPayload(content)
	request(&LeancloudPush{
		ExpirationInterval: expiration,
		Prod:               prod,
		Data:               pl,
		Cql:                cql,
	})
}

func buildPayload(content *appgo.PushData) *IosAndroidData {
	return &IosAndroidData{
		&IosData{
			Alert: &IosApnsAlert{
				Title: content.Title,
				Body:  content.Message,
			},
			Badge: content.Badge,
			Sound: content.Sound,
		},
		&AndroidData{
			Alert:     content.Message,
			Title:     content.Title,
			Action:    androidAction,
			PlaySound: len(content.Sound) > 0,
			Custom:    content.Custom,
		},
	}
}

func request(p *LeancloudPush) {
	_, ret, errs := gorequest.New().SetDebug(true).
		Post(sendUrl).
		Set("X-LC-Id", appId).
		Set("X-LC-Key", appKey).
		SendStruct(p).
		End()
	if errs != nil {
		log.WithFields(log.Fields{
			"errors": errs,
			"url":    sendUrl,
			"body":   ret,
		}).Error("Failed to request leancloud push")
		return
	}
	log.Infoln("leancloud push return: ", ret)
}
