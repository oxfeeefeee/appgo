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
		Directory:     appgo.Conf.TemplatePath,
		IndentJSON:    appgo.Conf.DevMode,
		IsDevelopment: appgo.Conf.DevMode,
	})
}

func renderJSON(w http.ResponseWriter, v interface{}) {
	err := renderer.JSON(w, http.StatusOK, v)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"data":  v,
		}).Error("Error rendering json")
	}
}

func renderJsonError(w http.ResponseWriter, err *appgo.ApiError) {
	renderJSON(w, err)
}

func renderHtml(w http.ResponseWriter, template string, data interface{}) {
	err := renderer.HTML(w, http.StatusOK, template, data)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"data":  data,
		}).Error("Error rendering html")
	}
}

func renderHtmlError(w http.ResponseWriter, err *appgo.ApiError) {
	renderHtml(w, "content", err)
}
