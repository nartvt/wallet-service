package migratedata

import (
	"context"
	"fmt"
	"time"

	"ariga.io/atlas/sql/migrate"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/schema"
	"github.com/indikay/wallet-service/ent"
	"github.com/indikay/wallet-service/ent/icoround"
	"github.com/indikay/wallet-service/ent/userwallet"
	"github.com/indikay/wallet-service/internal/constant"
	"github.com/shopspring/decimal"
)

func ApplyTokenomic(next schema.Applier) schema.Applier {
	return schema.ApplyFunc(func(ctx context.Context, conn dialect.ExecQuerier, plan *migrate.Plan) error {
		client := ent.NewClient(
			ent.Driver(sql.NewDriver(dialect.Postgres, sql.Conn{ExecQuerier: conn.(*sql.Tx)})),
		)

		count, err := client.Ico.Query().Count(ctx)
		if err != nil {
			return err
		}
		if count != 3 {
			err := initICOData(ctx, client)
			if err != nil {
				return err
			}
		}

		count, err = client.UserWallet.Query().Where(userwallet.UserID(constant.WALLET_SYS_TOKEN)).Count(ctx)
		if err != nil {
			return err
		}
		if count == 0 {
			err = initWalletData(ctx, client)
			if err != nil {
				return err
			}
		}
		return next.Apply(ctx, conn, plan)
	})
}

// InitICO tokenomic
func initICOData(ctx context.Context, client *ent.Client) error {
	icos := make([]*ent.IcoCreate, 3)
	startPrice, _ := decimal.NewFromString("0.022")
	numSub := int32(100)
	subRound := int32(17)
	var icoRounds []*ent.IcoRoundCreate
	icoRoundIdx := 0
	numToken := decimal.NewFromInt(30000000)
	subRoundToken := numToken.Div(decimal.NewFromInt32(numSub)).String()
	startTime := time.Now()

	for i := 0; i < len(icos); i++ {
		if i > 0 {
			startPrice = startPrice.Add(startPrice.Mul(decimal.NewFromFloat(float64(1.0))))
		}
		icos[i] = client.Ico.Create().SetRoundID(int32(i + 1)).SetNumToken(numToken.String()).SetPrice(startPrice.String()).SetNumSub(numSub).SetRoundName(fmt.Sprintf("Round %d", i+1)).SetPriceGap("100%").SetCreatedAt(time.Now())
		for j := int32(1); j <= numSub; j++ {
			if i == 0 && j < subRound {
				continue
			}
			icoRounds = append(icoRounds, client.IcoRound.Create().SetRoundID(int32(i+1)).SetBoughtToken("0").SetNumToken(subRoundToken).SetPrice(startPrice.String()).SetSubRound(j).SetIsClose(false))
			icoRoundIdx++
		}
	}

	client.Ico.Delete().Exec(ctx)
	err := client.Ico.CreateBulk(icos...).Exec(ctx)
	if err != nil {
		return err
	}

	_, err = client.IcoRound.Delete().Where(icoround.And(icoround.RoundID(1), icoround.SubRoundGTE(subRound))).Exec(ctx)
	if err != nil {
		return err
	}
	_, err = client.IcoRound.Delete().Where(icoround.RoundIDGT(1)).Exec(ctx)
	if err != nil {
		return err
	}

	err = client.IcoRound.CreateBulk(icoRounds...).Exec(ctx)
	if err != nil {
		return err
	}
	err = client.IcoRound.Update().Where(icoround.And(icoround.RoundID(1), icoround.SubRoundLT(subRound)), icoround.CreatedAtGT(startTime)).Exec(ctx)
	return err
}

