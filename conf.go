package appgo

import (
	log "github.com/Sirupsen/logrus"
	"github.com/jinzhu/configor"
	"gitlab.wallstcn.com/wscnbackend/ivankastd"
	"os"
	"path/filepath"
)

var (
	RootDir string
)

var Conf struct {
	DevMode      bool
	LogLevel     log.Level
	RootKey      string
	TemplatePath string
	CdnDomain    string
	Pprof        struct {
		Enable bool
		Port   string
	}
	Mysql []struct {
		Host        string
		Port        string
		User        string
		Password    string
		DbName      string
		Charset     string
		MaxConn     int
		MaxLifetime int
	}
	Redis   []RedisConf
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
		Author   int
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
		UseHttps       bool
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
	Leancloud struct {
		AppId              string
		AppKey             string
		AndroidAction      string
		ExpirationInterval int
	}
	Graylog struct {
		Ip       string
		Port     string
		Facility string
	}
	Prometheus struct {
		Enable bool
		Port   string
	}
	IvankaLog ivankastd.ConfigLog
}

type RedisConf struct {
	Tag         string
	Host        string
	Port        string
	Password    string
	MaxIdle     int
	IdleTimeout int
}

func initConfig() {
	filePath, rd := searchConfigFile()
	RootDir = rd
	err := configor.Load(&Conf, filePath)
	if err != nil {
		log.WithField("error", err).Panicln("Failed to load config file")
	}
	if len(Conf.RootKey) != 16 {
		log.Println("bad root key size")
	}
}

func searchConfigFile() (fullPath string, rootDir string) {
	rootDir = "./"
	// Search up for config file
	left, upALevel, right := "./", "../", "conf/appgo.yml"
	for i := 0; i < 5; i++ {
		p := filepath.Join(left, right)
		if _, err := os.Stat(p); err == nil {
			fullPath = p
			rootDir = left
			break
		}
		left = filepath.Join(left, upALevel)
	}
	return
}
