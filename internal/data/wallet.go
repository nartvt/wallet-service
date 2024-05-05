package data

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/indikay/wallet-service/ent"
	"github.com/indikay/wallet-service/ent/predicate"
	"github.com/indikay/wallet-service/ent/userwallet"
	"github.com/indikay/wallet-service/internal/biz"
	"github.com/indikay/wallet-service/internal/constant"
)

type walletRepo struct {
	data *Data
	log  *log.Helper
}

func NewWalletRepo(data *Data) biz.UserWalletRepo {
	return &walletRepo{data: data, log: log.NewHelper(log.DefaultLogger)}
}

func (r *walletRepo) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return r.data.WithTx(ctx, fn)
}

// CreateWallet implements biz.UserWalletRepo.
func (r *walletRepo) CreateWallet(ctx context.Context, userId string, symbol, walletType string) (*biz.UserWallet, error) {
	err := r.data.GetClient(ctx).UserWallet.Create().SetUserID(userId).SetSymbol(symbol).SetBalance("0").SetType(walletType).SetIsActive(true).OnConflictColumns(userwallet.FieldUserID, userwallet.FieldSymbol, userwallet.FieldType).DoNothing().Exec(ctx)
	if err != nil {
		return nil, err
	}

	we, err := r.data.GetClient(ctx).UserWallet.Query().Where(userwallet.UserID(userId), userwallet.Symbol(symbol)).First(ctx)
	if err != nil {
		return nil, err
	}

	return r.mapToBiz(we), nil
}

func (r *walletRepo) CalculateBalance(ctx context.Context, userId, amount, symbol string) (bool, error) {
	resp, err := r.data.GetClient(ctx).UserWallet.Query().
		Where(userwallet.UserID(userId), userwallet.Symbol(symbol)).
		Where(userwallet.BalanceGTE(amount)).
		First(ctx)
	if err != nil {
		return false, err
	}
	return resp != nil, nil
}

// DecreaseBalance implements biz.UserWalletRepo.
func (r *walletRepo) DecreaseBalance(ctx context.Context, userId string, symbol string, amount string, walletType string) (int, error) {
	return r.data.GetClient(ctx).UserWallet.Update().Modify(func(u *sql.UpdateBuilder) {
		u.Set(userwallet.FieldBalance, sql.ExprFunc(func(b *sql.Builder) {
			b.Ident(userwallet.FieldBalance).WriteOp(sql.OpSub).Arg(amount)
		}))
	}).Where(userwallet.BalanceGTE(amount), userwallet.UserID(userId), userwallet.Symbol(symbol), userwallet.TypeEQ(walletType)).Save(ctx)
}

// IncreaseBalance implements biz.UserWalletRepo.
func (r *walletRepo) IncreaseBalance(ctx context.Context, userId string, symbol string, amount string, walletType string) (int, error) {
	return r.data.GetClient(ctx).UserWallet.Update().Modify(func(u *sql.UpdateBuilder) {
		u.Set(userwallet.FieldBalance, sql.ExprFunc(func(b *sql.Builder) {
			b.Ident(userwallet.FieldBalance).WriteOp(sql.OpAdd).Arg(amount)
		}))
	}).Where(userwallet.UserID(userId), userwallet.Symbol(symbol), userwallet.TypeEQ(walletType)).Save(ctx)
}

// GetWalletByUserId implements biz.UserWalletRepo.
func (r *walletRepo) GetWalletByUserId(ctx context.Context, userId, symbol, walletType string) ([]*biz.UserWallet, error) {
	where := []predicate.UserWallet{userwallet.UserIDEQ(userId)}
	if len(symbol) > 0 {
		where = append(where, userwallet.Symbol(symbol))
	}

	if len(walletType) > 0 {
		where = append(where, userwallet.Type(walletType))
	}
	wallets, err := r.data.GetClient(ctx).UserWallet.Query().Where(where...).All(ctx)
	if err != nil {
		return nil, err
	}

	var rs = make([]*biz.UserWallet, len(wallets))
	for i, uw := range wallets {
		rs[i] = r.mapToBiz(uw)
	}
	return rs, nil
}

func (r *walletRepo) mapToBiz(en *ent.UserWallet) *biz.UserWallet {
	return &biz.UserWallet{ID: en.ID, UserID: en.UserID, Type: en.Type, Symbol: en.Symbol, Balance: en.Balance, IsActive: en.IsActive,
		CreatedAt: en.CreatedAt, UpdatedAt: en.UpdatedAt}
}

func (r *walletRepo) InitData(ctx context.Context) error {
	err := r.data.GetClient(ctx).UserWallet.Create().SetUserID(constant.WALLET_ICO).SetSymbol(constant.TokenSymbolIND).SetBalance("400000000").SetType(constant.WALLET_TYPE_SYSTEM).SetIsActive(true).OnConflictColumns(userwallet.FieldUserID, userwallet.FieldSymbol, userwallet.FieldType).DoNothing().Exec(ctx)

	err = r.data.GetClient(ctx).UserWallet.Create().SetUserID(constant.WALLET_SYS_ICO_BACKUP).SetSymbol(constant.TokenSymbolIND).SetBalance("0").SetType(constant.WALLET_TYPE_SYSTEM).SetIsActive(true).OnConflictColumns(userwallet.FieldUserID, userwallet.FieldSymbol, userwallet.FieldType).DoNothing().Exec(ctx)

	return err
}
