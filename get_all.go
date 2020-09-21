package ro

import (
	"context"

	"github.com/pkg/errors"

	"github.com/alunir/ro/rq"
)

// GetAll implements the types.Store interface.
func (s *redisStore) GetAll(ctx context.Context, dest Model, mods ...rq.Modifier) error {
	keys, err := s.selectKeys(ctx, mods)
	if err != nil {
		return errors.WithStack(err)
	}
	err = s.getByKeys(ctx, dest, keys)
	if err != nil {
		return errors.Wrapf(err, "failed to get by keys %v", keys)
	}
	return nil
}
