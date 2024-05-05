//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"github.com/indikay/go-core/server"

	coreConf "github.com/indikay/go-core/conf"
	"github.com/indikay/wallet-service/internal/biz"
	"github.com/indikay/wallet-service/internal/client"
	"github.com/indikay/wallet-service/internal/conf"
	data "github.com/indikay/wallet-service/internal/data"
	"github.com/indikay/wallet-service/internal/messaging"
	"github.com/indikay/wallet-service/internal/queue"
	"github.com/indikay/wallet-service/internal/service"
)

// initApp init kratos application.
func initApp(*coreConf.Server, *conf.Data, log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(data.ProviderSet, messaging.ProviderSet, queue.NewQueue, client.ProviderSet, biz.ProviderSet, service.ProviderSet, server.ProviderSet, initService))
}