// InitICO tokenomic
func initWalletData(ctx context.Context, client *ent.Client) error {

	// Add transaction ICO wallet
	err := client.Transaction.Create().SetTransType(constant.TRANS_INTERNAL).
		SetSource(constant.WALLET_SYS_TOKEN).SetSrcSymbol(constant.TokenSymbolIND).SetSrcAmount("400000000").
		SetDestination(constant.WALLET_ICO).SetDestSymbol(constant.TokenSymbolIND).SetDestAmount("400000000").
		SetRate("").SetSourceService("").SetSourceID("").SetStatus("COMPLETED").Exec(ctx)
	if err != nil {
		return err
	}

	// Move wallet SYS_ICO_BACKUP  to ICO 8.000.000
	amount := "8000000"
	err = client.Transaction.Create().SetTransType(constant.TRANS_INTERNAL).
		SetSource(constant.WALLET_SYS_ICO_BACKUP).SetSrcSymbol(constant.TokenSymbolIND).SetSrcAmount(amount).
		SetDestination(constant.WALLET_ICO).SetDestSymbol(constant.TokenSymbolIND).SetDestAmount(amount).
		SetRate("").SetSourceService("").SetSourceID("").SetStatus("COMPLETED").Exec(ctx)
	if err != nil {
		return err
	}

	err = client.UserWallet.Update().Modify(func(u *sql.UpdateBuilder) {
		u.Set(userwallet.FieldBalance, sql.ExprFunc(func(b *sql.Builder) {
			b.Ident(userwallet.FieldBalance).WriteOp(sql.OpSub).Arg(amount)
		}))
	}).Where(userwallet.BalanceGTE(amount), userwallet.UserID(constant.WALLET_SYS_ICO_BACKUP), userwallet.Symbol(constant.TokenSymbolIND), userwallet.TypeEQ(constant.WALLET_TYPE_SYSTEM)).Exec(ctx)

	err = client.UserWallet.Update().Modify(func(u *sql.UpdateBuilder) {
		u.Set(userwallet.FieldBalance, sql.ExprFunc(func(b *sql.Builder) {
			b.Ident(userwallet.FieldBalance).WriteOp(sql.OpAdd).Arg(amount)
		}))
	}).Where(userwallet.BalanceGTE(amount), userwallet.UserID(constant.WALLET_ICO), userwallet.Symbol(constant.TokenSymbolIND), userwallet.TypeEQ(constant.WALLET_TYPE_SYSTEM)).Exec(ctx)

	// Make transaction ICO -> TOKENOMIC
	subAmount := "310000000"
	err = client.Transaction.Create().SetTransType(constant.TRANS_INTERNAL).
		SetSource(constant.WALLET_ICO).SetSrcSymbol(constant.TokenSymbolIND).SetSrcAmount(subAmount).
		SetDestination(constant.WALLET_SYS_TOKEN).SetDestSymbol(constant.TokenSymbolIND).SetDestAmount(subAmount).
		SetRate("").SetSourceService("").SetSourceID("").SetStatus("COMPLETED").Exec(ctx)
	if err != nil {
		return err
	}

	err = client.UserWallet.Update().Modify(func(u *sql.UpdateBuilder) {
		u.Set(userwallet.FieldBalance, sql.ExprFunc(func(b *sql.Builder) {
			b.Ident(userwallet.FieldBalance).WriteOp(sql.OpSub).Arg(subAmount)
		}))
	}).Where(userwallet.BalanceGTE(subAmount), userwallet.UserID(constant.WALLET_ICO), userwallet.Symbol(constant.TokenSymbolIND), userwallet.TypeEQ(constant.WALLET_TYPE_SYSTEM)).Exec(ctx)

	// Setup team wallet
	err = client.Transaction.Create().SetTransType(constant.TRANS_INTERNAL).
		SetSource(constant.WALLET_SYS_TOKEN).SetSrcSymbol(constant.TokenSymbolIND).SetSrcAmount("300000000").
		SetDestination(constant.WALLET_SYS_TEAM).SetDestSymbol(constant.TokenSymbolIND).SetDestAmount("300000000").
		SetRate("").SetSourceService("").SetSourceID("").SetStatus("COMPLETED").Exec(ctx)
	if err != nil {
		return err
	}

	err = client.UserWallet.Create().SetUserID(constant.WALLET_SYS_TEAM).SetSymbol(constant.TokenSymbolIND).SetBalance("300000000").SetType(constant.WALLET_TYPE_SYSTEM).SetIsActive(true).OnConflictColumns(userwallet.FieldUserID, userwallet.FieldSymbol, userwallet.FieldType).DoNothing().Exec(ctx)
	if err != nil && err.Error() != "sql: no rows in result set" {
		return err
	}

	// Setup team marketing
	err = client.Transaction.Create().SetTransType(constant.TRANS_INTERNAL).
		SetSource(constant.WALLET_SYS_TOKEN).SetSrcSymbol(constant.TokenSymbolIND).SetSrcAmount("100000000").
		SetDestination(constant.WALLET_SYS_MARKETING).SetDestSymbol(constant.TokenSymbolIND).SetDestAmount("100000000").
		SetRate("").SetSourceService("").SetSourceID("").SetStatus("COMPLETED").Exec(ctx)
	if err != nil {
		return err
	}

	err = client.UserWallet.Create().SetUserID(constant.WALLET_SYS_MARKETING).SetSymbol(constant.TokenSymbolIND).SetBalance("100000000").SetType(constant.WALLET_TYPE_SYSTEM).SetIsActive(true).OnConflictColumns(userwallet.FieldUserID, userwallet.FieldSymbol, userwallet.FieldType).DoNothing().Exec(ctx)
	if err != nil && err.Error() != "sql: no rows in result set" {
		return err
	}

	// Setup team liquidity
	err = client.Transaction.Create().SetTransType(constant.TRANS_INTERNAL).
		SetSource(constant.WALLET_SYS_TOKEN).SetSrcSymbol(constant.TokenSymbolIND).SetSrcAmount("150000000").
		SetDestination(constant.WALLET_SYS_LIQUIDITY).SetDestSymbol(constant.TokenSymbolIND).SetDestAmount("150000000").
		SetRate("").SetSourceService("").SetSourceID("").SetStatus("COMPLETED").Exec(ctx)
	if err != nil {
		return err
	}

	err = client.UserWallet.Create().SetUserID(constant.WALLET_SYS_LIQUIDITY).SetSymbol(constant.TokenSymbolIND).SetBalance("150000000").SetType(constant.WALLET_TYPE_SYSTEM).SetIsActive(true).OnConflictColumns(userwallet.FieldUserID, userwallet.FieldSymbol, userwallet.FieldType).DoNothing().Exec(ctx)
	if err != nil && err.Error() != "sql: no rows in result set" {
		return err
	}

	// Setup team reserve
	err = client.Transaction.Create().SetTransType(constant.TRANS_INTERNAL).
		SetSource(constant.WALLET_SYS_TOKEN).SetSrcSymbol(constant.TokenSymbolIND).SetSrcAmount("15000000").
		SetDestination(constant.WALLET_SYS_RESERVE).SetDestSymbol(constant.TokenSymbolIND).SetDestAmount("15000000").
		SetRate("").SetSourceService("").SetSourceID("").SetStatus("COMPLETED").Exec(ctx)
	if err != nil {
		return err
	}

	err = client.UserWallet.Create().SetUserID(constant.WALLET_SYS_RESERVE).SetSymbol(constant.TokenSymbolIND).SetBalance("15000000").SetType(constant.WALLET_TYPE_SYSTEM).SetIsActive(true).OnConflictColumns(userwallet.FieldUserID, userwallet.FieldSymbol, userwallet.FieldType).DoNothing().Exec(ctx)
	if err != nil && err.Error() != "sql: no rows in result set" {
		return err
	}

	// Setup team advisor
	err = client.Transaction.Create().SetTransType(constant.TRANS_INTERNAL).
		SetSource(constant.WALLET_SYS_TOKEN).SetSrcSymbol(constant.TokenSymbolIND).SetSrcAmount("10000000").
		SetDestination(constant.WALLET_SYS_ADVISOR).SetDestSymbol(constant.TokenSymbolIND).SetDestAmount("10000000").
		SetRate("").SetSourceService("").SetSourceID("").SetStatus("COMPLETED").Exec(ctx)
	if err != nil {
		return err
	}

	err = client.UserWallet.Create().SetUserID(constant.WALLET_SYS_ADVISOR).SetSymbol(constant.TokenSymbolIND).SetBalance("10000000").SetType(constant.WALLET_TYPE_SYSTEM).SetIsActive(true).OnConflictColumns(userwallet.FieldUserID, userwallet.FieldSymbol, userwallet.FieldType).DoNothing().Exec(ctx)
	if err != nil && err.Error() != "sql: no rows in result set" {
		return err
	}

	// Setup team partner
	err = client.Transaction.Create().SetTransType(constant.TRANS_INTERNAL).
		SetSource(constant.WALLET_SYS_TOKEN).SetSrcSymbol(constant.TokenSymbolIND).SetSrcAmount("135000000").
		SetDestination(constant.WALLET_SYS_PARTNER).SetDestSymbol(constant.TokenSymbolIND).SetDestAmount("135000000").
		SetRate("").SetSourceService("").SetSourceID("").SetStatus("COMPLETED").Exec(ctx)
	if err != nil {
		return err
	}

	err = client.UserWallet.Create().SetUserID(constant.WALLET_SYS_PARTNER).SetSymbol(constant.TokenSymbolIND).SetBalance("135000000").SetType(constant.WALLET_TYPE_SYSTEM).SetIsActive(true).OnConflictColumns(userwallet.FieldUserID, userwallet.FieldSymbol, userwallet.FieldType).DoNothing().Exec(ctx)
	if err != nil && err.Error() != "sql: no rows in result set" {
		return err
	}

	// Setup team ecosystem fund
	err = client.Transaction.Create().SetTransType(constant.TRANS_INTERNAL).
		SetSource(constant.WALLET_SYS_TOKEN).SetSrcSymbol(constant.TokenSymbolIND).SetSrcAmount("200000000").
		SetDestination(constant.WALLET_SYS_ECOFUND).SetDestSymbol(constant.TokenSymbolIND).SetDestAmount("200000000").
		SetRate("").SetSourceService("").SetSourceID("").SetStatus("COMPLETED").Exec(ctx)
	if err != nil {
		return err
	}

	err = client.UserWallet.Create().SetUserID(constant.WALLET_SYS_ECOFUND).SetSymbol(constant.TokenSymbolIND).SetBalance("200000000").SetType(constant.WALLET_TYPE_SYSTEM).SetIsActive(true).OnConflictColumns(userwallet.FieldUserID, userwallet.FieldSymbol, userwallet.FieldType).DoNothing().Exec(ctx)
	if err != nil && err.Error() != "sql: no rows in result set" {
		return err
	}

	// Setup wallet tokenomic
	err = client.UserWallet.Create().SetUserID(constant.WALLET_SYS_TOKEN).SetSymbol(constant.TokenSymbolIND).SetBalance("0").SetType(constant.WALLET_TYPE_SYSTEM).SetIsActive(true).OnConflictColumns(userwallet.FieldUserID, userwallet.FieldSymbol, userwallet.FieldType).DoNothing().Exec(ctx)
	if err != nil && err.Error() != "sql: no rows in result set" {
		return err
	}

	return nil
}
