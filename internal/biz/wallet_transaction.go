package biz

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/go-kratos/kratos/v2/log"
	pb "github.com/indikay/wallet-service/api/wallet/v1"
	v1 "github.com/indikay/wallet-service/api/wallet/v1"
	"github.com/indikay/wallet-service/internal/constant"
	"github.com/shopspring/decimal"
)

var (
	SUBSCRIPTION     = "SUBSCRIPTION"
	CHARGE_FEE       = "CHARGE_FEE"
	REFERRAL_REWARD  = "REFERRAL_REWARD"
	MarketingReward  = "MARKETING_REWARD"
	ICO              = "ICO"
	ICO_COMISSION    = "ICO_COMISSION"
	ICO_CASHBACK     = "ICO_CASHBACK"
	DEPOSITE         = "DEPOSITE"
	TRANS_STATUS     = "COMPLETED"
	CURRENCY_SUPPORT = []string{"VND", "USD", "USDT"}
)

type WalletTransactionUseCase struct {
	transRepo        TransactionRepo
	walletRepo       UserWalletRepo
	currencyRateRepo CurrencyRateRepo
	icoRepo          ICORepo
	icoCoupon        IcoCouponRepo
	queue            QueueJob
	lockRepo         LockRepo
	icoUc            *ICOUsecase
	log              *log.Helper
	publisher        TransactionPublisher
}

func NewWalletTransactionUseCase(repo TransactionRepo, walletRepo UserWalletRepo, icoRepo ICORepo, currencyRateRepo CurrencyRateRepo, icoCoupon IcoCouponRepo, queue QueueJob, publisher TransactionPublisher, icoUc *ICOUsecase, lockRepo LockRepo) *WalletTransactionUseCase {
	return &WalletTransactionUseCase{
		transRepo:        repo,
		walletRepo:       walletRepo,
		icoRepo:          icoRepo,
		currencyRateRepo: currencyRateRepo,
		icoCoupon:        icoCoupon,
		queue:            queue,
		publisher:        publisher,
		icoUc:            icoUc,
		lockRepo:         lockRepo,
		log:              log.NewHelper(log.DefaultLogger),
	}
}

func (uc *WalletTransactionUseCase) InitData(ctx context.Context) {
	// err := uc.walletRepo.InitData(ctx)
	// if err != nil {
	// 	panic(err)
	// }
	// err = uc.icoRepo.InitData(ctx, time.Now().AddDate(0, 0, 7))
	// if err != nil {
	// 	panic(err)
	// }
	// err = uc.currencyRateRepo.InitData(ctx)
	// if err != nil {
	// 	panic(err)
	// }

	round, err := uc.icoRepo.GetCurrentSubRound(ctx)
	if err != nil {
		panic(err)
	}
	uc.queue.Enqueue(ctx, &Task{Data: round.ID.String(), Name: fmt.Sprintf("%s%s", QUEUE_PREFIX, ENDROUND_SCHEDULE), ProcessAt: round.EndAt})
}

func (uc *WalletTransactionUseCase) GetUserWallet(ctx context.Context, userId string) ([]*UserWallet, error) {
	resp, err := uc.walletRepo.GetWalletByUserId(ctx, userId, "", "")
	if err != nil {
		uc.log.Error("GetUserWallet ", err)
		return nil, errors.New(constant.ERROR_INTERNAL)
	}

	if len(resp) == 0 {
		w, err := uc.walletRepo.CreateWallet(ctx, userId, constant.TokenSymbolIND, constant.WALLET_TYPE_USER)
		if err != nil {
			uc.log.Error("GetUserWallet ", err)
			return nil, errors.New(constant.ERROR_INTERNAL)
		}
		resp = append(resp, w)
	}

	return resp, nil
}

func (uc *WalletTransactionUseCase) GetUserWalletWithSymbol(ctx context.Context, userId, symbol string) ([]*UserWallet, error) {
	resp, err := uc.walletRepo.GetWalletByUserId(ctx, userId, symbol, "")
	if err != nil {
		uc.log.Error("GetUserWallet ", err)
		return nil, errors.New(constant.ERROR_INTERNAL)
	}

	if len(resp) == 0 {
		w, err := uc.walletRepo.CreateWallet(ctx, userId, symbol, constant.WALLET_TYPE_USER)
		if err != nil {
			uc.log.Error("GetUserWallet ", err)
			return nil, errors.New(constant.ERROR_INTERNAL)
		}
		resp = append(resp, w)
	}

	return resp, nil
}

