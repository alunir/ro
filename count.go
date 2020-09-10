package ro

import (
	"context"

	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"

	"github.com/alunir/ro/rq"
)

// Count implements the types.Store interface.
func (s *redisStore) Count(ctx context.Context, mods ...rq.Modifier) (int, error) {
	conn, err := s.pool.GetContext(ctx)
	if err != nil {
		return 0, errors.Wrap(err, "failed to acquire a connection")
	}
	defer conn.Close()

	cmd, err := s.injectKeyPrefix(rq.Count(mods...)).Build()
	if err != nil {
		return 0, errors.WithStack(err)
	}

	cnt, err := redis.Int(conn.Do(cmd.Name, cmd.Args...))
	if err != nil {
		return 0, errors.Wrapf(err, "faild to execute %v", cmd)
	}
	return cnt, nil
}
