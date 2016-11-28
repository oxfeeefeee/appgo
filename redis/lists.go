package redis

import (
	"fmt"
)

type Lists struct {
	namespace string
}

func NewLists(namespace string) *Lists {
	return &Lists{namespace}
}

func (l *Lists) keyStr(k interface{}) string {
	return fmt.Sprintf("%s:%v", l.namespace, k)
}

func (l *Lists) AllItems(k interface{}) ([]interface{}, error) {
	if items, err := Do("LRANGE", l.keyStr(k), 0, -1); err != nil {
		return nil, err
	} else {
		return items.([]interface{}), nil
	}
}

func (l *Lists) Prepend(key interface{}, item interface{}) (int64, error) {
	if newSize, err := Do("LPUSH", l.keyStr(key), item); err != nil {
		return 0, err
	} else {
		return newSize.(int64), nil
	}
}

func (l *Lists) PopLastItem(key interface{}) (interface{}, error) {
	if item, err := Do("RPOP", l.keyStr(key)); err != nil {
		return nil, err
	} else {
		return item, nil
	}
}
