package data

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/indikay/wallet-service/ent"
	"github.com/indikay/wallet-service/ent/icocoupon"
	"github.com/indikay/wallet-service/internal/biz"
)

type icoCouponRepo struct {
	data *Data
	log  *log.Helper
}

func NewIcoCouponRepo(data *Data) biz.IcoCouponRepo {
	return &icoCouponRepo{data: data, log: log.NewHelper(log.DefaultLogger)}
}

// AddCoupon implements biz.IcoCouponRepo.
func (r *icoCouponRepo) AddCoupon(ctx context.Context, icoCoupon *biz.IcoCoupon) error {
	return r.data.GetClient(ctx).IcoCoupon.Create().SetUserID(icoCoupon.UserID).SetCoupon(icoCoupon.Coupon).SetCashback(icoCoupon.Cashback).SetReward(icoCoupon.Reward).OnConflictColumns(icocoupon.FieldCoupon).UpdateNewValues().Exec(ctx)
}

// GetCoupon implements biz.IcoCouponRepo.
func (r *icoCouponRepo) GetCoupon(ctx context.Context, coupon string) (*biz.IcoCoupon, error) {
	ico, err := r.data.GetClient(ctx).IcoCoupon.Query().Where(icocoupon.CouponEqualFold(coupon), icocoupon.DeletedAtIsNil()).First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return &biz.IcoCoupon{ID: ico.ID, UserID: ico.UserID, Coupon: ico.Coupon, Cashback: ico.Cashback, Reward: ico.Reward}, nil
}
