package data

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/indikay/wallet-service/ent"
	"github.com/indikay/wallet-service/ent/ico"
	"github.com/indikay/wallet-service/ent/icoround"
	"github.com/indikay/wallet-service/internal/biz"
	"github.com/indikay/wallet-service/internal/constant"
	"github.com/rs/xid"
	"github.com/shopspring/decimal"
)

type icoRepo struct {
	data *Data
	log  *log.Helper
}

func NewIcoRepo(data *Data) biz.ICORepo {
	return &icoRepo{data: data, log: log.NewHelper(log.DefaultLogger)}
}

func (r *icoRepo) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return r.data.WithTx(ctx, fn)
}

// GetCurrentSubRound implements biz.ICORepo.
func (r *icoRepo) GetCurrentSubRound(ctx context.Context) (*biz.ICOSubRound, error) {
	round, err := r.data.GetClient(ctx).IcoRound.Query().Where(icoround.IsClose(false)).Order(ent.Asc(icoround.FieldRoundID, icoround.FieldSubRound)).First(ctx)
	if err != nil {
		return nil, err
	}

	resp := biz.ICOSubRound{ID: round.ID,
		RoundId:     round.RoundID,
		SubRound:    round.SubRound,
		Price:       round.Price,
		TotalToken:  round.NumToken,
		BoughtToken: round.BoughtToken,
		IsEnded:     round.IsClose,
	}

	if round.EndAt != nil {
		resp.EndAt = *round.EndAt
	}

	return &resp, nil
}

// GetSubRoundById implements biz.ICORepo.
func (r *icoRepo) GetSubRoundById(ctx context.Context, id string) (*biz.ICOSubRound, error) {
	rid, err := xid.FromString(id)
	if err != nil {
		return nil, err
	}
	round, err := r.data.GetClient(ctx).IcoRound.Get(ctx, rid)
	if err != nil {
		return nil, err
	}

	resp := &biz.ICOSubRound{ID: round.ID,
		RoundId:     round.RoundID,
		SubRound:    round.SubRound,
		Price:       round.Price,
		TotalToken:  round.NumToken,
		BoughtToken: round.BoughtToken,
		IsEnded:     round.IsClose}
	if round.EndAt != nil {
		resp.EndAt = *round.EndAt
	}
	return resp, nil
}

func (r *icoRepo) CloseSubRound(ctx context.Context, roundId, subRound int32, boughtToken string) (*biz.ICOSubRound, error) {
	updated, err := r.data.GetClient(ctx).IcoRound.Update().SetIsClose(true).SetBoughtToken(boughtToken).Where(icoround.RoundID(roundId), icoround.SubRound(subRound), icoround.IsClose(false)).Save(ctx)
	if err != nil {
		return nil, err
	}

	if updated == 0 {
		return nil, fmt.Errorf("Round %d has already closed", roundId)
	}

	round, err := r.data.GetClient(ctx).IcoRound.Query().Where(icoround.IsClose(false)).Order(ent.Asc(icoround.FieldRoundID, icoround.FieldSubRound)).First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	now := time.Now()
	lifetime, _ := strconv.Atoi(constant.SUBROUND_LIFETIME)
	round, err = r.data.GetClient(ctx).IcoRound.UpdateOneID(round.ID).SetStartAt(now).SetEndAt(now.Add(time.Duration(lifetime) * time.Minute)).Save(ctx)
	if err != nil {
		return nil, err
	}
	return &biz.ICOSubRound{ID: round.ID, RoundId: round.RoundID, SubRound: round.SubRound, Price: round.Price, TotalToken: round.NumToken, BoughtToken: round.BoughtToken, EndAt: *round.EndAt, IsEnded: round.IsClose}, nil
}

func (r *icoRepo) UpdateSubRoundBought(ctx context.Context, roundId, subRound int32, boughtToken string) error {
	_, err := r.data.GetClient(ctx).IcoRound.Update().SetBoughtToken(boughtToken).Where(icoround.RoundID(roundId), icoround.SubRound(subRound)).Save(ctx)
	return err
}

func (r *icoRepo) SaveHistories(ctx context.Context, histories []biz.ICOHistory) error {
	domainHistories := make([]*ent.IcoHistoryCreate, len(histories))
	for i, v := range histories {
		domainHistories[i] = r.data.GetClient(ctx).IcoHistory.Create().SetRoundID(v.RoundId).SetNumToken(v.NumToken).SetPrice(v.Price).SetSubRound(v.SubRound).SetUserID(v.UserId).SetType(v.Type)
	}
	_, err := r.data.GetClient(ctx).IcoHistory.CreateBulk(domainHistories...).Save(ctx)
	return err
}

