package messaging

import (
	"encoding/json"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"github.com/indikay/wallet-service/internal/biz"
	"github.com/indikay/wallet-service/internal/conf"
	"github.com/nats-io/nats.go"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewPublisher)

type transPublisher struct {
	natsCli *nats.Conn
	topic   string
	logger  *log.Helper
}

func NewPublisher(c *conf.Data) biz.TransactionPublisher {
	logger := log.NewHelper(log.DefaultLogger)
	nc, err := nats.Connect(c.Nats.NatsHost)
	if err != nil {
		logger.Error("Nats connect err ", err)
		return nil
	}

	return &transPublisher{natsCli: nc, topic: c.Nats.TopicPublishTransaction, logger: logger}
}

// Publish implements biz.TransactionPublisher.
func (t *transPublisher) Publish(msg *biz.TransactionMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		t.logger.Errorf("error marshal message %v", err)
	}
	t.natsCli.Publish(t.topic, data)
}
