package data

import (
	"context"
	"database/sql"
	"fmt"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/schema"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"github.com/indikay/go-core/database/redisdb"
	"github.com/indikay/wallet-service/ent"
	"github.com/indikay/wallet-service/ent/migrate/migratedata"
	"github.com/indikay/wallet-service/internal/conf"

	// init postgres driver
	_ "github.com/jackc/pgx/v5/stdlib"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewIcoRepo, NewWalletRepo, NewTransactionRepo, NewCurrencyRepo, NewIcoCouponRepo, NewLockRepo)

type Data struct {
	db       *ent.Client
	sql      *entsql.Driver
	redisCli *redisdb.RedisClient
}

var DataClient *ent.Client

// NewData .
func NewData(conf *conf.Data) (*Data, func(), error) {
	log := log.NewHelper(log.DefaultLogger)
	db, err := sql.Open(
		conf.Database.Driver,
		conf.Database.Source,
	)
	if err != nil {
		log.Errorf("failed opening connection to db: %v", err)
		return nil, nil, err
	}

	sqlDrv := entsql.OpenDB(dialect.Postgres, db)
	// sqlDrv := dialect.DebugWithContext(drv, func(ctx context.Context, i ...interface{}) {
	// 	log.WithContext(ctx).Debug(i...)
	// 	tracer := otel.Tracer("ent.")
	// 	kind := trace.SpanKindServer
	// 	_, span := tracer.Start(ctx,
	// 		"Query",
	// 		trace.WithAttributes(
	// 			attribute.String("sql", fmt.Sprint(i...)),
	// 		),
	// 		trace.WithSpanKind(kind),
	// 	)
	// 	span.End()
	// })
	client := ent.NewClient(ent.Driver(sqlDrv))

	// Run the auto migration tool.
	if err := client.Schema.Create(context.Background(), schema.WithApplyHook(migratedata.ApplyTokenomic)); err != nil {
		log.Errorf("failed creating schema resources: %v", err)
		return nil, nil, err
	}

	DataClient = client
	d := &Data{
		db:       client,
		sql:      sqlDrv,
		redisCli: redisdb.NewRedisClient(conf.Redis),
	}
	return d, func() {
		log.Info("message", "closing the data resources")
		if err := d.db.Close(); err != nil {
			log.Error(err)
		}
	}, nil
}

func (d *Data) GetSQL(ctx context.Context) *entsql.Driver {
	return d.sql
}

func (d *Data) GetClient(ctx context.Context) *ent.Client {
	if tx, ok := d.TxFromContext(ctx); ok {
		return tx.Client()
	}

	return d.db
}

// avoid error: should not use built-in type string as key for value; define your own type to avoid collisions
type entTransaction string

const (
	entTransactionKey = entTransaction("entTransaction")
)

func (d *Data) NewTxContext(ctx context.Context, tx *ent.Tx) context.Context {
	return context.WithValue(ctx, entTransactionKey, tx)
}

func (d *Data) TxFromContext(ctx context.Context) (tx *ent.Tx, ok bool) {
	tx, ok = ctx.Value(entTransactionKey).(*ent.Tx)
	return
}

func (d *Data) TxCompleted(ctx context.Context) context.Context {
	return context.WithValue(ctx, entTransactionKey, nil)
}

func (r *Data) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	var err error
	tx, ok := r.TxFromContext(ctx)
	if !ok || tx == nil {
		log.Infof("Transaction: %v", tx)
		ntx, err := r.GetClient(ctx).Tx(ctx)
		if err != nil {
			return err
		}
		tx = ntx
		ctx = r.NewTxContext(ctx, tx)
	}
	// tx, err := r.db.Tx(ctx)
	// if err != nil {
	// 	return err
	// }
	newCtx := r.NewTxContext(ctx, tx)
	defer func() {
		log.Infof("Defer Recovered: %v, Transaction : %v", err, tx != nil)
		if err == nil {
			return
		}
		if v := recover(); v != nil {
			log.Infof("Panic Recovered: %v, Transaction : %v", v, tx != nil)
			if tx != nil {
				tx.Rollback()
			}
			panic(v)
		}
	}()
	if err := fn(newCtx); err != nil {
		if rerr := tx.Rollback(); rerr != nil {
			err = fmt.Errorf("%w: rolling back transaction: %v", err, rerr)
		}
		return err
	}
	if err := tx.Commit(); err != nil {
		err = fmt.Errorf("committing transaction: %w", err)
		return err
	}

	return nil
}
