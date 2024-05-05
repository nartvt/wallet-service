package service

import (
	"context"
	"errors"

	"github.com/indikay/go-core/middleware/jwt"
	pb "github.com/indikay/wallet-service/api/wallet/v1"
	"github.com/indikay/wallet-service/internal/biz"
	"github.com/indikay/wallet-service/internal/util"
	"google.golang.org/protobuf/types/known/emptypb"
)

type UserWalletService struct {
	walletUC *biz.WalletTransactionUseCase
	pb.UnimplementedUserWalletServiceServer
}

func NewUserWalletService(walletUC *biz.WalletTransactionUseCase) *UserWalletService {
	return &UserWalletService{walletUC: walletUC}
}

func (s *UserWalletService) GetWalletByUserId(ctx context.Context, req *emptypb.Empty) (*pb.UserWalletResponse, error) {
	userId, _ := jwt.GetUserId(ctx)
	if len(userId) == 0 {
		return nil, util.UnAuthorizeError()
	}

	userWallet, err := s.walletUC.GetUserWallet(ctx, userId)
	if err != nil {
		return nil, util.InternalServerError(err)
	}

	data := make([]*pb.UserWallet, len(userWallet))
	for i, v := range userWallet {
		data[i] = &pb.UserWallet{Symbol: pb.SymbolType(pb.SymbolType_value[v.Symbol]), Balance: v.Balance, WalletType: pb.WalletType(pb.WalletType_value[v.Type])}
	}

	return &pb.UserWalletResponse{Data: data}, nil
}
func (s *UserWalletService) GetWalletHistories(ctx context.Context, req *pb.GetWalletHistoryRequest) (*pb.GetWalletHistoryResponse, error) {
	userId, _ := jwt.GetUserId(ctx)
	if len(userId) == 0 {
		return nil, util.UnAuthorizeError()
	}

	transactions, next, err := s.walletUC.GetTransactionsByUserId(ctx, userId, req.Next, req.Limit)
	if err != nil {
		return nil, util.InternalServerError(err)
	}

	data := make([]*pb.Transaction, len(transactions))
	for i, v := range transactions {
		if userId == v.Destination {
			data[i] = &pb.Transaction{Id: v.ID.String(), UserId: v.Destination, Type: v.TransType, Symbol: pb.SymbolType(pb.SymbolType_value[v.DestSymbol]), Amount: v.DestAmount, CreatedAt: int32(v.CreatedAt.Unix()), Status: v.Status, TransInOut: "IN"}
		} else {
			data[i] = &pb.Transaction{Id: v.ID.String(), UserId: v.Source, Type: v.TransType, Symbol: pb.SymbolType(pb.SymbolType_value[v.SrcSymbol]), Amount: v.SrcAmount, CreatedAt: int32(v.CreatedAt.Unix()), Status: v.Status, TransInOut: "OUT"}
		}
	}

	return &pb.GetWalletHistoryResponse{Data: &pb.GetWalletHistoryResponse_Data{Histories: data, Next: next}}, nil
}

func (c *UserWalletService) GetCurrentRateBySymbol(ctx context.Context, req *pb.CurrentRateRequest) (*pb.CurrentRate, error) {
	if req == nil {
		return &pb.CurrentRate{Code: 1, Msg: "input for get current rate by symbol invalid", MsgKey: "FAILED"}, util.BadRequestError(errors.New("input request invalid"))
	}
	rate, err := c.walletUC.GetCurrentRate(ctx, req)
	if err != nil {
		return &pb.CurrentRate{Code: 1, Msg: err.Error(), MsgKey: "FAILED"}, err
	}

	return &pb.CurrentRate{
		Code: 0, Msg: "SUCCESS", MsgKey: "SUCCESS", Data: &pb.CurrentRate_Data{
			Symbol: rate.Symbol, Rate: rate.Rate,
		},
	}, nil
}
