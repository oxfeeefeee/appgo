package toolkit

import (
	"sync"
)

type Map struct {
	m map[string]interface{}
	sync.RWMutex
}

func NewMap() *Map {
	return &Map{
		m: make(map[string]interface{}),
	}
}

func (m *Map) Add(key string, val interface{}) {
	m.Lock()
	defer m.Unlock()
	m.m[key] = val
}

func (m *Map) Remove(key string) {
	m.Lock()
	defer m.Unlock()
	delete(m.m, key)
}

func (m *Map) Has(key string) bool {
	m.RLock()
	defer m.RUnlock()
	_, ok := m.m[key]
	return ok
}

func (m *Map) Get(key string) interface{} {
	m.RLock()
	defer m.RUnlock()
	return m.m[key]
}

func (m *Map) GetString(key string) string {
	v := m.Get(key)
	if v == nil {
		return ""
	}
	return v.(string)
}
