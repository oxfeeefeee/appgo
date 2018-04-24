package server

import (
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	gkmetrics "github.com/go-kit/kit/metrics"
	gkprometheus "github.com/go-kit/kit/metrics/prometheus"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"github.com/oxfeeefeee/appgo"
	"github.com/oxfeeefeee/appgo/auth"
	"github.com/oxfeeefeee/appgo/toolkit/strutil"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"github.com/unrolled/render"
	"gitlab.wallstcn.com/wscnbackend/ivankastd"
	"net"
	"net/http"
	"reflect"
	"strings"
	"time"
)

const (
	UserIdFieldName         = "UserId__"
	AdminUserIdFieldName    = "AdminUserId__"
	AuthorIdFieldName       = "AuthorId__"
	ResIdFieldName          = "ResourceId__"
	ContentFieldName        = "Content__"
	RequestFieldName        = "Request__"
	ConfVerFieldName        = "ConfVer__"
	AppVerFieldName         = "AppVer__"
	PlatformFieldName       = "Platform__"
	AdminUserAuthsFieldName = "AdminUserAuths__"
	maxVersion              = 99
)

const (
	_ HandlerType = iota
	HandlerTypeJson
	HandlerTypeHtml
)

var decoder = schema.NewDecoder()

var metrics_req_dur gkmetrics.Histogram

var metrics_query_count map[string]gkmetrics.Counter

type HandlerType int

type httpFunc struct {
	requireAuth    bool
	requireAdmin   bool
	requireAuthor  bool
	hasResId       bool
	hasContent     bool
	hasRequest     bool
	hasConfVer     bool
	hasAppVer      bool
	hasPlatform    bool
	dummyInput     bool
	allowAnonymous bool
	inputType      reflect.Type
	contentType    reflect.Type
	funcValue      reflect.Value
}

type AdminAuthHandler func(r *http.Request, roleGruop string) (appgo.Id, appgo.Role, []string, error)

type handler struct {
	htype            HandlerType
	path             string
	template         string
	funcs            map[string]*httpFunc
	methodAuth       map[string]string
	supports         []string
	ts               TokenStore
	renderer         *render.Render
	adminAuthHandler AdminAuthHandler
}

