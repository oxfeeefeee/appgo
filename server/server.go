package server

import (
	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/meatballhat/negroni-logrus"
	"github.com/oxfeeefeee/appgo"
	"github.com/oxfeeefeee/appgo/auth"
	"github.com/rs/cors"
	"github.com/unrolled/render"
	"html/template"
	"net/http"
	"strings"
)

type Server struct {
	ts TokenStore
	*mux.Router
}
type TokenStore interface {
	Validate(token auth.Token) bool
}

func NewServer(ts TokenStore) *Server {
	return &Server{
		ts,
		mux.NewRouter(),
	}
}

func (s *Server) AddRest(path string, rests []interface{}) {
	renderer := render.New(render.Options{
		Directory:     "N/A",
		IndentJSON:    appgo.Conf.DevMode,
		IsDevelopment: appgo.Conf.DevMode,
	})
	for _, api := range rests {
		h := newHandler(api, HandlerTypeJson, s.ts, renderer)
		s.Handle(path+h.path, h).Methods(h.supports...)
	}
}

func (s *Server) AddHtml(path string, htmls []interface{}, funcs template.FuncMap) {
	renderer := render.New(render.Options{
		Directory:     appgo.Conf.TemplatePath,
		Funcs:         []template.FuncMap{funcs},
		IsDevelopment: appgo.Conf.DevMode,
	})
	for _, api := range htmls {
		h := newHandler(api, HandlerTypeHtml, s.ts, renderer)
		s.Handle(path+h.path, h).Methods("GET")
	}
}

func (s *Server) AddProxy(path string, handler http.Handler) {
	s.PathPrefix(path).Handler(http.StripPrefix(path, handler))
}

func (s *Server) AddStatic(path, fileDir string) {
	s.AddProxy(path, http.FileServer(http.Dir(fileDir)))
}

func (s *Server) Serve() {
	n := negroni.New()
	n.Use(negroni.NewRecovery())
	n.Use(negronilogrus.NewCustomMiddleware(
		appgo.Conf.LogLevel, &log.TextFormatter{}, "appgo"))
	n.Use(cors.New(corsOptions()))
	n.UseHandler(s)
	n.Run(appgo.Conf.Negroni.Port)
}

func corsOptions() cors.Options {
	origins := strings.Split(appgo.Conf.Cors.AllowedOrigins, ",")
	methods := strings.Split(appgo.Conf.Cors.AllowedMethods, ",")
	headers := strings.Split(appgo.Conf.Cors.AllowedHeaders, ",")
	return cors.Options{
		AllowedOrigins:     origins,
		AllowedMethods:     methods,
		AllowedHeaders:     headers,
		OptionsPassthrough: appgo.Conf.Cors.OptionsPassthrough,
		Debug:              appgo.Conf.Cors.Debug,
	}
}
