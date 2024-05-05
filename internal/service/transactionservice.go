package service

import (
	"context"

	"github.com/indikay/go-core/middleware/jwt"
	pb "github.com/indikay/wallet-service/api/wallet/v1"
	"github.com/indikay/wallet-service/internal/biz"
	"github.com/indikay/wallet-service/internal/util"
)

type TransactionService struct {
	pb.UnimplementedTransactionServiceServer
	transUC *biz.WalletTransactionUseCase
}

func NewTransactionService(transUC *biz.WalletTransactionUseCase) *TransactionService {
	return &TransactionService{transUC: transUC}
}

func (s *TransactionService) CalcChargeFee(ctx context.Context, req *pb.CalcChargeFeeRequest) (*pb.CalcChargeFeeResponse, error) {
	userId, _ := jwt.GetUserId(ctx)
	if len(userId) == 0 {
		return nil, util.UnAuthorizeError()
	}
	return s.transUC.CalcChargeFee(ctx, userId, req)
}
func (s *TransactionService) ChargeFee(ctx context.Context, req *pb.ChargeFeeRequest) (*pb.ChargeFeeResponse, error) {
	userId, _ := jwt.GetUserId(ctx)
	if len(userId) == 0 {
		return nil, util.UnAuthorizeError()
	}
	amount, err := s.transUC.ChargeFee(ctx, userId, req.Amount, req.Symbol.String(), req.SourceId, req.TypeFee.String())
	if err != nil {
		return &pb.ChargeFeeResponse{Code: 1, Msg: err.Error(), MsgKey: err.Error()}, nil
	}

	return &pb.ChargeFeeResponse{Code: 0, MsgKey: "CHARGE_FEE_SUCCESS", Msg: "CHARGE FEE SUCCESS", Fee: amount}, nil
}
func (s *TransactionService) Deposit(ctx context.Context, req *pb.DepositRequest) (*pb.DepositResponse, error) {
	userId, _ := jwt.GetUserId(ctx)
	if len(userId) == 0 {
		return nil, util.UnAuthorizeError()
	}

	err := s.transUC.DepositICO(ctx, userId, req.Amount, req.Symbol.String(), req.SourceId, req.IcoType)
	if err != nil {
		return &pb.DepositResponse{Code: 1, Msg: err.Error(), MsgKey: err.Error()}, nil
	}
	return &pb.DepositResponse{Code: 0, Msg: "DEPOSIT SUCCESS", MsgKey: "DEPOSIT_SUCCESS"}, nil
}

func (s *TransactionService) BuyICO(ctx context.Context, req *pb.BuyICORequest) (*pb.BuyICOResponse, error) {
	userId, _ := jwt.GetUserId(ctx)
	if len(userId) == 0 {
		return nil, util.UnAuthorizeError()
	}

	err := s.transUC.BuyICO(ctx, userId, req.Amount, req.Symbol.String(), req.SourceId, req.Coupon)
	if err != nil {
		return &pb.BuyICOResponse{Code: 1, Msg: err.Error(), MsgKey: err.Error()}, nil
	}

	return &pb.BuyICOResponse{Code: 0, Msg: "SUCCESS", MsgKey: "BUY ICO SUCCESS"}, nil
}
func (s *TransactionService) Subscription(ctx context.Context, req *pb.SubsciptionRequest) (*pb.SubsciptionResponse, error) {
	userId, _ := jwt.GetUserId(ctx)
	if len(userId) == 0 {
		return nil, util.UnAuthorizeError()
	}

	err := s.transUC.Subscription(ctx, userId, req.Amount, req.Symbol.String(), req.SourceId)
	if err != nil {
		return &pb.SubsciptionResponse{Code: 1, Msg: err.Error(), MsgKey: err.Error()}, nil
	}
	return &pb.SubsciptionResponse{Code: 0, Msg: "SUBSCRIPTION SUCCESS", MsgKey: "SUBSCRIPTION_SUCCESS"}, nil
}

func (s *TransactionService) ReferralReward(ctx context.Context, req *pb.ReferralRewardRequest) (*pb.ReferralRewardResponse, error) {
	userId, _ := jwt.GetUserId(ctx)
	if len(userId) == 0 {
		return nil, util.UnAuthorizeError()
	}

	err := s.transUC.ReferralReward(ctx, userId, req.Amount, req.Symbol.String(), req.SourceId)
	if err != nil {
		return &pb.ReferralRewardResponse{Code: 1, Msg: err.Error(), MsgKey: err.Error()}, nil
	}
	return &pb.ReferralRewardResponse{Code: 0, Msg: "", MsgKey: "REFERRAL_REWARD_SUCCESS"}, nil
}

func (s *TransactionService) MarketingRewardInternal(ctx context.Context, req *pb.MarketingRewardRequest) (*pb.MarketingRewardResponse, error) {
	userID, err := jwt.GetUserId(ctx)
	if err != nil {
		return &pb.MarketingRewardResponse{
			Code:   1,
			Msg:    err.Error(),
			MsgKey: err.Error(),
			Data:   nil,
		}, nil
	}

	if len(userID) == 0 {
		return nil, util.UnAuthorizeError()
	}

	transactionID, err := s.transUC.MarketingRewardInternal(ctx, userID, req.Amount, req.Symbol.String(), req.SourceId)
	if err != nil {
		return &pb.MarketingRewardResponse{
			Code:   1,
			Msg:    err.Error(),
			MsgKey: err.Error(),
			Data:   nil,
		}, nil
	}

	return &pb.MarketingRewardResponse{
		Code:   0,
		Msg:    "created reward transaction successfully",
		MsgKey: "MARKETING_REWARD_SUCCESS",
		Data:   &pb.MarketingRewardResponse_Data{Id: transactionID},
	}, nil
}