func (uc *WalletTransactionUseCase) GetTransactionsByUserId(ctx context.Context, userId, nextCursor string, limit int32) ([]*Transaction, string, error) {
	if limit <= 0 {
		limit = constant.DEFAULT_LIMIT
	}
	resp, next, err := uc.transRepo.GetTransactionsByUserId(ctx, userId, nextCursor, limit)
	if err != nil {
		uc.log.Error("GetTransactionsByUserId ", err)
		return nil, "", errors.New(constant.ERROR_INTERNAL)
	}
	return resp, next, nil
}

func (uc *WalletTransactionUseCase) Subscription(ctx context.Context, userId, amount, symbol, sourceId string) error {
	log.Infof("Subscription:Context: %v, userId: %s, amount: %s, symbol: %s, sourceId: %s", ctx != nil, userId, amount, symbol, sourceId)
	err := uc.walletRepo.WithTx(ctx, func(ctx context.Context) error {
		rate := decimal.RequireFromString("1")
		if symbol != constant.TokenSymbolIND {
			currencySymbol := fmt.Sprintf("%s_%s", symbol, constant.TokenSymbolIND)
			currency, err := uc.currencyRateRepo.GetCurrencyRate(ctx, currencySymbol)
			if err != nil {
				uc.log.Errorf("Subscription: %s", err.Error())
				return errors.New(constant.ERROR_INTERNAL)
			}
			rate = decimal.RequireFromString(currency.Rate)
		}

		amountRated := decimal.RequireFromString(amount).Div(rate)
		chargeSymbol := constant.TokenSymbolIND
		rs, err := uc.walletRepo.DecreaseBalance(ctx, userId, chargeSymbol, amountRated.String(), constant.WALLET_TYPE_USER)
		if err != nil {
			uc.log.Errorf("Subscription - DecreaseBalance: %s", err.Error())
			return errors.New(constant.ERROR_INTERNAL)
		}

		if rs == 0 {
			log.Errorf("Subscription - ERROR_BALANCE_NOT_ENOUGH: %s", constant.ERROR_BALANCE_NOT_ENOUGH)
			return errors.New(constant.ERROR_BALANCE_NOT_ENOUGH)
		}

		trans := &Transaction{TransType: SUBSCRIPTION, Source: userId, SrcAmount: amount, SrcSymbol: symbol, Destination: constant.WALLET_SYS_INCOME, DestSymbol: chargeSymbol, DestAmount: amountRated.String(), Rate: rate.String(), Status: TRANS_STATUS, SourceId: sourceId}

		trans, err = uc.CreateTransaction(ctx, trans)
		if err != nil {
			log.Errorf("Subscription - CreateTransaction: %s", err.Error())
			return err
		}
		uc.publisher.Publish(&TransactionMessage{Id: trans.ID.String(), UserId: userId, Amount: amount, Symbol: symbol, TransType: SUBSCRIPTION})
		return nil
	})

	return err
}

func (uc *WalletTransactionUseCase) DepositICO(ctx context.Context, userId, amount, symbol, sourceId, icoType string) error {
	if len(icoType) == 0 {
		icoType = SUBSCRIPTION
	}
	return uc.ICOTransaction(ctx, userId, amount, symbol, sourceId, DEPOSITE, icoType)

}

// func (uc *WalletTransactionUseCase) Deposit(ctx context.Context, userId, amount, symbol, sourceId string) error {
// 	return uc.walletRepo.WithTx(ctx, func(ctx context.Context) error {
// 		resp, err := uc.walletRepo.GetWalletByUserId(ctx, userId, constant.TokenSymbolIND, constant.WALLET_TYPE_USER)
// 		if err != nil {
// 			uc.log.Error("Deposit ", err)
// 			return errors.New(constant.ERROR_INTERNAL)
// 		}

// 		if len(resp) == 0 {
// 			w, err := uc.walletRepo.CreateWallet(ctx, userId, constant.TokenSymbolIND, constant.WALLET_TYPE_USER)
// 			if err != nil {
// 				uc.log.Error("Deposit ", err)
// 				return errors.New(constant.ERROR_BALANCE_NOT_ENOUGH)
// 			}
// 			resp = append(resp, w)
// 		}
// 		rs, err := uc.walletRepo.IncreaseBalance(ctx, userId, symbol, amount, constant.WALLET_TYPE_USER)
// 		if err != nil || rs == 0 {
// 			uc.log.Error("Deposit ", err)
// 			return errors.New(constant.ERROR_INTERNAL)
// 		}

