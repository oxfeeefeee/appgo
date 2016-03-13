package server

import (
	log "github.com/Sirupsen/logrus"
	"github.com/oxfeeefeee/appgo"
	"github.com/unrolled/render"
	"net/http"
)

var renderer *render.Render

func init() {
	renderer = render.New(render.Options{
		IndentJSON: appgo.Conf.DevMode,
	})
}

func renderJson(w http.ResponseWriter, v interface{}) {
	err := renderer.JSON(w, http.StatusOK, v)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"data":  v,
		}).Error("Error rendering json")
	}
}

func renderError(w http.ResponseWriter, err *appgo.ApiError) {
	renderJson(w, err)
}
