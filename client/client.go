package client

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	grpclib "github.com/go-kratos/kratos/v2/transport/grpc"
	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/indikay/go-core/middleware/jwt"
	ico "github.com/indikay/wallet-service/api/ico/v1"
	v1 "github.com/indikay/wallet-service/api/wallet/v1"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	WalletClient *walletClient
	url          string
	conn         *grpc.ClientConn
)

const TimeOut = 120 * time.Second
const maxMsgSize = 1024 * 1024 * 1024 //1GB

type walletClient struct {
	transactionClient v1.TransactionServiceClient
	clientWallet      v1.UserWalletServiceClient
	icoClient         ico.ICOServiceClient
}

func InitEnvironment(walletUrl string) {
	url = walletUrl
	initWithAddressUrl(walletUrl)
}
func registerClaim(subject string) jwtlib.RegisteredClaims {
	return jwtlib.RegisteredClaims{
		Subject:   subject,
		ExpiresAt: jwtlib.NewNumericDate(time.Now().Add(TimeOut)),
	}
}

func ContextWithTimeOut() context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), TimeOut)
	go cancelContext(cancel)
	return ctx
}

func cancelContext(cancel context.CancelFunc) {
	time.Sleep(TimeOut)
	cancel()
	log.Info("Cancel context successfully !")
}

func optionRegisterClaim(userId string) jwt.Option {
	return jwt.WithClaims(func() jwtlib.Claims {
		return registerClaim(userId)
	})
}

func middlewareWithJwt(userId string) middleware.Middleware {
	return jwt.Client(optionRegisterClaim(userId))
}

func clientOptionMiddlewareWithJwt(userId string) grpclib.ClientOption {
	return grpclib.WithMiddleware(
		middlewareWithJwt(userId),
	)
}

func userIntoContext(userId string) context.Context {
	ctx, err := jwt.ClientGrpcAuth(ContextWithTimeOut(), optionRegisterClaim(userId))
	if err != nil {
		log.Errorf("middlewareWithJwt: %s", err.Error())
		panic(err)
	}
	return ctx
}

func connectGrpc(url string) (*grpc.ClientConn, error) {
	return grpclib.DialInsecure(
		context.TODO(),
		grpclib.WithEndpoint(url),
	)
}

func initWithAddressUrl(url string) {
	var err error
	conn, err = connectGrpc(url)
	if err != nil {
		log.Errorf("CONNECT WALLET HAS AN ERROR: %s", err.Error())
		panic(err)
	}

	WalletClient = &walletClient{
		transactionClient: v1.NewTransactionServiceClient(conn),
		clientWallet:      v1.NewUserWalletServiceClient(conn),
		icoClient:         ico.NewICOServiceClient(conn),
	}
}

func CloseConnect() {
	if conn == nil {
		return
	}
	if err := conn.Close(); err != nil {
		log.Errorf("Close connection has an error: %s", err.Error())
		return
	}
	WalletClient = nil
	log.Info("Close connection successfully !")
}

func GrpcWalletClient() *walletClient {
	if WalletClient == nil {
		initWithAddressUrl(url)
	}
	return WalletClient
}

func (r *walletClient) GetUserBalance(userId string) (*v1.UserWalletResponse, error) {
	return r.clientWallet.GetWalletByUserId(userIntoContext(userId), &emptypb.Empty{})
}

func (r *walletClient) BuyIco(userId, symbol, amount, coupon string) (*v1.BuyICOResponse, error) {
	return GrpcWalletClient().transactionClient.BuyICO(userIntoContext(userId), &v1.BuyICORequest{
		Symbol: v1.SymbolType(v1.SymbolType_value[symbol]),
		Amount: amount,
		Coupon: coupon,
	})
}

func (r *walletClient) GetCoupon(userId, coupon string) (*ico.GetCouponResponse, error) {
	return r.icoClient.GetCoupon(userIntoContext(userId), &ico.GetCouponRequest{Coupon: coupon})
}

func (r *walletClient) GetCurrentRateBySymbol(userId, symbol string) (*v1.CurrentRate, error) {
	return r.clientWallet.GetCurrentRateBySymbol(userIntoContext(userId), &v1.CurrentRateRequest{Symbol: symbol})
}

func (r *walletClient) Deposit(userId, symbol, actionType, sourceId string, amount float64) (*v1.DepositResponse, error) {
	rq := &v1.DepositRequest{
		Symbol:   v1.SymbolType(v1.SymbolType_value[symbol]),
		Amount:   fmt.Sprintf("%f", amount),
		SourceId: sourceId,
		IcoType:  actionType,
	}
	return r.transactionClient.Deposit(userIntoContext(userId), rq)
}

func (r *walletClient) ChargeFee(userId string, req *v1.ChargeFeeRequest) (*v1.ChargeFeeResponse, error) {
	return r.transactionClient.ChargeFee(userIntoContext(userId), req)
}

func (r *walletClient) ChargeSubscription(userId, amount, symbol, sourceId string) (*v1.SubsciptionResponse, error) {
	req := v1.SubsciptionRequest{
		Symbol:   v1.SymbolType(v1.SymbolType_value[symbol]),
		Amount:   amount,
		SourceId: sourceId,
	}
	return r.transactionClient.Subscription(userIntoContext(userId), &req)
}

func (r *walletClient) SubmitCommissionReward(userId string, input *v1.ReferralRewardRequest) (*v1.ReferralRewardResponse, error) {
	return r.transactionClient.ReferralReward(userIntoContext(userId), input)
}
