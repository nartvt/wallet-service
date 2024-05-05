package biz

import (
	"github.com/google/wire"
	"github.com/shopspring/decimal"
)

// ProviderSet is biz providers.
var (
	ProviderSet      = wire.NewSet(NewICOUseCase, NewWalletTransactionUseCase)
	usdt_fee_percent = decimal.NewFromFloat32(0.125).Div(decimal.NewFromInt(100))
)
