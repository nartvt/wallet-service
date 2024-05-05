package data

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/indikay/wallet-service/ent"
	"github.com/indikay/wallet-service/ent/predicate"
	"github.com/indikay/wallet-service/ent/transaction"
	"github.com/indikay/wallet-service/internal/biz"
	"github.com/rs/xid"
)

type transactionRepo struct {
	data *Data
	log  *log.Helper
}

func NewTransactionRepo(data *Data) biz.TransactionRepo {
	return &transactionRepo{data: data, log: log.NewHelper(log.DefaultLogger)}
}

func (r *transactionRepo) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return r.data.WithTx(ctx, fn)
}

// CreateTransaction implements biz.TransactionRepo.
func (r *transactionRepo) CreateTransaction(ctx context.Context, input *biz.Transaction) (*biz.Transaction, error) {
	rs, err := r.data.GetClient(ctx).Transaction.Create().SetTransType(input.TransType).
		SetSource(input.Source).SetSrcSymbol(input.SrcSymbol).SetSrcAmount(input.SrcAmount).
		SetDestination(input.Destination).SetDestSymbol(input.DestSymbol).SetDestAmount(input.DestAmount).
		SetRate(input.Rate).SetSourceService(input.SourceService).SetSourceID(input.SourceId).SetStatus(input.Status).Save(ctx)

	if err != nil {
		return nil, err
	}

	return r.mapToBiz(rs), nil
}

// GetTransactionsByUserId implements biz.TransactionRepo.
func (r *transactionRepo) GetTransactionsByUserId(ctx context.Context, userId string, cursor string, limit int32) ([]*biz.Transaction, string, error) {
	where := []predicate.Transaction{transaction.TransTypeNotIn(biz.DEPOSITE), transaction.Or(transaction.Source(userId), transaction.Destination(userId))}
	if len(cursor) > 0 {
		id, err := xid.FromString(cursor)
		if err != nil {
			return nil, "", err
		}
		where = append(where, transaction.IDLT(id))
	}

	trans, err := r.data.GetClient(ctx).Transaction.Query().Where(where...).Order(ent.Desc(transaction.FieldCreatedAt)).Limit(int(limit)).All(ctx)
	if err != nil {
		return nil, "", err
	}

	var rs = make([]*biz.Transaction, len(trans))
	next := ""
	for i, tr := range trans {
		rs[i] = r.mapToBiz(tr)
		next = tr.ID.String()
	}

	if len(rs) < int(limit) {
		next = ""
	}

	return rs, next, nil
}

func (r *transactionRepo) mapToBiz(en *ent.Transaction) *biz.Transaction {
	return &biz.Transaction{ID: en.ID, TransType: en.TransType, Source: en.Source, SrcSymbol: en.SrcSymbol, SrcAmount: en.SrcAmount,
		Destination: en.Destination, DestSymbol: en.DestSymbol, DestAmount: en.DestAmount, Rate: en.Rate, SourceService: en.SourceService,
		Status: en.Status, CreatedAt: en.CreatedAt, UpdatedAt: en.UpdatedAt}
}
