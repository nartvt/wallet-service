package client

import (
	"context"
	"time"

	kgrpc "github.com/go-kratos/kratos/v2/transport/grpc"
	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/google/wire"
	"github.com/indikay/go-core/middleware/jwt"
	"google.golang.org/grpc"
)

var ProviderSet = wire.NewSet(NewProfileClient)

func grpcConnection(host string, timeout time.Duration) (*grpc.ClientConn, error) {
	return kgrpc.DialInsecure(
		context.Background(),
		kgrpc.WithEndpoint(host),
		kgrpc.WithTimeout(timeout),
	)
}

func buildAuth(ctx context.Context, userId string) context.Context {
	ctx, _ = jwt.ClientGrpcAuth(ctx, jwt.WithClaims(func() jwtlib.Claims {
		return jwtlib.RegisteredClaims{Subject: userId, Issuer: "SYSTEM", ExpiresAt: jwtlib.NewNumericDate(time.Now().Add(time.Minute))}
	}))

	return ctx
}