func init() {
	decoder.IgnoreUnknownKeys(true)

	if appgo.Conf.Prometheus.Enable {
		metrics_req_dur = gkprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
			Namespace: "appgo",
			Subsystem: "http",
			Name:      "request_duration_microseconds",
			Help:      "Total time spent serving requests.",
		}, []string{})
		metrics_query_count = map[string]gkmetrics.Counter{
			"all": gkprometheus.NewCounterFrom(stdprometheus.CounterOpts{
				Namespace: "appgo",
				Subsystem: "http",
				Name:      "request_counter",
				Help:      "Total served requests count.",
			}, []string{})}
	}
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer addMetrics(r, time.Now())
	startTime := time.Now()
	var userId int64 // for use with log

	method := r.Method
	ver := apiVersionFromHeader(r)
	if ver > 1 && ver <= maxVersion {
		method += strutil.FromInt(ver)
	}
	f, ok := h.funcs[method]
	if !ok {
		h.renderError(w, appgo.NewApiErr(
			appgo.ECodeNotFound,
			"Bad API version"))
		return
	}
	var input reflect.Value
	if f.dummyInput {
		input = reflect.ValueOf((*appgo.DummyInput)(nil))
	} else {
		input = reflect.New(f.inputType)
		if err := decoder.Decode(input.Interface(), r.URL.Query()); err != nil {
			h.renderError(w, appgo.NewApiErr(appgo.ECodeBadRequest, err.Error()))
			return
		}
	}
	if f.requireAuth {
		user, _ := h.authByHeader(r)
		s := input.Elem()
		field := s.FieldByName(UserIdFieldName)
		if user == 0 {
			if f.allowAnonymous {
				field.SetInt(appgo.AnonymousId)
			} else {
				h.renderError(w, appgo.NewApiErr(
					appgo.ECodeUnauthorized,
					"either remove UserId__ in your input define, or add allowAnonymous tag",
				))
				return
			}
		} else {
			field.SetInt(int64(user))
			userId = int64(user)
			go auth.RecordLastActiveAt(appgo.Id(int64(user)))
		}
	} else if f.requireAdmin {
		token := r.Header.Get(appgo.CustomWallStTokenHeaderName)
		var (
			user  appgo.Id
			role  appgo.Role
			auths []string
			err   error
		)
		if token != "" && h.adminAuthHandler != nil {
			roleGroup := h.methodAuth[method]
			user, role, auths, err = h.adminAuthHandler(r, roleGroup)
			if err != nil {
				h.renderError(w, appgo.ApiErrFromGoErr(err))
				return
			} else {
				if len(auths) <= 0 {
					h.renderError(w, appgo.NewApiErr(appgo.ECodeUnauthorized, "not authorized"))
					return
				}
			}
		} else {
			user, role = h.authByHeader(r)
			auths = []string{"xgb_admin"}
		}
		s := input.Elem()

		m := make(map[string]bool)
		userAuthsValue := reflect.ValueOf(m)
		for _, auth := range auths {
			userAuthsValue.SetMapIndex(reflect.ValueOf(auth), reflect.ValueOf(true))
		}
		fieldName := s.FieldByName(AdminUserAuthsFieldName)
		if fieldName.IsValid() {
			fieldName.Set(userAuthsValue)
		}
		f := s.FieldByName(AdminUserIdFieldName)
		if user == 0 || role != appgo.RoleWebAdmin {
			h.renderError(w, appgo.NewApiErr(
				appgo.ECodeUnauthorized,
				"admin role required, you could remove AdminUserId__ in your input define"))
			return
		}
		f.SetInt(int64(user))
		userId = int64(user)
	}
	if f.requireAuthor {
		authorId, expired := h.authorIdFromHeader(r)
		if authorId == 0 {
			h.renderError(w, appgo.NewApiErr(
				appgo.ECodeUnauthorized,
				"author role required, check if your header has correct author token"))
			return
		} else if expired {
			h.renderError(w, appgo.NewApiErr(
				appgo.ECodeUnauthorized,
				"author token expired"))
			return
		}
		s := input.Elem()
		f := s.FieldByName(AuthorIdFieldName)
		f.SetInt(int64(authorId))
	}
	if f.hasResId {
		vars := mux.Vars(r)
		id := appgo.IdFromStr(vars["id"])
		if id == 0 {
			h.renderError(w, appgo.NewApiErr(
				appgo.ECodeNotFound,
				"ResourceId ('{id}' in url) required, you could remove ResourceId__ in your input define"))
			return
		}
		s := input.Elem()
		f := s.FieldByName(ResIdFieldName)
		f.SetInt(int64(id))
	}
	if f.hasContent {
		content := reflect.New(f.contentType.Elem())
		if err := json.NewDecoder(r.Body).Decode(content.Interface()); err != nil {
			h.renderError(w, appgo.NewApiErr(appgo.ECodeBadRequest, err.Error()))
			return
		}
		s := input.Elem()
		f := s.FieldByName(ContentFieldName)
		f.Set(content)
	}
	if f.hasRequest {
		s := input.Elem()
		f := s.FieldByName(RequestFieldName)
		f.Set(reflect.ValueOf(r))
	}
	if f.hasConfVer {
		ver := confVersionFromHeader(r)
		s := input.Elem()
		f := s.FieldByName(ConfVerFieldName)
		f.Set(reflect.ValueOf(ver))
	}
	if f.hasAppVer {
		appVer := appVersionFromHeader(r)
		s := input.Elem()
		field := s.FieldByName(AppVerFieldName)
		field.SetString(appVer)
	}
	if f.hasPlatform {
		platform := platformFromHeader(r)
		s := input.Elem()
		field := s.FieldByName(PlatformFieldName)
		field.SetString(platform)
	}

	argsIn := []reflect.Value{input}
	returns := f.funcValue.Call(argsIn)
	rl := len(returns)
	if !(rl == 1 || rl == 2 || (rl == 3 && h.htype == HandlerTypeHtml)) {
		h.renderError(w, appgo.NewApiErr(appgo.ECodeInternal, "Bad api-func format"))
		return
	}
	// returns (reply, template-name, error) or (reply, error) or returns (error)
	retErr := returns[rl-1]
	// First check if err is nil
	if retErr.IsNil() {
		if rl == 3 {
			template := returns[1].Interface().(string)
			h.renderHtml(w, template, returns[0].Interface())
		} else if rl == 2 {
			logUserActivity(r, startTime, userId, int(appgo.ECodeOK), -1)
			h.renderData(w, returns[0].Interface())
		} else { // Empty return
			logUserActivity(r, startTime, userId, int(appgo.ECodeOK), -1)
			h.renderData(w, map[string]string{})
		}
	} else {
		if aerr, ok := retErr.Interface().(*appgo.ApiError); !ok {
			aerr = appgo.NewApiErr(appgo.ECodeInternal, "Bad api-func format")
		} else {
			if h.htype == HandlerTypeHtml && aerr.Code == appgo.ECodeRedirect {
				http.Redirect(w, r, aerr.Msg, http.StatusFound)
				return
			}
			logUserActivity(r, startTime, userId, appgo.ECodeInternal, -1)
			h.renderError(w, aerr)
		}
	}
}

