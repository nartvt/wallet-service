package service

import "github.com/google/wire"

var ProviderSet = wire.NewSet(NewTransactionService, NewUserWalletService, NewICOService)
