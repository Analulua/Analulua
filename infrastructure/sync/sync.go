package sync

import (
	"context"
	goredislib "github.com/go-redis/redis/v8"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v8"
	"newdemo1/resource"
)

type (
	Sync interface {
		Lock(ctx context.Context, key string, options ...redsync.Option) (Unlock, error)
	}
	sync struct {
		resource *resource.Resource
		redsSync *redsync.Redsync
	}

	Unlock interface {
		Unlock(ctx context.Context) error
	}
	unlock struct {
		mutex *redsync.Mutex
	}
)

func (s *sync) Lock(ctx context.Context, key string, options ...redsync.Option) (Unlock, error) {
	mutex := s.redsSync.NewMutex(key, options...)
	err := mutex.LockContext(ctx)
	if err != nil {
		return nil, err
	}
	return &unlock{
		mutex: mutex,
	}, nil

}
func (u *unlock) Unlock(ctx context.Context) error {
	_, err := u.mutex.UnlockContext(ctx)
	return err
}
func New(resource *resource.Resource) Sync {
	client := goredislib.NewClient(&goredislib.Options{
		Addr:     resource.Credential.Redis.Host,
		Password: resource.Credential.Redis.Password,
	})

	pool := goredis.NewPool(client)
	rs := redsync.New(pool)
	return &sync{
		resource: resource,
		redsSync: rs,
	}
}
