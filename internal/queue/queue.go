package queue

import (
	"context"
	"crypto/tls"
	"fmt"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/hibiken/asynq"
	"github.com/indikay/wallet-service/internal/biz"
	"github.com/indikay/wallet-service/internal/conf"
)

type redisQueue struct {
	*biz.QueueRunner
	asynqCli  *asynq.Client
	asynqSrv  *asynq.Server
	asynqConf asynq.RedisClientOpt
	log       *log.Helper
}

// NewData .
func NewQueue(c *conf.Data, repo biz.ICORepo, walletRepo biz.UserWalletRepo, transRepo biz.TransactionRepo, icoUc *biz.ICOUsecase, lockRepo biz.LockRepo) biz.QueueJob {
	logHelper := log.NewHelper(log.DefaultLogger)

	clienConfig := asynq.RedisClientOpt{Addr: c.Redis.Addr, Password: c.Redis.Pass}
	if c.Redis.Ssl {
		clienConfig.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
	}
	client := asynq.NewClient(clienConfig)
	queue := &redisQueue{asynqCli: client, asynqConf: clienConfig, log: logHelper, QueueRunner: biz.NewQueueRunner(repo, walletRepo, transRepo, icoUc, lockRepo)}

	return queue
}

// Enqueue implements biz.QueueJob.
func (r *redisQueue) Enqueue(ctx context.Context, bizTask *biz.Task) error {
	// new task without conflict with other
	t := asynq.NewTask(bizTask.Name, []byte(bizTask.Data), asynq.ProcessAt(bizTask.ProcessAt), asynq.MaxRetry(1))
	info, err := r.asynqCli.Enqueue(t, asynq.TaskID(fmt.Sprintf("%s-%d", bizTask.Data, bizTask.ProcessAt.Unix())))
	log.Infof("Enqueue %v - %s", info != nil, bizTask.ProcessAt)
	if err != nil {
		log.Errorf("Enqueue -Info: %v, ProcessAt:  %s, Error: %v", info, bizTask.ProcessAt, err)
	} else {
		log.Infof("Enqueue %s - %s", info.ID, bizTask.ProcessAt)
	}
	return nil
}

// Execute implements biz.QueueJob.
func (r *redisQueue) Start(ctx context.Context) error {
	r.asynqSrv = asynq.NewServer(
		r.asynqConf,
		asynq.Config{
			// Specify how many concurrent workers to use
			Concurrency: 3,
			// Optionally specify multiple queues with different priority.
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
			// See the godoc for other configuration options
		},
	)

	mux := asynq.NewServeMux()
	mux.HandleFunc(biz.QUEUE_PREFIX, func(ctx context.Context, t *asynq.Task) error {
		err := r.Execute(ctx, &biz.Task{Name: t.Type(), Data: string(t.Payload())})
		if err != nil {
			return err
		}
		round, err := r.QueueRunner.GetICOCurrentRound(ctx)
		if err != nil {
			return err
		}
		r.Enqueue(ctx, &biz.Task{Data: round.ID.String(), Name: fmt.Sprintf("%s%s", biz.QUEUE_PREFIX, biz.ENDROUND_SCHEDULE), ProcessAt: round.EndAt})
		return nil
	})

	return r.asynqSrv.Run(mux)
}

func (r *redisQueue) Stop(ctx context.Context) error {
	r.asynqSrv.Shutdown()
	return r.asynqCli.Close()
}
