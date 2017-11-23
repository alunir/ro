package store

import (
	"reflect"

	"github.com/garyburd/redigo/redis"
)

// Remove implements the types.Store interface.
func (s *ConcreteStore) Remove(src interface{}) error {
	keys := []string{}

	rv := reflect.ValueOf(src)
	if rv.Kind() == reflect.Slice {
		for i := 0; i < rv.Len(); i++ {
			m, err := s.toModel(rv.Index(i))
			if err != nil {
				return err
			}
			keys = append(keys, s.getKey(m))
		}
	} else {
		m, err := s.toModel(rv)
		if err != nil {
			return err
		}
		keys = append(keys, s.getKey(m))
	}

	return s.removeByKeys(keys)
}

func (s *ConcreteStore) removeByKeys(keys []string) error {
	conn := s.getConn()
	defer conn.Close()

	keysByZsetKey := map[string][]string{}
	for _, k := range keys {
		zsetKeys, err := redis.Strings(conn.Do("SMEMBERS", s.getScoreSetKeysKeyByKey(k)))
		if err != nil {
			return err
		}
		for _, zk := range zsetKeys {
			keysByZsetKey[zk] = append(keysByZsetKey[zk], k)
		}
	}

	err := conn.Send("MULTI")
	if err != nil {
		return err
	}

	err = conn.Send("DEL", redis.Args{}.AddFlat(keys)...)
	if err != nil {
		return err
	}

	for zk, hkeys := range keysByZsetKey {
		err = conn.Send("ZREM", redis.Args{}.Add(zk).AddFlat(hkeys)...)
		if err != nil {
			return err
		}
	}

	_, err = conn.Do("EXEC")
	return err
}
