package server

import (
	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/meatballhat/negroni-logrus"
	"github.com/oxfeeefeee/appgo"
	"github.com/oxfeeefeee/appgo/auth"
	"github.com/phyber/negroni-gzip/gzip"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/cors"
	"github.com/unrolled/render"
	"html/template"
	"net/http"
	_ "net/http/pprof"
	"strings"
)

type Server struct {
	ts          TokenStore
	middlewares []negroni.Handler
	ver         *versioning
	*mux.Router
}

type TokenStore interface {
	Validate(uid appgo.Id, role appgo.Role, token auth.Token, platform string) bool
}

type MetricsSchema interface {
	KeysGen(r *http.Request) map[string]string
}

func NewServer(ts TokenStore, middlewares []negroni.Handler,
	mschema []MetricsSchema) *Server {
	for _, s := range mschema {
		m := newMetrics(s)
		middlewares = append(middlewares, m)
	}
	return &Server{
		ts,
		middlewares,
		newVersioning(),
		mux.NewRouter(),
	}
}

func (s *Server) AddRest(path string, rests []interface{}, tokenParser AdminAuthHandler) {
	renderer := render.New(render.Options{
		Directory:     "N/A",
		IndentJSON:    appgo.Conf.DevMode,
		IsDevelopment: appgo.Conf.DevMode,
	})
	for _, api := range rests {
		h := newHandler(api, HandlerTypeJson, s.ts, renderer, tokenParser)
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
		h := newHandler(api, HandlerTypeHtml, s.ts, renderer, nil)
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
	if appgo.Conf.Prometheus.Enable {
		go func() {
			mux := http.NewServeMux()
			mux.Handle("/metrics", prometheus.Handler())
			log.Infoln(http.ListenAndServe(":"+appgo.Conf.Prometheus.Port, mux))
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
		n.Use(mw)
	}
	if appgo.Conf.Negroni.GZip {
		n.Use(gzip.Gzip(gzip.BestSpeed))
	}
	n.UseHandler(s)
	n.Run(appgo.Conf.Negroni.Port)
}

func GetUserFromToken(r *http.Request) appgo.Id {
	token := auth.Token(r.Header.Get(appgo.CustomTokenHeaderName))
	user, _, _, _ := token.Parse()
	return user
}

func corsOptions() cors.Options {
	origins := strings.Split(appgo.Conf.Cors.AllowedOrigins, ",")
	// trim spaces of origin strings
	for idx, _ := range origins {
		origins[idx] = strings.TrimSpace(origins[idx])
	}
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
