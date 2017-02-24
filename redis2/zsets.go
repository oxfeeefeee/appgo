package redis2

import (
	"fmt"
	redigo "github.com/garyburd/redigo/redis"
	"github.com/oxfeeefeee/appgo"
	"strconv"
)

type Zsets struct {
	namespace string
}

type ZsetIncrbyParams struct {
	Key   interface{}
	Item  interface{}
	Score float64
}

func NewZsets(namespace string) *Zsets {
	return &Zsets{namespace}
}

func (z *Zsets) Add(key interface{}, score float64, item interface{}) error {
	if _, err := Do("ZADD", z.keystr(key), score, item); err != nil {
		return err
	}
	return nil
}

func (z *Zsets) Rem(key, item interface{}) error {
	if _, err := Do("ZREM", z.keystr(key), item); err != nil {
		return err
	}
	return nil
}

func (z *Zsets) Clear(key interface{}) error {
	if _, err := Do("DEL", z.keystr(key)); err != nil {
		return err
	}
	return nil
}

func (z *Zsets) RemByScore(key interface{}, min, max float64) error {
	if _, err := Do("ZREMRANGEBYSCORE", z.keystr(key), min, max); err != nil {
		return err
	}
	return nil
}

func (z *Zsets) Card(key interface{}) (int, error) {
	return redigo.Int(Do("ZCARD", z.keystr(key)))
}

func (z *Zsets) Count(key, min, max interface{}) (int, error) {
	return redigo.Int(Do("ZCOUNT", z.keystr(key), min, max))
}

func (z *Zsets) Incrby(key interface{}, score float64, item interface{}) error {
	if _, err := Do("ZINCRBY", z.keystr(key), score, item); err != nil {
		return err
	}
	return nil
}

func (z *Zsets) BatchIncrby(params []ZsetIncrbyParams) error {
	trans := BeginTrans()
	for _, p := range params {
		trans.Send("ZINCRBY", z.keystr(p.Key), p.Score, p.Item)
	}
	_, err := trans.Exec()
	return err
}

func (z *Zsets) Min(key interface{}) (interface{}, error) {
	return oneFromSlice(z.Range(key, 0, 0, false, false))
}

func (z *Zsets) Max(key interface{}) (interface{}, error) {
	return oneFromSlice(z.Range(key, -1, -1, false, false))
}

func (z *Zsets) Range(key interface{}, b, e int, rev, withscores bool) ([]interface{}, error) {
	cmd := "ZRANGE"
	if rev {
		cmd = "ZREVRANGE"
	}
	if withscores {
		return redigo.Values(Do(cmd, z.keystr(key), b, e, "WITHSCORES"))
	} else {
		return redigo.Values(Do(cmd, z.keystr(key), b, e))
	}
}

func (z *Zsets) RangeByScore(
	key interface{}, max float64, maxInclusive bool, min float64, minInclusive bool,
	offset, limit int, rev, withscores bool) ([]interface{}, error) {
	cmd := "ZRANGEBYSCORE"
	if rev {
		cmd = "ZREVRANGEBYSCORE"
	}
	maxstr := strconv.FormatFloat(max, 'f', -1, 64)
	if !maxInclusive {
		maxstr = "(" + maxstr
	}
	minstr := strconv.FormatFloat(min, 'f', -1, 64)
	if !minInclusive {
		minstr = "(" + minstr
	}
	if withscores {
		return redigo.Values(Do(cmd, z.keystr(key), maxstr, minstr, "WITHSCORES", "LIMIT", offset, limit))
	} else {
		return redigo.Values(Do(cmd, z.keystr(key), maxstr, minstr, "LIMIT", offset, limit))
	}
}

func (z *Zsets) MinStr(key interface{}) (string, error) {
	return redigo.String(z.Min(key))
}

func (z *Zsets) MaxStr(key interface{}) (string, error) {
	return redigo.String(z.Max(key))
}

func (z *Zsets) RangeStr(key interface{}, b, e int, rev, withscores bool) ([]string, error) {
	return redigo.Strings(z.Range(key, b, e, rev, withscores))
}

func (z *Zsets) RangeByScoreStr(
	key interface{}, max float64, maxInclusive bool, min float64, minInclusive bool,
	offset, limit int, rev, withscores bool) ([]string, error) {
	return redigo.Strings(z.RangeByScore(
		key, max, maxInclusive, min, minInclusive, offset, limit, rev, withscores))
}

func (z *Zsets) keystr(k interface{}) string {
	return fmt.Sprintf("%s:%v", z.namespace, k)
}

func oneFromSlice(vals []interface{}, err error) (interface{}, error) {
	if err != nil {
		return nil, err
	}
	if len(vals) == 0 {
		return nil, appgo.NotFoundErr
	}
	return vals[0], nil
}