func addMetrics(r *http.Request, begin time.Time) {
	if !appgo.Conf.Prometheus.Enable {
		return
	}
	metrics_req_dur.Observe(float64(time.Since(begin) / time.Microsecond))

	path := r.RequestURI
	if i := strings.IndexByte(path, '?'); i > 0 {
		path = path[:i]
	}
	path = strings.Replace(path, "/", "_", -1)
	key := r.Method + path
	if _, ok := metrics_query_count[key]; !ok {
		metrics_query_count[key] = gkprometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Namespace: "appgo",
			Subsystem: "http",
			Name:      "request_counter_" + key,
			Help:      fmt.Sprintf("Total served %s requests count.", key),
		}, []string{})
	}
	metrics_query_count["all"].Add(1)
	metrics_query_count[key].Add(1)
}

func (h *handler) authByHeader(r *http.Request) (appgo.Id, appgo.Role) {
	token := auth.Token(r.Header.Get(appgo.CustomTokenHeaderName))
	user, role := token.Validate()
	if user == 0 {
		return 0, 0
	}
	platform := platformFromHeader(r)
	if !h.ts.Validate(user, role, token, platform) {
		return 0, 0
	}
	return user, role
}

func (h *handler) authorIdFromHeader(r *http.Request) (appgo.Id, bool) {
	token := auth.Token(r.Header.Get(appgo.CustomAuthorTokenHeaderName))
	if authorId, _, expireAt, err := token.Parse(); err != nil {
		return 0, false
	} else {
		return authorId, expireAt.Before(time.Now())
	}
}

func apiVersionFromHeader(r *http.Request) int {
	v := r.Header.Get(appgo.CustomVersionHeaderName)
	return strutil.ToInt(v)
}

func confVersionFromHeader(r *http.Request) int64 {
	v := r.Header.Get(appgo.CustomConfVerHeaderName)
	return strutil.ToInt64(v)
}

func appVersionFromHeader(r *http.Request) string {
	return r.Header.Get(appgo.CustomAppVerHeaderName)
}

func platformFromHeader(r *http.Request) string {
	return r.Header.Get(appgo.CustomPlatformHeaderName)
}

func newHandler(funcSet interface{}, htype HandlerType,
	ts TokenStore, renderer *render.Render, h AdminAuthHandler) *handler {
	funcs := make(map[string]*httpFunc)
	// Let if panic if funSet's type is not right
	path := ""
	template := ""
	t := reflect.TypeOf(funcSet).Elem()
	if field, ok := t.FieldByName("META"); !ok {
		log.Panicln("Bad META setting (path, template)")
	} else {
		if p := field.Tag.Get("path"); p == "" {
			log.Panicln("Empty API path")
		} else {
			path = p
		}
		if htype == HandlerTypeHtml {
			t := field.Tag.Get("template")
			template = t
		}
	}
	methods := []string{"GET", "POST", "PUT", "DELETE"}

	methodsAuth := make(map[string]string)
	if field, ok := t.FieldByName("AUTH"); ok {
		for _, m := range methods {
			if auth := field.Tag.Get(strings.ToLower(m)); auth != "" {
				methodsAuth[m] = auth
			}
		}
	}

	structVal := reflect.Indirect(reflect.ValueOf(funcSet))
	supports := make([]string, 0, 4)
	if htype == HandlerTypeJson {
		for _, m := range methods {
			for i := 1; i <= maxVersion; i++ { //versions
				if i > 1 {
					m += strutil.FromInt(i)
				}
				if fun, err := newHttpFunc(structVal, m); err != nil {
					log.Panicln(err)
				} else if fun != nil {
					funcs[m] = fun
					supports = append(supports, m)
				}
			}
		}
		if len(supports) == 0 {
			log.Panicln("API supports no HTTP method")
		}
	} else if htype == HandlerTypeHtml {
		if fun, err := newHttpFunc(structVal, "HTML"); err != nil {
			log.Panicln(err)
		} else if fun == nil {
			log.Panicln("No HTML function for html")
		} else {
			funcs["GET"] = fun
		}
	} else {
		log.Panicln("Bad handler type")
	}
	return &handler{htype, path, template, funcs, methodsAuth, supports, ts, renderer, h}
}

