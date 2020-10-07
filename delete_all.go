package ro

import (
	"context"

	"github.com/pkg/errors"

	"github.com/alunir/ro/rq"
)

// DeleteAll implements the types.Store interface.
func (s *redisStore) DeleteAll(ctx context.Context, mods ...rq.Modifier) error {
	conn, err := s.pool.GetContext(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to acquire a connection")
	}
	defer conn.Close()
	conn.Do("SELECT", s.model.GetDatabaseNo())

	keys, err := s.selectKeys(conn, mods)
	if err != nil {
		return errors.WithStack(err)
	}
	err = s.deleteByKeys(ctx, keys)
	if err != nil {
		return errors.Wrapf(err, "failed to remove by keys %v", keys)
	}
	return nil
}
