package appgo

import (
	log "github.com/Sirupsen/logrus"
	"github.com/jinzhu/configor"
)

var Conf struct {
	DevMode  bool
	LogLevel log.Level
	Negroni  struct {
		Port string
	}
}

func init() {
	err := configor.Load(&Conf, "./conf/appgo.yml")
	if err != nil {
		log.WithField("error", err).Panicln("Failed to load config file")
	}
}
