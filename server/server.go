package server

import (
	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/meatballhat/negroni-logrus"
	"github.com/oxfeeefeee/appgo"
	"github.com/oxfeeefeee/appgo/auth"
	"github.com/phyber/negroni-gzip/gzip"
	"github.com/rs/cors"
	"github.com/unrolled/render"
	"html/template"
	"net/http"
	_ "net/http/pprof"
	"strings"
)

type Server struct {
	ts          TokenStore
	middlewares []negroni.HandlerFunc
	ver         *versioning
	*mux.Router
}
type TokenStore interface {
	Validate(token auth.Token) bool
}

func NewServer(ts TokenStore, middlewares []negroni.HandlerFunc) *Server {
	return &Server{
		ts,
		middlewares,
		newVersioning(),
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

func (s *Server) AddHtml(path, layout string, htmls []interface{}, funcs template.FuncMap) {
	// add "static" template function
	static := func(path string) string {
		return s.ver.getStatic(path)
	}
	if funcs == nil {
		funcs = template.FuncMap{}
	}
	funcs["static"] = static

	renderer := render.New(render.Options{
		Directory:     appgo.Conf.TemplatePath,
		Layout:        layout,
		Funcs:         []template.FuncMap{funcs},
		IsDevelopment: appgo.Conf.DevMode,
	})
	for _, api := range htmls {
		h := newHandler(api, HandlerTypeHtml, s.ts, renderer)
		s.Handle(path+h.path, h).Methods("GET")
	}
}

func (s *Server) AddFeed(path string, feeds []interface{}) {
	// renderer is only for rendering error
	renderer := render.New(render.Options{
		Directory:     "N/A",
		IsDevelopment: appgo.Conf.DevMode,
	})
	for _, api := range feeds {
		h := newHandler(api, HandlerTypeFeed, s.ts, renderer)
		s.Handle(path+h.path, h).Methods("GET")
	}
}

func (s *Server) AddProxy(path string, handler http.Handler) {
	s.PathPrefix(path).Handler(http.StripPrefix(path, handler))
}

func (s *Server) AddStatic(path, fileDir string) {
	s.ver.addMap(path, fileDir)
	s.AddProxy(path, http.FileServer(http.Dir(fileDir)))
}

func (s *Server) AddAppleAppSiteAsso(content []byte) {
	f := func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(content)
	}
	s.HandleFunc("/apple-app-site-association", f)
}

func (s *Server) Serve() {
	if appgo.Conf.Pprof.Enable {
		go func() {
			log.Infoln(http.ListenAndServe(":"+appgo.Conf.Pprof.Port, nil))
		}()
	}

	n := negroni.New()
	rec := negroni.NewRecovery()
	rec.StackAll = true
	n.Use(rec)
	llog := negronilogrus.NewCustomMiddleware(
		appgo.Conf.LogLevel, &log.TextFormatter{}, "appgo")
	llog.Logger = log.StandardLogger()
	n.Use(llog)
	n.Use(cors.New(corsOptions()))
	for _, mw := range s.middlewares {
		n.Use(negroni.HandlerFunc(mw))
	}
	if appgo.Conf.Negroni.GZip {
		n.Use(gzip.Gzip(gzip.BestSpeed))
	}
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