// 		trans := &Transaction{TransType: DEPOSITE, Source: userId, SrcAmount: amount, SrcSymbol: symbol, Destination: constant.WALLET_SYS_INCOME, DestSymbol: symbol, DestAmount: amount, Status: TRANS_STATUS}

// 		_, err = uc.CreateTransaction(ctx, trans)
// 		if err != nil {
// 			uc.log.Error("Deposit ", err)
// 			return errors.New(constant.ERROR_INTERNAL)
// 		}
// 		return nil
// 	})

// }

func (uc *WalletTransactionUseCase) calculateFeeUsdt(amount, symbol, typeFee string) (string, string) {
	if typeFee != v1.UsdtType_USDT_TYPE.String() {
		return amount, symbol
	}
	amoutUsdtFee := decimal.RequireFromString(amount).Mul(usdt_fee_percent)
	currency, err := uc.icoRepo.GetCurrentSubRound(context.Background())
	if err != nil {
		uc.log.Error("get currency rate has an error >>> ", err)
		return amount, symbol
	}
	amountIndChargeFee := amoutUsdtFee.Div(decimal.RequireFromString(currency.Price))
	return amountIndChargeFee.String(), constant.TokenSymbolIND
}

func (uc *WalletTransactionUseCase) ChargeFee(ctx context.Context, userId, amount, symbol, sourceId, typeFee string) (string, error) {
	log.Infof("ChargeFee:Context: %v, userId: %s, amount: %s, symbol: %s, sourceId: %s, typeFee: %s", ctx != nil, userId, amount, symbol, sourceId, typeFee)
	amount, symbol = uc.calculateFeeUsdt(amount, symbol, typeFee)
	err := uc.walletRepo.WithTx(ctx, func(ctx context.Context) error {
		rs, err := uc.walletRepo.DecreaseBalance(ctx, userId, symbol, amount, constant.WALLET_TYPE_USER)
		if err != nil {
			return errors.New(constant.ERROR_INTERNAL)
		}

		if rs == 0 {
			return errors.New(constant.ERROR_BALANCE_NOT_ENOUGH)
		}

		trans := &Transaction{TransType: CHARGE_FEE, Source: userId, SrcAmount: amount, SrcSymbol: symbol, Destination: constant.WALLET_SYS_INCOME, DestSymbol: symbol, DestAmount: amount, Status: TRANS_STATUS, SourceId: sourceId}
		trans, err = uc.CreateTransaction(ctx, trans)
		if err != nil {
			return err
		}

		uc.publisher.Publish(&TransactionMessage{Id: trans.ID.String(), UserId: userId, Amount: amount, Symbol: symbol, TransType: CHARGE_FEE})
		return err
	})

	return amount, err
}

func (uc *WalletTransactionUseCase) BuyICO(ctx context.Context, userId, amount, symbol, sourceId, couponCode string) error {
	err := uc.ICOTransaction(ctx, userId, amount, symbol, sourceId, ICO, ICO)
	if err != nil {
		return err
	}

	couponCode = strings.Trim(couponCode, " ")
	if len(couponCode) == 0 {
		return nil
	}
	coupon, err := uc.icoCoupon.GetCoupon(ctx, couponCode)
	if err != nil {
		return err
	}

	if coupon == nil {
		return nil
	}
	return uc.walletRepo.WithTx(ctx, func(ctx context.Context) error {
		_, err = uc.GetUserWalletOrCreateWithSymbol(ctx, coupon.UserID, symbol, constant.WALLET_TYPE_REWARD)
		if err != nil {
			return err
		}
		ownReward := decimal.RequireFromString(amount).Mul(decimal.RequireFromString(coupon.Reward)).String()
		_, err = uc.walletRepo.IncreaseBalance(ctx, coupon.UserID, symbol, ownReward, constant.WALLET_TYPE_REWARD)
		if err != nil {
			return err
		}
		trans := &Transaction{TransType: ICO_COMISSION, Source: constant.WALLET_SYS_ICO_REWARD, SrcAmount: ownReward, SrcSymbol: symbol, Destination: coupon.UserID, DestSymbol: symbol, DestAmount: ownReward, SourceId: sourceId, Status: TRANS_STATUS}

		_, err = uc.CreateTransaction(ctx, trans)
		if err != nil {
			return err
		}

		_, err = uc.GetUserWalletOrCreateWithSymbol(ctx, userId, symbol, constant.WALLET_TYPE_REWARD)
		if err != nil {
			return err
		}
		cashback := decimal.RequireFromString(amount).Mul(decimal.RequireFromString(coupon.Cashback)).String()
		_, err = uc.walletRepo.IncreaseBalance(ctx, userId, symbol, cashback, constant.WALLET_TYPE_REWARD)
		if err != nil {
			return err
		}
		trans = &Transaction{TransType: ICO_CASHBACK, Source: constant.WALLET_SYS_ICO_REWARD, SrcAmount: cashback, SrcSymbol: symbol, Destination: userId, DestSymbol: symbol, DestAmount: cashback, SourceId: sourceId, Status: TRANS_STATUS}

		_, err = uc.CreateTransaction(ctx, trans)
		return err
	})

}

