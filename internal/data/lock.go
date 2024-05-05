package data

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	redsync "github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/indikay/wallet-service/internal/biz"
)

type lockRepo struct {
	redisLock *redsync.Redsync
	log       *log.Helper
}

func NewLockRepo(data *Data) biz.LockRepo {
	pool := goredis.NewPool(data.redisCli.GetClient())
	rs := redsync.New(pool)
	return &lockRepo{redisLock: rs, log: log.NewHelper(log.DefaultLogger)}
}

// Lock implements biz.LockRepo.
func (l *lockRepo) Lock(ctx context.Context, key string) error {
	mutex := l.redisLock.NewMutex(key, redsync.WithRetryDelay(time.Second), redsync.WithTries(100))
	if err := mutex.LockContext(ctx); err != nil {
		return err
	}
	return nil
}

// UnLock implements biz.LockRepo.
func (l *lockRepo) UnLock(ctx context.Context, key string) error {
	mutex := l.redisLock.NewMutex(key)
	if ok, err := mutex.UnlockContext(ctx); !ok {
		return err
	}
	return nil
}
