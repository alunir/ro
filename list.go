package ro

import (
	"context"
	"reflect"

	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"

	"github.com/alunir/ro/rq"
)

// List implements the types.Store interface.
func (s *redisStore) List(ctx context.Context, dest interface{}, mods ...rq.Modifier) error {
	dt := reflect.ValueOf(dest)
	if dt.Kind() != reflect.Ptr || dt.IsNil() {
		return errors.New("must pass a slice ptr")
	}
	dt = dt.Elem()
	if dt.Kind() != reflect.Slice {
		return errors.New("must pass a slice ptr")
	}

	// keys, err := s.selectKeys(ctx, mods)
	// if err != nil {
	// 	return errors.Wrap(err, "failed to select query")
	// }

	conn, err := s.pool.GetContext(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to acquire a connection")
	}
	defer conn.Close()

	var keys []string
	if s.HashStoreEnabled {
		cmd, err := s.injectKeyPrefix(rq.List(mods...)).Build()
		if err != nil {
			return errors.WithStack(err)
		}

		keys, err = redis.Strings(conn.Do(cmd.Name, cmd.Args...))
		if err != nil {
			return errors.WithStack(err)
		}

		for _, key := range keys {
			err := conn.Send("HGETALL", key)
			if err != nil {
				return errors.Wrapf(err, "failed to send HGETALL %s", key)
			}
		}
	} else if len(s.model.Serialized()) > 0 {
		cmd, err := rq.List(mods...).Build()
		if err != nil {
			return errors.WithStack(err)
		}

		keys, err = redis.Strings(conn.Do(cmd.Name, cmd.Args...))
		if err != nil {
			return errors.WithStack(err)
		}

		err = conn.Send("HMGET", s.KeyPrefix, keys)
		if err != nil {
			return errors.Wrapf(err, "faild to send HMGET %s %s", s.KeyPrefix, keys)
		}
	}

	conn.Flush()

	vt := dt.Type().Elem().Elem()

	for _, key := range keys {
		v, err := redis.Values(conn.Receive())
		if err != nil {
			return errors.Wrap(err, "faild to receive or cast redis command result")
		}
		vv := reflect.New(vt)
		err = redis.ScanStruct(v, vv.Interface())
		if err != nil {
			return errors.Wrapf(err, "faild to scan struct %s %x", key, v)
		}
		dt.Set(reflect.Append(dt, vv))
	}

	return nil
}
