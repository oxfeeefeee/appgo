package server

import (
	//log "github.com/Sirupsen/logrus"
	"github.com/oxfeeefeee/appgo"
	"github.com/oxfeeefeee/appgo/auth"
	"github.com/oxfeeefeee/appgo/redis"
	"net/http"
)

const (
	zsetNamespace = "metrics:access"
	totalKey      = "total"

	contextKeyUser = "user"
)

type Metrics struct {
	schema MetricsSchema
	zsets  *redis.Zsets
}

func newMetrics(schema MetricsSchema) *Metrics {
	zsets := redis.NewZsets(zsetNamespace)
	return &Metrics{
		schema, zsets,
	}
}

func (m *Metrics) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	next(rw, r)

	keys := m.schema.KeysGen(r)
	id := getUserFromToken(r)
	for _, key := range keys {
		m.zsets.Incrby(id.String(), 1, key)
		m.zsets.Incrby(totalKey, 1, key)
	}
}

type DefaultSchema struct {
}

func NewDefaultSchema() *DefaultSchema {
	return &DefaultSchema{}
}

func (s *DefaultSchema) KeysGen(r *http.Request) (keys []string) {
	return []string{r.Method + r.URL.Path}
}

func getUserFromToken(r *http.Request) appgo.Id {
	token := auth.Token(r.Header.Get(appgo.CustomTokenHeaderName))
	user, _ := token.Validate()
	return user
}
