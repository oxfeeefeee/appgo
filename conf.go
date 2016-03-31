package appgo

import (
	log "github.com/Sirupsen/logrus"
	"github.com/jinzhu/configor"
)

var Conf struct {
	DevMode      bool
	LogLevel     log.Level
	RootKey      string
	TemplatePath string
	Mysql        struct {
		Host     string
		Port     string
		User     string
		Password string
		DbName   string
		Charset  string
	}
	Negroni struct {
		Port string
	}
	Cors struct {
		AllowedOrigins     string
		AllowedMethods     string
		AllowedHeaders     string
		OptionsPassthrough bool
		Debug              bool
	}
	TokenLifetime struct {
		AppUser  int
		WebUser  int
		WebAdmin int
	}
	Weixin struct {
		AppId  string
		Secret string
	}
	Weibo struct {
		AppId       string
		Secret      string
		RedirectUrl string
	}
	Qq struct {
		AppId string
	}
	Qiniu struct {
		AccessKey      string
		Secret         string
		Bucket         string
		Domain         string
		DefaultExpires int
	}
	Umeng struct {
		Android struct {
			AppKey          string
			AppMasterSecret string
		}
		Ios struct {
			AppKey          string
			AppMasterSecret string
		}
	}
}

func init() {
	err := configor.Load(&Conf, "./conf/appgo.yml")
	if err != nil {
		log.WithField("error", err).Panicln("Failed to load config file")
	}
	if len(Conf.RootKey) != 16 {
		log.Println("bad root key size")
	}
}
