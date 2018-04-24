package appgo

import (
//"strconv"
)

const AnonymousId = 6666

const InternalTestId = 7777

const InternalTestToken = "sjadfjlksadfjkljfwoeifshgsdhgsldfjf"

const CustomTokenHeaderName = "X-Appgo-Token"

const CustomWallStTokenHeaderName = "X-Ivanka-Token"

const CustomAuthorTokenHeaderName = "X-Appgo-Author-Token"

const CustomVersionHeaderName = "X-Appgo-Api-Version"

const CustomConfVerHeaderName = "X-Appgo-Conf-Version"

const CustomAppVerHeaderName = "X-Appgo-App-Version"

const CustomPlatformHeaderName = "X-Appgo-Platform"

const (
	RoleAppUser  Role = 100
	RoleWebUser       = 101
	RoleAuthor        = 150
	RoleWebAdmin      = 200
)

type Role int

const (
	SexDefault Sex = iota
	SexMale
	SexFemale
)

type Sex int8

const (
	_ SmsTemplate = iota
	SmsTemplateRegister
	SmsTemplatePwReset
	SmsTemplateSetMobile
	SmsTemplateSMSLogin
)

type SmsTemplate int

const (
	_ Platform = iota
	PlatformIos
	PlatformAndroid
)

type Platform int8

type DummyInput struct{}

type KvStore interface {
	Set(k, v string, timeout int) error
	Get(k string) (string, error)
}

type MobileMsgSender interface {
	SendMobileCode(mobile string, template SmsTemplate, code string) error
}

type PushInfo struct {
	Platform     Platform
	Manufacturer string
	Provider     string
	Token        string
}

type PushData struct {
	Title       string
	Message     string
	Sound       string
	Badge       int
	PlayVibrate bool
	PlayLights  bool
	PlaySound   bool
	Custom      map[string]interface{}
}

type Pusher interface {
	Name() string
	PushNotif(pushInfo map[Id]*PushInfo, content *PushData)
}

func init() {
	initConfig()
	initLog()
}
