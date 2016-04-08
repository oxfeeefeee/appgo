package server

import (
	"encoding/json"
	"errors"
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"github.com/oxfeeefeee/appgo"
	"github.com/oxfeeefeee/appgo/auth"
	"github.com/unrolled/render"
	"net/http"
	"reflect"
)

const (
	UserIdFieldName      = "UserId__"
	AdminUserIdFieldName = "AdminUserId__"
	ResIdFieldName       = "ResourceId__"
	ContentFieldName     = "Content__"
	RequestFieldName     = "Request__"
)

const (
	_ HandlerType = iota
	HandlerTypeJson
	HandlerTypeHtml
)

var decoder = schema.NewDecoder()

type HandlerType int

type httpFunc struct {
	requireAuth    bool
	requireAdmin   bool
	hasResId       bool
	hasContent     bool
	hasRequest     bool
	dummyInput     bool
	allowAnonymous bool
	inputType      reflect.Type
	contentType    reflect.Type
	funcValue      reflect.Value
}

type handler struct {
	htype    HandlerType
	path     string
	template string
	funcs    map[string]*httpFunc
	supports []string
	ts       TokenStore
	renderer *render.Render
}

func init() {
	decoder.IgnoreUnknownKeys(true)
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f := h.funcs[r.Method]
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
		}
	} else if f.requireAdmin {
		user, role := h.authByHeader(r)
		s := input.Elem()
		f := s.FieldByName(AdminUserIdFieldName)
		if user == 0 || role != appgo.RoleWebAdmin {
			h.renderError(w, appgo.NewApiErr(
				appgo.ECodeUnauthorized,
				"admin role required, you could remove AdminUserId__ in your input define"))
			return
		}
		f.SetInt(int64(user))
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
	argsIn := []reflect.Value{input}
	returns := f.funcValue.Call(argsIn)
	if len(returns) == 0 || len(returns) > 2 {
		h.renderError(w, appgo.NewApiErr(appgo.ECodeInternal, "Bad api-func format"))
		return
	}
	// Either returns (reply, error) or returns (error)
	var retErr reflect.Value
	if len(returns) == 1 {
		retErr = returns[0]
	} else {
		retErr = returns[1]
	}
	// First check is err is nil
	if retErr.IsNil() {
		if len(returns) == 2 {
			h.renderData(w, returns[0].Interface())
		} else { // Empty return
			h.renderData(w, map[string]string{})
		}
	} else {
		if aerr, ok := retErr.Interface().(*appgo.ApiError); !ok {
			aerr = appgo.NewApiErr(appgo.ECodeInternal, "Bad api-func format")
		} else {
			h.renderError(w, aerr)
		}
	}
}

func (h *handler) authByHeader(r *http.Request) (appgo.Id, appgo.Role) {
	token := auth.Token(r.Header.Get(appgo.CustomTokenHeaderName))
	user, role := token.Validate()
	if user == 0 {
		return 0, 0
	}
	if !h.ts.Validate(token) {
		return 0, 0
	}
	return user, role
}

func newHandler(funcSet interface{}, htype HandlerType,
	ts TokenStore, renderer *render.Render) *handler {
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
			if t := field.Tag.Get("template"); t == "" {
				log.Panicln("Empty HTML template")
			} else {
				template = t
			}
		}
	}
	structVal := reflect.Indirect(reflect.ValueOf(funcSet))
	supports := make([]string, 0, 4)
	if htype == HandlerTypeJson {
		methods := []string{"GET", "POST", "PUT", "DELETE"}
		for _, m := range methods {
			if fun, err := newHttpFunc(structVal, m); err != nil {
				log.Panicln(err)
			} else if fun != nil {
				funcs[m] = fun
				supports = append(supports, m)
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
	return &handler{htype, path, template, funcs, supports, ts, renderer}
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
	return &httpFunc{requireAuth, requireAdmin, hasResId, hasContent, hasRequest,
		dummyInput, allowAnonymous, inputType, contentType, fieldVal}, nil
}