func (uc *WalletTransactionUseCase) ICOTransaction(ctx context.Context, userId, amount, symbol, sourceId, transType, icoType string) error {
	uc.GetUserWalletOrCreateWithSymbol(ctx, userId, constant.TokenSymbolIND, constant.WALLET_TYPE_USER)
	err := uc.lockRepo.Lock(ctx, constant.ICO_LOCK)
	if err != nil {
		uc.log.Error("ICOTransaction ", err)
		return errors.New(constant.ERROR_LOCK)
	}
	defer func() {
		uc.lockRepo.UnLock(ctx, constant.ICO_LOCK)
	}()
	log.Debugf("ICOTransaction:Context: %v, userId: %s, amount: %s, symbol: %s, sourceId: %s, transType: %s", ctx != nil, userId, amount, symbol, sourceId, transType)
	return uc.walletRepo.WithTx(ctx, func(ctx context.Context) error {
		uc.GetUserWalletWithSymbol(ctx, userId, constant.TokenSymbolIND)
		totalToken, err := uc.icoUc.ICOHistories(ctx, userId, amount, symbol, icoType)
		if err != nil {
			uc.log.Error("ICOTransaction ", err)
			return errors.New(constant.ERROR_INTERNAL)
		}

		log.Debugf("ICOTransaction:TotalToken: %v", totalToken)

		amount = totalToken.String()
		symbol = constant.TokenSymbolIND
		rs, err := uc.walletRepo.DecreaseBalance(ctx, constant.WALLET_ICO, symbol, amount, constant.WALLET_TYPE_SYSTEM)
		if err != nil {
			uc.log.Error("ICOTransaction ", err)
			return errors.New(constant.ERROR_INTERNAL)
		}

		if rs == 0 {
			return errors.New(constant.ERROR_BALANCE_NOT_ENOUGH)
		}

		uc.walletRepo.IncreaseBalance(ctx, userId, symbol, amount, constant.WALLET_TYPE_USER)
		trans := &Transaction{TransType: transType, Source: constant.WALLET_ICO, SrcAmount: amount, SrcSymbol: symbol, Destination: userId, DestSymbol: symbol,
			DestAmount: amount, SourceId: sourceId, Status: TRANS_STATUS}

		_, err = uc.CreateTransaction(ctx, trans)
		if err != nil {
			uc.log.Error("ICOTransaction ", err)
			return errors.New(constant.ERROR_INTERNAL)
		}
		currentRound, _ := uc.icoRepo.GetCurrentSubRound(ctx)
		uc.queue.Enqueue(ctx, &Task{Data: currentRound.ID.String(), Name: fmt.Sprintf("%s%s", QUEUE_PREFIX, ENDROUND_SCHEDULE), ProcessAt: currentRound.EndAt})
		return nil
	})
}

func (uc *WalletTransactionUseCase) CreateTransaction(ctx context.Context, trans *Transaction) (*Transaction, error) {
	transResult, err := uc.transRepo.CreateTransaction(ctx, trans)
	if err != nil {
		uc.log.Error("CreateTransaction ", err)
		return nil, errors.New(constant.ERROR_INTERNAL)
	}
	return transResult, nil
}

func (uc *WalletTransactionUseCase) GetUserWalletOrCreateWithSymbol(ctx context.Context, userId, symbol, walletType string) ([]*UserWallet, error) {
	resp, err := uc.walletRepo.GetWalletByUserId(ctx, userId, symbol, walletType)
	if err != nil {
		uc.log.Error("GetUserWalletOrCreateWithSymbol ", err)
		return nil, errors.New(constant.ERROR_INTERNAL)
	}

	if len(resp) == 0 {
		w, err := uc.walletRepo.CreateWallet(ctx, userId, symbol, walletType)
		if err != nil {
			uc.log.Error("GetUserWalletOrCreateWithSymbol ", err)
			return nil, errors.New(constant.ERROR_INTERNAL)
		}
		resp = append(resp, w)
	}

	return resp, nil
}

