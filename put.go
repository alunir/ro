package ro

import (
	"context"
	"fmt"
	"reflect"
	"strconv"

	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
)

// Put implements the types.Store interface.
func (s *redisStore) Put(ctx context.Context, src interface{}, ttl int) error {
	if ttl == 0 {
		return errors.New("ttl is zero")
	}

	conn, err := s.pool.GetContext(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to acquire a connection")
	}
	defer conn.Close()

	// TODO: watch all key if s.HashStoreEnabled
	if !s.HashStoreEnabled {
		key := s.KeyPrefix
		_, err = conn.Do("WATCH", key)
		if err != nil {
			return errors.Wrapf(err, "failed to send WATCH %s", key)
		}
	}

	err = conn.Send("MULTI")
	if err != nil {
		return errors.Wrap(err, "faild to send MULTI command")
	}

	conn.Send("SELECT", s.model.GetDatabaseNo())

	rv := reflect.ValueOf(src)
	if rv.Kind() == reflect.Slice {
		for i := 0; i < rv.Len(); i++ {
			err = s.set(conn, rv.Index(i), ttl)
			if err != nil {
				break
			}
		}
	} else {
		err = s.set(conn, rv, ttl)
	}

	if err != nil {
		conn.Do("DISCARD")
		return errors.Wrap(err, "faild to send any commands")
	}

	r, err := redis.Values(conn.Do("EXEC"))
	if err != nil {
		return errors.Wrap(err, "faild to EXEC commands")
	}
	if r[0] != "OK" {
		return errors.Wrap(err, "return FAILED after EXEC commands")
	}
	return nil
}

func (s *redisStore) set(conn redis.Conn, src reflect.Value, ttl int) error {
	m, err := s.toModel(src)
	if err != nil {
		return errors.Wrap(err, "failed to convert to model")
	}

	var key string
	if s.HashStoreEnabled {
		key, err = s.getKey(m)
		if err != nil {
			return errors.Wrap(err, "failed to get key")
		}
		err = conn.Send("HMSET", redis.Args{}.Add(key).AddFlat(m)...)
		if err != nil {
			return errors.Wrapf(err, "failed to send HMSET %s %v", key, m)
		}
		err = conn.Send("EXPIRE", key, ttl)
		if err != nil {
			return errors.Wrapf(err, "failed to send EXPIRE %s %v", key, m)
		}
	} else {
		key = m.GetKeySuffix()
		if len(m.Serialized()) == 0 {
			return errors.Errorf("failed to implement Serialized %s %v", key, m)
		}
		err = conn.Send("HSET", s.KeyPrefix, m.GetKeySuffix(), m.Serialized())
		if err != nil {
			return errors.Wrapf(err, "failed to send HSET %s %s", key, m.GetKeySuffix())
		}
		err = conn.Send("EXPIRE", s.KeyPrefix, ttl)
		if err != nil {
			return errors.Wrapf(err, "failed to send EXPIRE %s %s", key, m.GetKeySuffix())
		}
	}

	scoreMap := m.GetScoreMap()
	if scoreMap == nil {
		return errors.Errorf("%s's GetScoreMap() should be present", key)
	} else if len(scoreMap) == 0 {
		return nil
	}

	zsetKeys := make([]string, 0, len(scoreMap))
	for ks, score := range scoreMap {
		if len(ks) == 0 {
			return errors.Errorf("key in %s's GetScoreMap() should be present", key)
		}
		_, err := strconv.ParseFloat(fmt.Sprint(score), 64)
		if err != nil {
			return errors.Wrapf(err, "%s's GetScoreMap()[%s] should be number", key, ks)
		}
		scoreSetKey := s.getScoreSetKey(ks)
		err = conn.Send("ZADD", scoreSetKey, score, key)
		if err != nil {
			return errors.Wrapf(err, "failed to send ZADD %s %v %s", scoreSetKey, score, key)
		}
		err = conn.Send("EXPIRE", scoreSetKey, ttl)
		if err != nil {
			return errors.Wrapf(err, "failed to send EXPIRE %s %v %s", scoreSetKey, score, key)
		}
		zsetKeys = append(zsetKeys, scoreSetKey)
	}

	scoreSetKeysKey := s.getScoreSetKeysKeyByKey()
	err = conn.Send("SADD", redis.Args{}.Add(scoreSetKeysKey).AddFlat(zsetKeys)...)
	if err != nil {
		return errors.Wrapf(err, "failed to send SADD %s %v", scoreSetKeysKey, zsetKeys)
	}
	err = conn.Send("EXPIRE", scoreSetKeysKey, ttl)
	if err != nil {
		return errors.Wrapf(err, "failed to send EXPIRE %s %v", scoreSetKeysKey, zsetKeys)
	}

	return nil
}
