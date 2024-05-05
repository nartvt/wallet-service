//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"github.com/google/wire"

	"github.com/indikay/wallet-service/internal/biz"
	"github.com/indikay/wallet-service/internal/conf"
	"github.com/indikay/wallet-service/internal/data"
	"github.com/indikay/wallet-service/internal/messaging"
	"github.com/indikay/wallet-service/internal/queue"
)

// initApp init kratos application.
func initApp(*conf.Data) (*biz.WalletTransactionUseCase, func(), error) {
	panic(wire.Build(data.ProviderSet, messaging.ProviderSet, queue.NewQueue, biz.ProviderSet)) //data.NewIcoRepo,
}