func (r *icoRepo) GetRoundByRoundId(ctx context.Context, roundId int32) (*biz.ICORound, error) {
	round, err := r.data.GetClient(ctx).Ico.Query().Where(ico.RoundID(roundId)).First(ctx)
	if err != nil {
		return nil, err
	}

	return &biz.ICORound{ID: round.ID, RoundId: int32(round.RoundID), RoundName: round.RoundName, Price: round.Price, NumToken: round.NumToken, PriceGap: round.PriceGap}, nil
}

func (r *icoRepo) EndRoundByRoundId(ctx context.Context, roundId int32) error {
	_, err := r.data.GetClient(ctx).Ico.Update().SetEndedAt(time.Now()).Where(ico.RoundID(roundId)).Save(ctx)
	return err
}

// GetRounds implements biz.ICORepo.
func (r *icoRepo) GetRounds(ctx context.Context) ([]*biz.ICORound, error) {
	rounds, err := r.data.GetClient(ctx).Ico.Query().Order(ent.Asc(ico.FieldRoundID)).All(ctx)
	if err != nil {
		return nil, err
	}

	var rs = make([]*biz.ICORound, len(rounds))
	for i, round := range rounds {
		rs[i] = &biz.ICORound{ID: round.ID, RoundId: int32(round.RoundID), RoundName: round.RoundName, Price: round.Price, NumToken: round.NumToken, PriceGap: round.PriceGap}
	}
	return rs, nil
}

func (r *icoRepo) InitData(ctx context.Context, startTime time.Time) error {
	return r.WithTx(ctx, func(ctx context.Context) error {
		icos := make([]*ent.IcoCreate, 5)
		startPrice, _ := decimal.NewFromString("0.022")
		numSub := int32(100)
		icoRounds := make([]*ent.IcoRoundCreate, numSub*5)
		icoRoundIdx := 0
		numToken := decimal.NewFromInt(80000000)
		subRoundToken := numToken.Div(decimal.NewFromInt32(numSub)).String()

		for i := 0; i < len(icos); i++ {
			if i > 0 {
				startPrice = startPrice.Add(startPrice.Mul(decimal.NewFromFloat(float64(0.5))))
			}
			icos[i] = r.data.GetClient(ctx).Ico.Create().SetRoundID(int32(i + 1)).SetNumToken(numToken.String()).SetPrice(startPrice.String()).SetNumSub(numSub).SetRoundName(fmt.Sprintf("Round %d", i+1)).SetPriceGap("50%")
			for j := int32(1); j <= numSub; j++ {
				icoRounds[icoRoundIdx] = r.data.GetClient(ctx).IcoRound.Create().SetRoundID(int32(i + 1)).SetBoughtToken("0").SetNumToken(subRoundToken).SetPrice(startPrice.String()).SetSubRound(j).SetIsClose(false)
				icoRoundIdx++
			}
		}

		// r.data.GetClient(ctx).Ico.Delete().Exec(ctx)
		// r.data.GetClient(ctx).IcoRound.Delete().Exec(ctx)
		_, err := r.data.GetClient(ctx).Ico.CreateBulk(icos...).Save(ctx)
		if err != nil {
			return err
		}

		lifetime, _ := strconv.Atoi(constant.SUBROUND_LIFETIME)
		icoRounds[0].SetStartAt(startTime).SetEndAt(startTime.Add(time.Duration(lifetime) * time.Minute))
		_, err = r.data.GetClient(ctx).IcoRound.CreateBulk(icoRounds...).Save(ctx)
		return err
	})

}

func (r *icoRepo) GetBuyICOUser(ctx context.Context, limit, offset int) ([]*biz.ICOUserBought, error) {
	rs, err := r.data.GetClient(ctx).QueryContext(ctx, fmt.Sprintf("SELECT ROW_NUMBER () OVER ( ORDER BY ih.num_token DESC) rank, ih.user_id, ih.num_token FROM (select user_id, SUM(num_token::DECIMAL) as num_token from ico_histories group by user_id) as ih limit %d offset %d", limit, offset))
	if err != nil {
		return nil, err
	}

	var userBought []*biz.ICOUserBought

	for rs.Next() {
		var rank int
		var userId string
		var token string
		err := rs.Scan(&rank, &userId, &token)
		if err != nil {
			return userBought, err
		}
		userBought = append(userBought, &biz.ICOUserBought{Rank: rank, UserId: userId, NumToken: token})
	}
	rs.Close()

	return userBought, nil

}

func (r *icoRepo) GetBuyICOTotalUser(ctx context.Context) (int, error) {
	rs, err := r.data.GetClient(ctx).QueryContext(ctx, "select COUNT(distinct user_id) from ico_histories")
	if err != nil {
		return 0, err
	}

	count := 0
	rs.Next()
	err = rs.Scan(&count)
	if err != nil {
		return count, err
	}
	rs.Close()

	return count, nil

}
