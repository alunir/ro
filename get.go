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
		key, err := s.getKey(m)
		if err != nil {
			return errors.Wrap(err, "failed to get key")
		}
		keys[i] = key
		if s.HashStoreEnabled {
			err = conn.Send("HGETALL", key)
			if err != nil {
				return errors.Wrapf(err, "faild to send HGETALL %s", key)
			}
		}
	}

	if !s.HashStoreEnabled {
		err = conn.Send("MGET", keys)
		if err != nil {
			return errors.Wrapf(err, "faild to send MGET %s", keys)
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
		err = redis.ScanStruct(v, d)
		if err != nil {
			return errors.Wrapf(err, "faild to scan struct %s %x", keys[i], v)
		}
	}

	return nil
}

func (s *redisStore) getByKeys(ctx context.Context, dest Model, keys []string) error {
	conn, err := s.pool.GetContext(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to acquire a connection")
	}

	if len(keys) > 0 {
		err = conn.Send("MGET", redis.Args{}.AddFlat(keys)...)
		if err != nil {
			return errors.Wrapf(err, "faild to send MGET %v", keys)
		}
	}

	err = conn.Flush()
	if err != nil {
		return errors.Wrap(err, "faild to flush commands")
	}

	v, err := redis.Values(conn.Receive())
	if err != nil {
		return errors.Wrap(err, "faild to receive or cast redis command result")
	}
	err = redis.ScanStruct(v, dest)
	if err != nil {
		return errors.Wrapf(err, "faild to scan struct %s %x", keys, v)
	}

	return nil
}
