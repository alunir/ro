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

	keys, err := s.selectKeys(ctx, mods)
	if err != nil {
		return errors.Wrap(err, "failed to select query")
	}

	conn, err := s.pool.GetContext(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to acquire a connection")
	}
	defer conn.Close()

	if s.HashStoreEnabled {
		for _, key := range keys {
			err := conn.Send("HGETALL", key)
			if err != nil {
				return errors.Wrapf(err, "failed to send HGETALL %s", key)
			}
		}
	} else if len(s.model.Serialized()) > 0 {
		err := conn.Send("HGETALL", s.KeyPrefix)
		if err != nil {
			return errors.Wrapf(err, "faild to send HGETALL %s", s.KeyPrefix)
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
