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

func (l *Lists) keystr(k interface{}) string {
	return fmt.Sprintf("%s:%v", l.namespace, k)
}

func (l *Lists) Prepend(key interface{}, item interface{}) (int64, error) {
	if newSize, err := Do("LPUSH", l.keystr(key), item); err != nil {
		return 0, err
	} else {
		return newSize.(int64), nil
	}
}

func (l *Lists) PopLastItem(key interface{}) (interface{}, error) {
	if item, err := Do("RPOP", l.keystr(key)); err != nil {
		return nil, err
	} else {
		return item, nil
	}
}
