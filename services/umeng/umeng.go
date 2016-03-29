package umeng

import (
	"encoding/hex"
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"github.com/oxfeeefeee/appgo"
	"github.com/oxfeeefeee/appgo/toolkit/crypto"
	"github.com/oxfeeefeee/appgo/toolkit/strutil"
	"github.com/parnurzeal/gorequest"
	"strings"
	"time"
)

const (
	sendMethod = "POST"
	sendUrl    = "http://msg.umeng.com/api/send"
)

var (
	androidAppKey string
	androidSecret string
	iosAppKey     string
	iosSecret     string
	devMode       bool
)

type AndroidBody struct {
	Ticker      string      `json:"ticker,omitempty"`
	Title       string      `json:"title,omitempty"`
	Text        string      `json:"text,omitempty"`
	Sound       string      `json:"sound,omitempty"`
	PlayVibrate bool        `json:"play_vibrate,omitempty"`
	PlayLights  bool        `json:"play_lights,omitempty"`
	PlaySound   bool        `json:"play_sound,omitempty"`
	AfterOpen   string      `json:"after_open,omitempty"`
	Url         string      `json:"url,omitempty"`
	Activity    string      `json:"activity,omitempty"`
	Custom      interface{} `json:"custom,omitempty"`
}

type IosApnsAlert struct {
	Title        string   `json:"title,omitempty"`
	Body         string   `json:"body,omitempty"`
	ActionLocKey string   `json:"action-loc-key,omitempty"`
	LocKey       string   `json:"loc-key,omitempty"`
	LocArgs      []string `json:"loc-args,omitempty"`
	LaunchImage  string   `json:"launch-image,omitempty"`
}

type IosApns struct {
	Alert            *IosApnsAlert `json:"alert,omitempty"`
	Badge            int           `json:"badge,omitempty"`
	Sound            string        `json:"sound,omitempty"`
	ContentAvailable int           `json:"content-available,omitempty"`
	Category         string        `json:"category,omitempty"`
}

type Data struct {
	AppKey         string      `json:"appkey,omitempty"`
	Timestamp      string      `json:"timestamp,omitempty"`
	Type           string      `json:"type,omitempty"`
	DeviceTokens   string      `json:"device_tokens,omitempty"`
	AliasType      string      `json:"alias_type,omitempty"`
	FileId         string      `json:"file_id,omitempty"`
	Filter         string      `json:"filter,omitempty"`
	Payload        interface{} `json:"payload,omitempty"`
	ProductionMode bool        `json:"production_mode"`
	Description    string      `json:"Description,omitempty"`
	ThirdPartyId   string      `json:"thirdparty_id,omitempty"`
}

type Umeng struct{}

func init() {
	androidAppKey = appgo.Conf.Umeng.Android.AppKey
	androidSecret = appgo.Conf.Umeng.Android.AppMasterSecret
	iosAppKey = appgo.Conf.Umeng.Ios.AppKey
	iosSecret = appgo.Conf.Umeng.Ios.AppMasterSecret
	devMode = appgo.Conf.DevMode
}

func (_ Umeng) Name() string {
	return "umeng"
}

func (_ Umeng) PushNotif(platform appgo.Platform, tokens []string, content *appgo.PushData) {
	var payload interface{}
	var appkey, secret string
	if platform == appgo.PlatformIos {
		payload = buildIosPayload(content)
		appkey, secret = iosAppKey, iosSecret
	} else if platform == appgo.PlatformAndroid {
		payload = buildAndroidPayload(content)
		appkey, secret = androidAppKey, androidSecret
	}
	if payload == nil {
		return
	}
	tokenstr := strings.Join(tokens, ",")
	data := getAllData(appkey, tokenstr, payload)
	jdata, err := json.Marshal(data)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("Failed to marshal umeng push request")
		return
	}
	sig := sign(string(jdata), secret)
	request(string(jdata), sig)
}

func buildIosPayload(content *appgo.PushData) interface{} {
	ret := make(map[string]interface{})
	for k, v := range content.Custom {
		ret[k] = v
	}
	ret["aps"] = &IosApns{
		Badge: content.Badge,
		Sound: content.Sound,
		Alert: &IosApnsAlert{
			Title: content.Title,
			Body:  content.Message,
		},
	}
	for k, v := range content.Custom {
		ret[k] = v
	}
	return ret
}

func buildAndroidPayload(content *appgo.PushData) interface{} {
	ret := make(map[string]interface{})
	ret["display_type"] = "notification"
	ret["body"] = &AndroidBody{
		Ticker:    content.Message,
		Title:     content.Title,
		Text:      content.Message,
		Sound:     content.Sound,
		PlaySound: len(content.Sound) > 0,
		AfterOpen: "go_custom",
		Custom:    content.Custom,
	}
	return ret
}

func getAllData(appkey, tokens string, payload interface{}) *Data {
	return &Data{
		AppKey:         appkey,
		Timestamp:      strutil.FromInt64(time.Now().Unix()),
		DeviceTokens:   tokens,
		Payload:        payload,
		Type:           "listcast",
		ProductionMode: !devMode,
	}
}

func request(body, sig string) {
	url := sendUrl + "?sign=" + sig
	req := gorequest.New().Post(url).Type("json")
	req.RawString = body
	req.BounceToRawString = true
	_, ret, errs := req.End()
	if errs != nil {
		log.WithFields(log.Fields{
			"errors": errs,
			"url":    url,
			"body":   ret,
		}).Error("Failed to request umeng push")
	}
}

func sign(body, secret string) string {
	ss := []string{sendMethod, sendUrl, body, secret}
	data := []byte(strings.Join(ss, ""))
	s := hex.EncodeToString(crypto.Md5(data))
	return s
}
