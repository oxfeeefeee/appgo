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
	sendMethod      = "POST"
	sendUrl         = "http://msg.umeng.com/api/send"
	push_batch_size = 450
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
	PlayVibrate string      `json:"play_vibrate,omitempty"`
	PlayLights  string      `json:"play_lights,omitempty"`
	PlaySound   string      `json:"play_sound,omitempty"`
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

func (u Umeng) PushNotif(pushInfo map[appgo.Id]*appgo.PushInfo, content *appgo.PushData) {
	iosTokens := make([]string, 0, len(pushInfo))
	androidTokens := make([]string, 0, len(pushInfo))
	for _, pi := range pushInfo {
		if pi.Platform == appgo.PlatformIos {
			iosTokens = append(iosTokens, pi.Token)
		} else if pi.Platform == appgo.PlatformAndroid {
			androidTokens = append(androidTokens, pi.Token)
		}
	}
	iosPayload := buildIosPayload(content)
	androidPayload := buildAndroidPayload(content)
	if iosPayload != nil && len(iosTokens) > 0 {
		for i := 0; i < len(iosTokens); i += push_batch_size {
			end := i + push_batch_size
			if end > len(iosTokens) {
				end = len(iosTokens)
			}
			go u.doPushNotif(iosTokens[i:end], iosPayload, iosAppKey, iosSecret)
		}
	}
	if androidPayload != nil && len(androidTokens) > 0 {
		for i := 0; i < len(androidTokens); i += push_batch_size {
			end := i + push_batch_size
			if end > len(androidTokens) {
				end = len(androidTokens)
			}
			go u.doPushNotif(androidTokens[i:end], androidPayload, androidAppKey, androidSecret)
		}
	}
}

func (_ Umeng) doPushNotif(tokens []string, payload interface{}, appkey, secret string) {
	defer func() {
		if r := recover(); r != nil {
			log.Errorln("doPushNotif paniced: ", r)
		}
	}()
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
	ps := "false"
	if len(content.Sound) > 0 {
		ps = "true"
	}
	ret["body"] = &AndroidBody{
		Ticker:      content.Message,
		Title:       content.Title,
		Text:        content.Message,
		Sound:       content.Sound,
		PlayVibrate: "false",
		PlayLights:  "false",
		PlaySound:   ps,
		AfterOpen:   "go_custom",
		Custom:      content.Custom,
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
	_, ret, errs := req.EndBytes()
	if errs != nil {
		log.WithFields(log.Fields{
			"errors": errs,
			"url":    url,
			"body":   ret,
		}).Error("Failed to request umeng push")
	}
	retval := struct {
		Ret  string      `json:"ret"`
		Data interface{} `json:"data"`
	}{}
	json.Unmarshal(ret, &retval)
	if retval.Ret != "SUCCESS" {
		log.WithField("error", retval.Data).Error("Umeng push returns error")
	}
}

func sign(body, secret string) string {
	ss := []string{sendMethod, sendUrl, body, secret}
	data := []byte(strings.Join(ss, ""))
	s := hex.EncodeToString(crypto.Md5(data))
	return s
}
