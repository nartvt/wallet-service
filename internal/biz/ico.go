package biz

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/indikay/wallet-service/internal/constant"
	"github.com/shopspring/decimal"
)

type ICOUsecase struct {
	repo             ICORepo
	icoCoupon        IcoCouponRepo
	currencyRateRepo CurrencyRateRepo
	log              *log.Helper
}

func NewICOUseCase(repo ICORepo, icoCoupon IcoCouponRepo, currencyRateRepo CurrencyRateRepo) *ICOUsecase {
	return &ICOUsecase{
		repo:             repo,
		icoCoupon:        icoCoupon,
		currencyRateRepo: currencyRateRepo,
		log:              log.NewHelper(log.DefaultLogger),
	}
}

func (uc *ICOUsecase) GetICORounds(ctx context.Context) ([]*ICORound, error) {
	return uc.repo.GetRounds(ctx)
}

func (uc *ICOUsecase) GetICOCurrentRound(ctx context.Context) (*ICOSubRound, error) {
	return uc.repo.GetCurrentSubRound(ctx)
}

func (uc *ICOUsecase) AddICOCoupon(ctx context.Context, userId, coupon, reward, cashback string) error {
	coupon = strings.ToUpper(coupon)
	return uc.icoCoupon.AddCoupon(ctx, &IcoCoupon{UserID: userId, Coupon: coupon, Reward: reward, Cashback: cashback})
}

func (uc *ICOUsecase) GetICOCoupon(ctx context.Context, coupon string) (*IcoCoupon, error) {
	coupon = strings.ToUpper(coupon)
	couponData, err := uc.icoCoupon.GetCoupon(ctx, coupon)
	if err != nil {
		return nil, err
	}
	return couponData, nil
}

func (uc *ICOUsecase) GetICOUserHistory(ctx context.Context, next string, limit int) ([]*ICOUserBought, int, error) {
	totalUser, err := uc.repo.GetBuyICOTotalUser(ctx)
	var offset = 0
	if offset, err = strconv.Atoi(next); err != nil {
		offset = 0
	}
	if limit <= 0 || limit > 50 {
		limit = 50
	}

	if offset >= totalUser {
		return []*ICOUserBought{}, totalUser, nil
	}

	userBought, err := uc.repo.GetBuyICOUser(ctx, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	return userBought, totalUser, nil
}

func (uc *ICOUsecase) ICOHistories(ctx context.Context, userId, amount, symbol, icoType string) (decimal.Decimal, error) {
	totalToken := decimal.NewFromInt(0)

	histories := []ICOHistory{}
	remaining := decimal.RequireFromString(amount)
	stop := false

	for !stop && !remaining.IsZero() {
		currentRound, err := uc.repo.GetCurrentSubRound(ctx)
		if err != nil {
			uc.log.Error("ICOHistories ", err)
			return totalToken, err
		}
		currencySymbol := fmt.Sprintf("%s_%s", symbol, constant.TokenSymbolIND)
		currency, err := uc.currencyRateRepo.GetCurrencyRate(ctx, currencySymbol)
		if err != nil {
			uc.log.Error("ICOHistories ", err)
			return totalToken, err
		}
		boughtCurrent := decimal.RequireFromString(currentRound.BoughtToken)
		roundToken := decimal.RequireFromString(currentRound.TotalToken)
		roundRemainToken := roundToken.Sub(boughtCurrent)
		rate := decimal.RequireFromString(currency.Rate)
		roundRemainAmount := roundRemainToken.Mul(rate)
		bought := roundRemainAmount
		if roundRemainAmount.GreaterThan(remaining) {
			stop = true
			bought = remaining
		}

		numToken := bought.Div(rate)
		totalToken = totalToken.Add(numToken)
		histories = append(histories, ICOHistory{
			RoundId:  currentRound.RoundId,
			SubRound: currentRound.SubRound,
			UserId:   userId,
			Price:    currentRound.Price,
			NumToken: numToken.String(),
			Type:     icoType,
		})

		if !stop {
			currentRound.BoughtToken = boughtCurrent.Add(numToken).String()
			err = uc.CloseSubRound(ctx, currentRound)
			if err != nil {
				uc.log.Error("ICOHistories ", err)
				return totalToken, err
			}
			currency, err = uc.currencyRateRepo.GetCurrencyRate(ctx, currencySymbol)
			if err != nil {
				uc.log.Error("ICOHistories ", err)
				return totalToken, err
			}
			remaining = remaining.Sub(bought)
		} else {
			err := uc.repo.UpdateSubRoundBought(ctx, currentRound.RoundId, currentRound.SubRound, boughtCurrent.Add(numToken).String())
			if err != nil {
				uc.log.Error("ICOHistories ", err)
				return totalToken, err
			}
		}
	}

	err := uc.repo.SaveHistories(ctx, histories)
	if err != nil {
		uc.log.Error("ICOHistories ", err)
		return totalToken, err
	}
	return totalToken, nil

}

func (uc *ICOUsecase) CloseSubRound(ctx context.Context, currentRound *ICOSubRound) error {
	newSubRound, err := uc.repo.CloseSubRound(ctx, currentRound.RoundId, currentRound.SubRound, currentRound.BoughtToken)
	if err != nil {
		uc.log.Error("CloseSubRound ", err)
		return err
	}
	if newSubRound.RoundId != currentRound.RoundId {
		err = uc.repo.EndRoundByRoundId(ctx, currentRound.RoundId)
		if err != nil {
			uc.log.Error("CloseSubRound ", err)
			return err
		}

		err = uc.UpdateCurrencyRateICO(ctx, currentRound.Price, newSubRound.Price)
		if err != nil {
			uc.log.Error("CloseSubRound ", err)
			return err
		}
	}
	return nil
}

func (uc *ICOUsecase) UpdateCurrencyRateICO(ctx context.Context, oldPrice, newPrice string) error {
	return uc.currencyRateRepo.UpdateCurrencyRateICO(ctx, decimal.RequireFromString(newPrice).Div(decimal.RequireFromString(oldPrice)).String())
}
