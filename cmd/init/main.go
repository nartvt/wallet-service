package main

import (
	"context"
	"flag"
	"os"

	"github.com/indikay/wallet-service/internal/conf"

	"github.com/joho/godotenv"

	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/env"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/tracing"

	logcore "github.com/indikay/go-core/log"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	// Name is the name of the compiled software.
	Name string
	// Version is the version of the compiled software.
	Version string
	// flagconf is the config flag.
	flagconf string

	id, _ = os.Hostname()
)

func init() {
	flag.StringVar(&flagconf, "conf", "../../configs", "config path, eg: -conf config.yaml")
	godotenv.Load()
}

func main() {
	flag.Parse()

	c := config.New(
		config.WithSource(
			env.NewSource("IND_"),
			file.NewSource(flagconf),
		),
		config.WithResolver(CustomResolver),
	)
	if err := c.Load(); err != nil {
		panic(err)
	}

	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		panic(err)
	}

	log.DefaultLogger = log.With(logcore.LogrusConfig(bc.Server.Log),
		"ts", log.DefaultTimestamp,
		"caller", log.DefaultCaller,
		"service.id", id,
		"service.name", Name,
		"service.version", Version,
		"trace_id", tracing.TraceID(),
		"span_id", tracing.SpanID(),
	)
	uc, cleanup, err := initApp(bc.Data)
	if err != nil {
		panic(err)
	}
	uc.InitData(context.Background())
	// err = walletRepo.InitData(context.Background())
	// log.Errorf("err %v", err)
	cleanup()
}
