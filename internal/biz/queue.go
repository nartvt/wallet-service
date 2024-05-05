package biz

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/indikay/wallet-service/internal/constant"
	"github.com/shopspring/decimal"
)

const (
	QUEUE_PREFIX      = "ico-subround:"
	ENDROUND_SCHEDULE = "endround"
)

type Task struct {
	ID        string
	Data      string
	Name      string
	ProcessAt time.Time
}

type QueueJob interface {
	Enqueue(ctx context.Context, task *Task) error
	Execute(ctx context.Context, task *Task) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

type QueueRunner struct {
	repo       ICORepo
	walletRepo UserWalletRepo
	transRepo  TransactionRepo
	lockRepo   LockRepo
	icoUc      *ICOUsecase
	log        *log.Helper
}

func NewQueueRunner(repo ICORepo, walletRepo UserWalletRepo, transRepo TransactionRepo, icoUc *ICOUsecase, lockRepo LockRepo) *QueueRunner {
	return &QueueRunner{
		repo:       repo,
		walletRepo: walletRepo,
		transRepo:  transRepo,
		lockRepo:   lockRepo,
		icoUc:      icoUc,
		log:        log.NewHelper(log.DefaultLogger),
	}
}

func (r *QueueRunner) GetICOCurrentRound(ctx context.Context) (*ICOSubRound, error) {
	return r.repo.GetCurrentSubRound(ctx)
}

func (q *QueueRunner) Execute(ctx context.Context, task *Task) error {
	err := q.lockRepo.Lock(ctx, constant.ICO_LOCK)
	if err != nil {
		q.log.Error("ICOTransaction ", err)
	}
	defer func() {
		q.lockRepo.UnLock(ctx, constant.ICO_LOCK)
	}()
	taskName := strings.Replace(task.Name, QUEUE_PREFIX, "", 1)
	if taskName == ENDROUND_SCHEDULE {
		subRound, err := q.repo.GetSubRoundById(ctx, task.Data)
		if err != nil {
			return err
		}
		endedTime := time.Since(subRound.EndAt)
		if !subRound.IsEnded && endedTime >= 0 {
			err := q.icoUc.CloseSubRound(ctx, subRound)
			if err != nil {
				q.log.Error("Execute End Round ", err)
				return nil
			}

			q.repo.WithTx(ctx, func(ctx context.Context) error {
				amount := decimal.RequireFromString(subRound.TotalToken).Sub(decimal.RequireFromString(subRound.BoughtToken)).String()
				rs, err := q.walletRepo.DecreaseBalance(ctx, constant.WALLET_ICO, constant.TokenSymbolIND, amount, constant.WALLET_TYPE_SYSTEM)
				if err != nil || rs == 0 {
					q.log.Error("DecreaseBalance ", err)
					return nil
				}

				rs, err = q.walletRepo.IncreaseBalance(ctx, constant.WALLET_SYS_ICO_BACKUP, constant.TokenSymbolIND, amount, constant.WALLET_TYPE_SYSTEM)
				if err != nil || rs == 0 {
					q.log.Error("IncreaseBalance ", err)
					return nil
				}

				trans := &Transaction{TransType: DEPOSITE, Source: constant.WALLET_ICO, SrcAmount: amount, SrcSymbol: constant.TokenSymbolIND, Destination: constant.WALLET_SYS_ICO_BACKUP, DestSymbol: constant.TokenSymbolIND, DestAmount: amount, Status: TRANS_STATUS, SourceId: fmt.Sprintf("%d-%d", subRound.RoundId, subRound.SubRound)}
				_, err = q.transRepo.CreateTransaction(ctx, trans)
				if err != nil {
					q.log.Error("CreateTransaction ", err)
					return nil
				}

				return nil
			})
		}
	}
	return nil
}