func newHttpFunc(structVal reflect.Value, fieldName string) (*httpFunc, error) {
	fieldVal := structVal.MethodByName(fieldName)
	if !fieldVal.IsValid() {
		return nil, nil
	}
	ftype := fieldVal.Type()
	inNum := ftype.NumIn()
	if inNum != 1 {
		return nil, errors.New("API func needs to have exact 1 parameter")
	}
	inputType := ftype.In(0)
	dummyInput := false
	if inputType.Kind() != reflect.Ptr {
		return nil, errors.New("API func's parameter needs to be a pointer")
	}
	if inputType == reflect.TypeOf((*appgo.DummyInput)(nil)) {
		dummyInput = true
	}
	inputType = inputType.Elem()
	requireAuth := false
	allowAnonymous := false
	if fromIdField, ok := inputType.FieldByName(UserIdFieldName); ok {
		requireAuth = true
		if fromIdField.Type.Kind() != reflect.Int64 {
			return nil, errors.New("API func's 2nd parameter needs to be Int64")
		}
		aa := fromIdField.Tag.Get("allowAnonymous")
		allowAnonymous = (aa == "true")
	}
	requireAdmin := false
	if fromIdType, ok := inputType.FieldByName(AdminUserIdFieldName); ok {
		requireAdmin = true
		if fromIdType.Type.Kind() != reflect.Int64 {
			return nil, errors.New("API func's 2nd parameter needs to be Int64")
		}
	}

	if requireAdmin {
		if fromIdType, ok := inputType.FieldByName(AdminUserAuthsFieldName); ok {
			if fromIdType.Type.Kind() != reflect.Map {
				return nil, errors.New("API func's AdminUserAuths__ parameter needs to be map[string]bool")
			}
		}
	}

	requireAuthor := false
	if fromIdField, ok := inputType.FieldByName(AuthorIdFieldName); ok {
		requireAuthor = true
		if fromIdField.Type.Kind() != reflect.Int64 {
			return nil, errors.New("AuthorId needs to be of type Int64")
		}
	}
	hasResId := false
	if resIdType, ok := inputType.FieldByName(ResIdFieldName); ok {
		hasResId = true
		if resIdType.Type.Kind() != reflect.Int64 {
			return nil, errors.New("ResId needs to be Int64")
		}
	}
	hasContent := false
	var contentType reflect.Type
	if ctype, ok := inputType.FieldByName(ContentFieldName); ok {
		hasContent = true
		contentType = ctype.Type
		if ctype.Type.Kind() != reflect.Ptr {
			return nil, errors.New("Content needs to be a pointer")
		}
	}
	hasRequest := false
	if ctype, ok := inputType.FieldByName(RequestFieldName); ok {
		hasRequest = true
		if ctype.Type.Kind() != reflect.Ptr {
			return nil, errors.New("Request needs to be a pointer to http.Request")
		}
		if ctype.Type.Elem() != reflect.TypeOf((*http.Request)(nil)).Elem() {
			return nil, errors.New("Request needs to be a pointer to http.Request")
		}
	}
	hasConfVer := false
	if confVerType, ok := inputType.FieldByName(ConfVerFieldName); ok {
		hasConfVer = true
		if confVerType.Type.Kind() != reflect.Int64 {
			return nil, errors.New("ConfVer needs to be Int64")
		}
	}
	hasAppVer := false
	if appVerType, ok := inputType.FieldByName(AppVerFieldName); ok {
		hasAppVer = true
		if appVerType.Type.Kind() != reflect.String {
			return nil, errors.New("AppVer needs to be string")
		}
	}
	hasPlatform := false
	if platformType, ok := inputType.FieldByName(PlatformFieldName); ok {
		hasPlatform = true
		if platformType.Type.Kind() != reflect.String {
			return nil, errors.New("Platform needs to be string")
		}
	}
	return &httpFunc{requireAuth, requireAdmin, requireAuthor,
		hasResId, hasContent, hasRequest, hasConfVer, hasAppVer, hasPlatform,
		dummyInput, allowAnonymous, inputType, contentType, fieldVal}, nil
}

func logUserActivity(r *http.Request, startTime time.Time, userId int64, resCode int, bytesOut int) {
	var remoteIp string
	remoteIp = r.Header.Get("x-client-ip")
	if strings.TrimSpace(remoteIp) == "" {
		remoteIp = r.Header.Get("X-Forwarded-For")
		if strings.TrimSpace(remoteIp) == "" {
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err == nil {
				remoteIp = ip
			}
		}
	}

	deviceId := r.Header.Get("X-Device-Id")
	deviceId = strings.TrimSpace(deviceId)
	var deviceType string
	if len(deviceId) > 30 {
		if strings.HasPrefix(deviceId, "android") {
			deviceType = "android"
		} else {
			deviceType = "ios"
		}
	} else {
		deviceType = "web"
	}

	ivankastd.LogUserActivity(ivankastd.LogFields{
		"type":          "webaccess",
		"remote_ip":     remoteIp,
		"host":          r.Host,
		"uri":           r.RequestURI,
		"method":        r.Method,
		"path":          r.URL.Path,
		"route":         "undefined",
		"referer":       r.Referer(),
		"user_agent":    r.UserAgent(),
		"status":        200,
		"latency":       (time.Now().UnixNano() - startTime.UnixNano()) / 1000000,
		"bytes_in":      -1,
		"device_id":     deviceId,
		"device_type":   deviceType,
		"bytes_out":     bytesOut,
		"user_id":       userId,
		"response_code": resCode,
	}, "webaccess")
}
