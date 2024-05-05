package data

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/indikay/wallet-service/ent"
	"github.com/indikay/wallet-service/ent/currencyrate"
	"github.com/indikay/wallet-service/internal/biz"
	"github.com/shopspring/decimal"
)

type currencyRateRepo struct {
	data *Data
	log  *log.Helper
}

func NewCurrencyRepo(data *Data) biz.CurrencyRateRepo {
	return &currencyRateRepo{data: data, log: log.NewHelper(log.DefaultLogger)}
}

func (r *currencyRateRepo) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return r.data.WithTx(ctx, fn)
}

// GetCurrency implements biz.CurrencyRateRepo.
func (r *currencyRateRepo) GetCurrencyRate(ctx context.Context, symbol string) (*biz.CurrencyRate, error) {
	currency, err := r.data.GetClient(ctx).CurrencyRate.Query().Where(currencyrate.SymbolEQ(symbol), currencyrate.ExpiredAtIsNil()).First(ctx)
	if err != nil {
		return nil, err
	}

	return &biz.CurrencyRate{Symbol: currency.Symbol, Rate: currency.Rate}, nil
}

// UpdateCurrencyRate implements biz.CurrencyRateRepo.
func (r *currencyRateRepo) UpdateCurrencyRateICO(ctx context.Context, rate string) error {
	currenciesRate, err := r.data.GetClient(ctx).CurrencyRate.Query().Where(currencyrate.ExpiredAtIsNil()).All(ctx)
	if err != nil {
		return err
	}

	_, err = r.data.GetClient(ctx).CurrencyRate.Update().SetExpiredAt(time.Now()).Where(currencyrate.ExpiredAtIsNil()).Save(ctx)
	if err != nil {
		return err
	}

	updatedRate := make([]*ent.CurrencyRateCreate, len(currenciesRate))
	for i, v := range currenciesRate {
		newRate := decimal.RequireFromString(v.Rate).Mul(decimal.RequireFromString(rate))
		updatedRate[i] = r.data.GetClient(ctx).CurrencyRate.Create().SetSymbol(v.Symbol).SetRate(newRate.String())
	}

	_, err = r.data.GetClient(ctx).CurrencyRate.CreateBulk(updatedRate...).Save(ctx)
	return err
}

func (r *currencyRateRepo) InitData(ctx context.Context) error {
	r.data.GetClient(ctx).CurrencyRate.Create().SetSymbol("VND_IND").SetRate("550").Save(ctx)
	r.data.GetClient(ctx).CurrencyRate.Create().SetSymbol("USD_IND").SetRate("0.022").Save(ctx)
	r.data.GetClient(ctx).CurrencyRate.Create().SetSymbol("USDT_IND").SetRate("0.022").Save(ctx)
	return nil
}
