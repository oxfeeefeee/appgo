package appgo

import (
	log "github.com/Sirupsen/logrus"
	"github.com/jinzhu/configor"
	"os"
	"path/filepath"
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
	Redis struct {
		Host        string
		Port        string
		Password    string
		MaxIdle     int
		IdleTimeout int
	}
	Negroni struct {
		Port string
		GZip bool
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
	WeixinJssdk struct {
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

func initConfig() {
	filePath := searchConfigFile()
	err := configor.Load(&Conf, filePath)
	if err != nil {
		log.WithField("error", err).Panicln("Failed to load config file")
	}
	if len(Conf.RootKey) != 16 {
		log.Println("bad root key size")
	}
}

func searchConfigFile() string {
	// Search up for config file
	left, upALevel, right, result := "./", "../", "conf/appgo.yml", ""
	for i := 0; i < 5; i++ {
		p := filepath.Join(left, right)
		if _, err := os.Stat(p); err == nil {
			result = p
			break
		}
		left = filepath.Join(left, upALevel)
	}
	return result
}