func (uc *WalletTransactionUseCase) ReferralReward(ctx context.Context, userId, amount, symbol, sourceId string) error {
	_, err := uc.GetUserWalletOrCreateWithSymbol(ctx, userId, symbol, constant.WALLET_TYPE_REWARD)
	if err != nil {
		return err
	}

	err = uc.walletRepo.WithTx(ctx, func(ctx context.Context) error {
		_, err := uc.walletRepo.IncreaseBalance(ctx, userId, symbol, amount, constant.WALLET_TYPE_REWARD)
		if err != nil {
			return errors.New(constant.ERROR_INTERNAL)
		}

		trans := &Transaction{TransType: REFERRAL_REWARD, Source: constant.WALLET_SYS_REFERRAL_REWARD, SrcAmount: amount, SrcSymbol: symbol, Destination: userId, DestSymbol: symbol, DestAmount: amount, Status: TRANS_STATUS, SourceId: sourceId}

		_, err = uc.CreateTransaction(ctx, trans)
		if err != nil {
			return err
		}

		return err
	})

	return err
}

func (uc *WalletTransactionUseCase) GetCurrentRate(ctx context.Context, req *v1.CurrentRateRequest) (*CurrencyRate, error) {
	currencySymbol := fmt.Sprintf("%s_%s", req.Symbol, constant.TokenSymbolIND)
	uc.log.Infof("Input symbol value: %s, current value symbol for query: %s", req.Symbol, currencySymbol)
	currency, err := uc.currencyRateRepo.GetCurrencyRate(ctx, currencySymbol)
	if err != nil {
		uc.log.Errorf("Get current rate has an error: %s", err.Error())
		return nil, errors.New(err.Error())
	}
	if currency == nil {
		uc.log.Errorf("Get current rate is null with symbol: %s", req.Symbol)
		return nil, fmt.Errorf("get current rate is null with symbol: %s", req.Symbol)
	}
	return &CurrencyRate{Symbol: currencySymbol, Rate: currency.Rate}, nil
}

func (uc *WalletTransactionUseCase) CalcChargeFee(ctx context.Context, userId string, req *pb.CalcChargeFeeRequest) (*pb.CalcChargeFeeResponse, error) {
	if len(req.Symbol) == 0 || len(req.Amount) == 0 {
		log.Warn("CalcChargeFee has an error: symbol or amount is empty")
		return &pb.CalcChargeFeeResponse{Code: 1, Msg: "CALC_CHARGE_FEE_FAILURE", MsgKey: "Symbol or a amount is empty"}, nil
	}
	isEnough, err := uc.walletRepo.CalculateBalance(ctx, userId, req.Amount, req.Symbol)
	if err != nil {
		log.Errorf("CalcChargeFee has an error: %s", err.Error())
		return &pb.CalcChargeFeeResponse{Code: 1, Msg: "CALC_CHARGE_FEE_FAILURE", MsgKey: err.Error()}, nil
	}
	return &pb.CalcChargeFeeResponse{Code: 0, Msg: "CALC_CHARGE_FEE_SUCCESS", MsgKey: "CALC_CHARGE_FEE_SUCCESS", IsEnough: isEnough}, nil
}

// MarketingRewardInternal is used to increase balance for user with symbol and amount
func (uc *WalletTransactionUseCase) MarketingRewardInternal(ctx context.Context, userID, amount, symbol, sourceID string) (string, error) {
	// get user wallet by user id, symbol, type
	// create a new wallet if not exist
	if _, err := uc.GetUserWalletOrCreateWithSymbol(ctx, userID, symbol, constant.WALLET_TYPE_REWARD); err != nil {
		return "", err
	}

	// increase balance and create transaction with the same transaction
	var transactionID string
	if err := uc.walletRepo.WithTx(ctx, func(ctx context.Context) error {
		_, err := uc.walletRepo.IncreaseBalance(ctx, userID, symbol, amount, constant.WALLET_TYPE_REWARD)
		if err != nil {
			return errors.New(constant.ERROR_INTERNAL)
		}

		trans := &Transaction{
			TransType:   MarketingReward,
			Source:      constant.WalletSysMarketingReward,
			SrcAmount:   amount,
			SrcSymbol:   symbol,
			Destination: userID,
			DestSymbol:  symbol,
			DestAmount:  amount,
			Status:      TRANS_STATUS,
			SourceId:    sourceID,
		}

		createdTransaction, err := uc.CreateTransaction(ctx, trans)
		if err != nil {
			return err
		}

		transactionID = createdTransaction.ID.String()

		return nil
	}); err != nil {
		return "", err
	}

	return transactionID, nil
}
