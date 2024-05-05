package service

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/indikay/go-core/middleware/jwt"
	pb "github.com/indikay/wallet-service/api/ico/v1"
	"github.com/indikay/wallet-service/internal/biz"
	"github.com/indikay/wallet-service/internal/client"
	"github.com/indikay/wallet-service/internal/util"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ICOService struct {
	pb.UnimplementedICOServiceServer
	icoUc         *biz.ICOUsecase
	profileClient *client.ProfileClient
	logger        log.Helper
}

func NewICOService(icoUc *biz.ICOUsecase, profileClient *client.ProfileClient) *ICOService {
	return &ICOService{icoUc: icoUc, profileClient: profileClient, logger: *log.NewHelper(log.DefaultLogger)}
}

func (s *ICOService) GetICOInfo(ctx context.Context, req *emptypb.Empty) (*pb.GetICOInfoResponse, error) {
	data, err := s.icoUc.GetICORounds(ctx)
	if err != nil {
		return nil, util.InternalServerError(err)
	}

	respData := make([]*pb.ICOInfo, len(data))
	for i, v := range data {
		respData[i] = &pb.ICOInfo{RoundId: v.RoundId, RoundName: v.RoundName, Price: v.Price, NumToken: v.NumToken, PriceGap: v.PriceGap}
	}

	return &pb.GetICOInfoResponse{Data: respData}, nil
}

func (s *ICOService) GetCurrentRound(ctx context.Context, req *emptypb.Empty) (*pb.GetCurrentRoundResponse, error) {
	data, err := s.icoUc.GetICOCurrentRound(ctx)
	if err != nil {
		return nil, util.InternalServerError(err)
	}
	return &pb.GetCurrentRoundResponse{Data: &pb.ICORound{RoundId: data.RoundId, SubRound: data.SubRound, Price: data.Price, BoughtToken: data.BoughtToken, TotalToken: data.TotalToken, EndAt: timestamppb.New(data.EndAt)}}, nil
}

func (s *ICOService) AddICOCoupon(ctx context.Context, req *pb.AddICOCouponRequest) (*emptypb.Empty, error) {
	err := s.icoUc.AddICOCoupon(ctx, req.UserId, req.Coupon, req.Reward, req.Cashback)
	if err != nil {
		return nil, util.InternalServerError(err)
	}
	return &emptypb.Empty{}, nil
}

func (s *ICOService) GetCoupon(ctx context.Context, req *pb.GetCouponRequest) (*pb.GetCouponResponse, error) {
	log.Infof("GET COUPON REQUEST: %s", req.Coupon)
	coupon, err := s.icoUc.GetICOCoupon(ctx, req.Coupon)
	if err != nil {
		log.Errorf("GET COUPON HAS AN ERROR: %s", err.Error())
		return nil, util.InternalServerError(err)
	}
	if coupon == nil {
		log.Error("GET COUPON NOT FOUND")
		return &pb.GetCouponResponse{Code: 1, MsgKey: "COUPON_NOT_FOUND"}, nil
	}
	data, _ := json.Marshal(coupon)
	log.Infof("COUPON RESPONSE: %s", string(data))
	return &pb.GetCouponResponse{Code: 0, MsgKey: "SUCCESS", Data: &pb.Coupon{UserId: coupon.UserID, Coupon: coupon.Coupon, Reward: coupon.Reward, CashBack: coupon.Cashback}}, nil
}

func (s *ICOService) GetBuyICOUserHistory(ctx context.Context, req *pb.GetBuyICOUserHistoryRequest) (*pb.GetBuyICOUserHistoryResponse, error) {
	userLogged, _ := jwt.GetUserId(ctx)
	userBought, count, err := s.icoUc.GetICOUserHistory(ctx, req.Next, int(req.Limit))
	if err != nil {
		log.Error("GetBuyICOUserHistory ", err)
		return &pb.GetBuyICOUserHistoryResponse{Code: 1, MsgKey: "UNKNOWN"}, nil
	}
	resp := &pb.GetBuyICOUserHistoryResponse{Total: int32(count)}

	var userId []string
	for _, v := range userBought {
		userId = append(userId, v.UserId)
	}

	profiles, err := s.profileClient.GetProfileByUsers(ctx, userId)
	if err != nil {
		log.Error("GetBuyICOUserHistory ", err)
		return &pb.GetBuyICOUserHistoryResponse{Code: 1, MsgKey: "UNKNOWN"}, nil
	}

	log.Infof("user logged %s = profile %v", userLogged, profiles[userLogged])
	for idx, v := range userBought {
		email := util.MaskEmail(profiles[v.UserId].Email, 2, 2)
		if len(userLogged) > 0 && v.UserId == userLogged {
			email = profiles[v.UserId].Email
		}
		resp.Data = append(resp.Data, &pb.GetBuyICOUserHistoryResponse_Data{NumToken: v.NumToken, Order: int32(v.Rank), Email: email, FullName: profiles[v.UserId].FullName, Stack: "0"})
		if idx == int(req.Limit)-1 {
			resp.Next = strconv.Itoa(v.Rank)
		}
	}

	return resp, nil
}
