package main

import (
	"context"
	"fmt"
	"time"

	kgrpc "github.com/go-kratos/kratos/v2/transport/grpc"
	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/indikay/go-core/middleware/jwt"

	icov1 "github.com/indikay/wallet-service/api/ico/v1"
	// walletV1 "github.com/indikay/wallet-service/api/wallet/v1"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	con, _ := kgrpc.DialInsecure(
		context.Background(),
		kgrpc.WithEndpoint("localhost:9000"), // wallet-service.stag.svc.cluster.local:9000
		// kgrpc.WithMiddleware(
		// 	jwt.Client(jwt.WithClaims(func() jwtlib.Claims {
		// 		return jwtlib.RegisteredClaims{Subject: "1eb1a218-62cf-43dc-9d15-f1f0318c7813", ExpiresAt: jwtlib.NewNumericDate(time.Now().Add(time.Hour))}
		// 	})),
		// ),
		kgrpc.WithTimeout(time.Second*60*5),
	)

	// client := v1.NewICOServiceClient(con)
	// resp, _ := client.GetICOInfo(context.Background(), &emptypb.Empty{})
	// fmt.Println(resp)
	// fmt.Println(xid.New().String())
	// fmt.Println(xid.New().String())
	// fmt.Println(xid.FromString("clvftje5da0c73bbcpc1"))
	ctx := context.Background()
	ctx, _ = jwt.ClientGrpcAuth(ctx, jwt.WithClaims(func() jwtlib.Claims {
		return jwtlib.RegisteredClaims{Subject: "f6a02ca7-ace2-48cb-8af1-60392802d68e", ExpiresAt: jwtlib.NewNumericDate(time.Now().Add(time.Hour))}
	}))

	//1eb1a218-62cf-43dc-9d15-f1f0318c7813

	// walletCli := walletV1.NewUserWalletServiceClient(con)
	// resp, err := walletCli.GetWalletByUserId(ctx, &emptypb.Empty{})
	// if err != nil {
	// 	fmt.Println(err)
	// } else {
	// 	fmt.Println(resp)
	// }

	// tranCli := walletV1.NewTransactionServiceClient(con)
	// // resp, err := tranCli.Deposit(ctx, &walletV1.DepositRequest{SourceId: "truyet-test", Symbol: walletV1.SymbolType_VND, Amount: "300000000"})
	// // if err != nil {
	// // 	fmt.Println(err)
	// // } else {
	// // 	fmt.Println(resp)
	// // }

	// refReward, err := tranCli.ReferralReward(ctx, &walletV1.ReferralRewardRequest{SourceId: "truyet-test", Symbol: walletV1.SymbolType_VND, Amount: "10000"})
	// if err != nil {
	// 	fmt.Println(err)
	// } else {
	// 	fmt.Println(refReward)
	// }

	// icoResp, err := tranCli.BuyICO(ctx, &walletV1.BuyICORequest{SourceId: "truyet-test", Symbol: walletV1.SymbolType_VND, Amount: "1000000"})
	// if err != nil {
	// 	fmt.Println(err)
	// } else {
	// 	fmt.Println(icoResp)
	// }

	// sResp, err := tranCli.Subscription(ctx, &walletV1.SubsciptionRequest{SourceId: "truyet-test", Symbol: walletV1.SymbolType_IND, Amount: "100"})
	// if err != nil {
	// 	fmt.Println(err)
	// } else {
	// 	fmt.Println(sResp)
	// }

	// cfResp, err := tranCli.ChargeFee(ctx, &walletV1.ChargeFeeRequest{SourceId: "truyet-test", Symbol: walletV1.SymbolType_IND, Amount: "1"})
	// if err != nil {
	// 	fmt.Println(err)
	// } else {
	// 	fmt.Println(cfResp)
	// }

	// Add coupon
	// icoCli := icov1.NewICOServiceClient(con)
	// resp, err := icoCli.AddICOCoupon(ctx, &icov1.AddICOCouponRequest{UserId: "c85db913-2036-4844-bb4e-0a6dff80bd2a", Coupon: "TRUYETTEST", Reward: "0.08", Cashback: "0.02"})
	// if err != nil {
	// 	fmt.Println(err)
	// } else {
	// 	fmt.Println(resp)
	// }

	// tranCli := walletV1.NewTransactionServiceClient(con)
	// icoResp, err := tranCli.BuyICO(ctx, &walletV1.BuyICORequest{SourceId: "truyet", Symbol: walletV1.SymbolType_VND, Amount: "1000000", Coupon: "truyet"})
	// if err != nil {
	// 	fmt.Println(err)
	// } else {
	// 	fmt.Println(icoResp)
	// }

	icoCli := icov1.NewICOServiceClient(con)
	icoResp, err := icoCli.GetBuyICOUserHistory(ctx, &icov1.GetBuyICOUserHistoryRequest{})
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(icoResp)
	}

}
