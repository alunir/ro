package ro

import (
	"context"

	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
)

func (s *redisStore) Get(ctx context.Context, dests ...Model) error {
	conn, err := s.pool.GetContext(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to acquire a connection")
	}
	defer conn.Close()

	keys := make([]string, len(dests), len(dests))

	for i, m := range dests {
		var key string
		if s.HashStoreEnabled {
			key, err = s.getKey(m)
			if err != nil {
				return errors.Wrap(err, "failed to get key")
			}
		} else {
			if len(m.Serialized()) == 0 {
				return errors.Errorf("failed to implement Serialized %v", m)
			}
			key = s.KeyPrefix
		}
		keys[i] = key
		err = conn.Send("HGETALL", key)
		if err != nil {
			return errors.Wrapf(err, "faild to send HGETALL %s", key)
		}
	}

	err = conn.Flush()
	if err != nil {
		return errors.Wrap(err, "faild to flush commands")
	}

	for i, d := range dests {
		v, err := redis.Values(conn.Receive())
		if err != nil {
			return errors.Wrap(err, "faild to receive or cast redis command result")
		}
		if s.HashStoreEnabled {
			err = redis.ScanStruct(v, d)
			if err != nil {
				return errors.Wrapf(err, "faild to scan struct %s %x", keys[i], v)
			}
		} else {
			var key string
			for j, vv := range v {
				if j%2 == 0 {
					key = string(vv.([]byte))
					continue
				} else if key == d.GetKeySuffix() {
					d.Deserialized(vv.([]byte))
					break
				}
			}
		}
	}

	return nil
}
