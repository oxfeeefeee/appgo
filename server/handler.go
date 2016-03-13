package server

import (
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"github.com/oxfeeefeee/appgo"
	"net/http"
	"reflect"
)

const (
	UserIdFieldName      = "UserId__"
	AdminUserIdFieldName = "AdminUserId__"
	ResIdFieldName       = "ResourceId__"
	ContentFieldName     = "Content__"
)

var decoder = schema.NewDecoder()

type httpFunc struct {
	requireAuth    bool
	requireAdmin   bool
	hasResId       bool
	hasContent     bool
	dummyInput     bool
	allowAnonymous bool
	inputType      reflect.Type
	contentType    reflect.Type
	funcValue      reflect.Value
}

type handler struct {
	path     string
	funcs    map[string]*httpFunc
	supports []string
	auth     Authenticator
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f := h.funcs[r.Method]
	var input reflect.Value
	if f.dummyInput {
		input = reflect.ValueOf((*appgo.DummyInput)(nil))
	} else {
		input = reflect.New(f.inputType)
		if err := decoder.Decode(input.Interface(), r.URL.Query()); err != nil {
			renderError(w, appgo.NewApiErr(appgo.ECodeBadRequest, err.Error()))
			return
		}
	}
	token := r.Header.Get(appgo.CustomTokenHeaderName)
	user, role := h.auth.Validate(token)
	if f.requireAuth {
		s := input.Elem()
		field := s.FieldByName(UserIdFieldName)
		if user == 0 {
			if f.allowAnonymous {
				field.SetInt(appgo.AnonymousId)
			} else {
				renderError(w, appgo.NewApiErrWithCode(appgo.ECodeUnauthorized))
				return
			}
		}
		field.SetInt(int64(user))
	} else if f.requireAdmin {
		s := input.Elem()
		f := s.FieldByName(AdminUserIdFieldName)
		if user == 0 || role != appgo.RoleWebAdmin {
			renderError(w, appgo.NewApiErrWithCode(appgo.ECodeUnauthorized))
			return
		}
		f.SetInt(int64(user))
	}
	if f.hasResId {
		vars := mux.Vars(r)
		id := appgo.IdFromStr(vars["id"])
		if id == 0 {
			renderError(w, appgo.NewApiErrWithCode(appgo.ECodeNotFound))
			return
		}
		s := input.Elem()
		f := s.FieldByName(ResIdFieldName)
		f.SetInt(int64(id))
	}
	if f.hasContent {
		content := reflect.New(f.contentType.Elem())
		if err := json.NewDecoder(r.Body).Decode(content.Interface()); err != nil {
			renderError(w, appgo.NewApiErr(appgo.ECodeBadRequest, err.Error()))
			return
		}
		s := input.Elem()
		f := s.FieldByName(ContentFieldName)
		f.Set(content)
	}
	argsIn := []reflect.Value{input}
	returns := f.funcValue.Call(argsIn)
	if len(returns) == 0 || len(returns) > 2 {
		renderError(w, appgo.NewApiErr(appgo.ECodeInternal, "Bad api-func format"))
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
			renderJSON(w, returns[0].Interface())
		} else { // Empty return
			renderJSON(w, map[string]string{})
		}
	} else {
		if aerr, ok := retErr.Interface().(*appgo.ApiError); !ok {
			aerr = appgo.NewApiErr(appgo.ECodeInternal, "Bad api-func format")
		} else {
			renderError(w, aerr)
		}
	}
}

func newHandler(funcSet interface{}, auth Authenticator) *handler {
	funcs := make(map[string]*httpFunc)
	// Let if panic if funSet's type is not right
	path := ""
	t := reflect.TypeOf(funcSet).Elem()
	if field, ok := t.FieldByName("META"); !ok {
		log.Panicln("Bad API path")
	} else if p := field.Tag.Get("path"); p == "" {
		log.Panicln("Empty API path")
	} else {
		path = p
	}

	structVal := reflect.Indirect(reflect.ValueOf(funcSet))
	methods := []string{"GET", "POST", "PUT", "DELETE"}
	supports := make([]string, 0, 4)
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
	return &handler{path, funcs, supports, auth}
}

func newHttpFunc(structVal reflect.Value, fieldName string) (*httpFunc, error) {
	fieldVal := structVal.MethodByName(fieldName)
	if !fieldVal.IsValid() {
		return nil, nil
	}
	ftype := fieldVal.Type()
	inNum := ftype.NumIn()
	if inNum != 1 {
		return nil, fmt.Errorf("API func needs to have exact 1 parameter")
	}
	inputType := ftype.In(0)
	dummyInput := false
	if inputType.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("API func's parameter needs to be a pointer")
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
			return nil, fmt.Errorf("API func's 2nd parameter needs to be Int64")
		}
		aa := fromIdField.Tag.Get("allowAnonymous")
		allowAnonymous = (aa == "true")
	}
	requireAdmin := false
	if fromIdType, ok := inputType.FieldByName(AdminUserIdFieldName); ok {
		requireAdmin = true
		if fromIdType.Type.Kind() != reflect.Int64 {
			return nil, fmt.Errorf("API func's 2nd parameter needs to be Int64")
		}
	}
	hasResId := false
	if resIdType, ok := inputType.FieldByName(ResIdFieldName); ok {
		hasResId = true
		if resIdType.Type.Kind() != reflect.Int64 {
			return nil, fmt.Errorf("ResId needs to be Int64")
		}
	}
	hasContent := false
	var contentType reflect.Type
	if ctype, ok := inputType.FieldByName(ContentFieldName); ok {
		hasContent = true
		contentType = ctype.Type
		if ctype.Type.Kind() != reflect.Ptr {
			return nil, fmt.Errorf("Content needs to be Int64")
		}
	}
	return &httpFunc{requireAuth, requireAdmin, hasResId, hasContent,
		dummyInput, allowAnonymous, inputType, contentType, fieldVal}, nil
}
