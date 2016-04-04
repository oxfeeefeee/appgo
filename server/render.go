package server

import (
	log "github.com/Sirupsen/logrus"
	"github.com/oxfeeefeee/appgo"
	"net/http"
)

func (h *handler) renderData(w http.ResponseWriter, v interface{}) {
	if h.htype == HandlerTypeJson {
		h.renderJSON(w, v)
	} else if h.htype == HandlerTypeHtml {
		h.renderHtml(w, h.template, v)
	} else {
		panic("Bad handler type")
	}
}

func (h *handler) renderError(w http.ResponseWriter, err *appgo.ApiError) {
	if h.htype == HandlerTypeJson {
		h.renderJSON(w, err)
	} else if h.htype == HandlerTypeHtml {
		h.renderHtml(w, "content", err)
	} else {
		panic("Bad handler type")
	}
}

func (h *handler) renderJSON(w http.ResponseWriter, v interface{}) {
	err := h.renderer.JSON(w, http.StatusOK, v)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"data":  v,
		}).Error("Error rendering json")
	}
}

func (h *handler) renderHtml(w http.ResponseWriter, template string, data interface{}) {
	err := h.renderer.HTML(w, http.StatusOK, template, data)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"data":  data,
		}).Error("Error rendering html")
	}
}
