package main

import (
	"flag"
	"os"

	icoProto "github.com/indikay/wallet-service/api/ico/v1"
	apiProto "github.com/indikay/wallet-service/api/wallet/v1"

	"github.com/indikay/wallet-service/internal/biz"
	"github.com/indikay/wallet-service/internal/conf"

	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"

	"github.com/indikay/wallet-service/internal/service"
	"github.com/joho/godotenv"

	"github.com/go-kratos/kratos/v2"
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

func initService(logger log.Logger, hs *http.Server, gs *grpc.Server,
	userToken *service.UserWalletService,
	transaction *service.TransactionService, ico *service.ICOService, queue biz.QueueJob) *kratos.App {
	apiProto.RegisterUserWalletServiceServer(gs, userToken)
	apiProto.RegisterUserWalletServiceHTTPServer(hs, userToken)

	apiProto.RegisterTransactionServiceServer(gs, transaction)
	apiProto.RegisterTransactionServiceHTTPServer(hs, transaction)

	icoProto.RegisterICOServiceHTTPServer(hs, ico)
	icoProto.RegisterICOServiceServer(gs, ico)

	return kratos.New(kratos.ID(id), kratos.Name(Name), kratos.Version(Version), kratos.Server(hs, gs, queue))
}

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
	app, cleanup, err := initApp(bc.Server, bc.Data, log.DefaultLogger)
	if err != nil {
		panic(err)
	}

	// start and wait for stop signal
	if err := app.Run(); err != nil {
		panic(err)
	}
	cleanup()
}
