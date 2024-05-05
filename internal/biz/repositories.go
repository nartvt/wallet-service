package biz

import (
	"context"
	"time"

	"github.com/rs/xid"
)

type ICORound struct {
	ID        xid.ID
	RoundId   int32
	RoundName string
	Price     string
	NumToken  string
	PriceGap  string
	NumSub    int32
	EndedAt   *time.Time
}

type ICOSubRound struct {
	ID          xid.ID
	RoundId     int32
	RoundName   string
	SubRound    int32
	Price       string
	BoughtToken string
	TotalToken  string
	EndAt       time.Time
	IsEnded     bool
}

type ICOHistory struct {
	ID       xid.ID
	RoundId  int32
	SubRound int32
	UserId   string
	Price    string
	NumToken string
	Type     string
}

type ICOUserBought struct {
	Rank     int
	UserId   string
	NumToken string
}

type Tx interface {
	WithTx(ctx context.Context, fn func(ctx context.Context) error) error
}

type LockRepo interface {
	Lock(ctx context.Context, key string) error
	UnLock(ctx context.Context, key string) error
}

type ICORepo interface {
	Tx
	GetRounds(context.Context) ([]*ICORound, error)
	EndRoundByRoundId(ctx context.Context, roundId int32) error
	GetRoundByRoundId(ctx context.Context, roundId int32) (*ICORound, error)
	SaveHistories(ctx context.Context, histories []ICOHistory) error

	GetCurrentSubRound(context.Context) (*ICOSubRound, error)
	CloseSubRound(ctx context.Context, roundId, subRound int32, boughtToken string) (*ICOSubRound, error)
	UpdateSubRoundBought(ctx context.Context, roundId, subRound int32, boughtToken string) error
	GetSubRoundById(ctx context.Context, id string) (*ICOSubRound, error)

	InitData(ctx context.Context, startTime time.Time) error
	GetBuyICOUser(ctx context.Context, limit, offset int) ([]*ICOUserBought, error)
	GetBuyICOTotalUser(ctx context.Context) (int, error)
	// WithTx(ctx context.Context, fn func(ctx context.Context) error) error
}

// Transaction

type Transaction struct {
	ID xid.ID `json:"id,omitempty"`
	// CreatedAt holds the value of the "created_at" field.
	CreatedAt time.Time `json:"created_at,omitempty"`
	// UpdatedAt holds the value of the "updated_at" field.
	UpdatedAt time.Time `json:"updated_at,omitempty"`
	// TransType holds the value of the "trans_type" field.
	TransType string `json:"trans_type,omitempty"`
	// Source holds the value of the "source" field.
	Source string `json:"source,omitempty"`
	// SrcSymbol holds the value of the "src_symbol" field.
	SrcSymbol string `json:"src_symbol,omitempty"`
	// SrcAmount holds the value of the "src_amount" field.
	SrcAmount string `json:"src_amount,omitempty"`
	// Destination holds the value of the "destination" field.
	Destination string `json:"destination,omitempty"`
	// DestSymbol holds the value of the "dest_symbol" field.
	DestSymbol string `json:"dest_symbol,omitempty"`
	// DestAmount holds the value of the "dest_amount" field.
	DestAmount string `json:"dest_amount,omitempty"`
	// Rate holds the value of the "rate" field.
	Rate string `json:"rate,omitempty"`
	// SourceService holds the value of the "source_service" field.
	SourceService string `json:"source_service,omitempty"`
	SourceId      string `json:"source_id,omitempty"`
	// Status holds the value of the "status" field.
	Status string `json:"status,omitempty"`
}

type TransactionRepo interface {
	Tx
	CreateTransaction(ctx context.Context, input *Transaction) (*Transaction, error)
	GetTransactionsByUserId(ctx context.Context, userId, cursor string, limit int32) ([]*Transaction, string, error)
}

// Wallet

type UserWallet struct {
	ID        xid.ID    `json:"id,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
	UserID    string    `json:"user_id,omitempty"`
	Type      string    `json:"type,omitempty"`
	Symbol    string    `json:"symbol,omitempty"`
	Balance   string    `json:"balance,omitempty"`

	IsActive bool `json:"is_active,omitempty"`
}

type UserWalletRepo interface {
	Tx
	GetWalletByUserId(ctx context.Context, userId, symbol, walletType string) ([]*UserWallet, error)
	CreateWallet(ctx context.Context, userId, symbol, walletType string) (*UserWallet, error)
	DecreaseBalance(ctx context.Context, userId, symbol, amount, walletType string) (int, error)
	IncreaseBalance(ctx context.Context, userId, symbol, amount, walletType string) (int, error)
	InitData(ctx context.Context) error
	CalculateBalance(ctx context.Context, userId, amount, symbol string) (bool, error)
}

type CurrencyRate struct {
	Symbol string
	Rate   string
}

type CurrencyRateRepo interface {
	Tx
	GetCurrencyRate(ctx context.Context, symbol string) (*CurrencyRate, error)
	UpdateCurrencyRateICO(ctx context.Context, rate string) error
	InitData(ctx context.Context) error
}

// IcoCoupon is the model entity for the IcoCoupon schema.
type IcoCoupon struct {
	ID xid.ID `json:"id,omitempty"`
	// CreatedAt holds the value of the "created_at" field.
	CreatedAt time.Time `json:"created_at,omitempty"`
	// UpdatedAt holds the value of the "updated_at" field.
	UpdatedAt time.Time `json:"updated_at,omitempty"`
	// UserID holds the value of the "user_id" field.
	UserID string `json:"user_id,omitempty"`
	// Coupon holds the value of the "coupon" field.
	Coupon string `json:"coupon,omitempty"`
	// Reward holds the value of the "reward" field.
	Reward string `json:"reward,omitempty"`
	// Cashback holds the value of the "cashback" field.
	Cashback string `json:"cashback,omitempty"`
}

type IcoCouponRepo interface {
	GetCoupon(ctx context.Context, coupon string) (*IcoCoupon, error)
	AddCoupon(ctx context.Context, icoCoupon *IcoCoupon) error
}
