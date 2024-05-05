package client

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	profileV1 "github.com/indikay/profile-service/api/profile/v1"
	"github.com/indikay/wallet-service/internal/conf"
)

type ProfileClient struct {
	client profileV1.ProfileServiceClient
}

func NewProfileClient(conf *conf.Data) (*ProfileClient, error) {
	con, err := grpcConnection(conf.Profile.Addr, conf.Profile.Timeout.AsDuration())
	if err != nil {
		log.Fatal("Can not connect to Profile Service ", err)
	}
	return &ProfileClient{client: profileV1.NewProfileServiceClient(con)}, nil
}

func (p *ProfileClient) GetProfileByUsers(ctx context.Context, userIds []string) (map[string]profileV1.UserProfile, error) {
	resp, err := p.client.GetListUserProfileInternal(ctx, &profileV1.GetListUserProfileInternalRequest{UserIds: userIds})
	if err != nil {
		return nil, err
	}

	profiles := make(map[string]profileV1.UserProfile)
	if resp.Code == 0 {
		for _, v := range resp.Data {
			profiles[v.Id] = *v
		}
	}

	return profiles, nil
}
