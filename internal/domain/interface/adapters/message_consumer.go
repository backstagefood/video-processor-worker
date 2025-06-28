package adapters

import (
	"context"
	"github.com/IBM/sarama"
)

type MessageConsumer interface {
	ConsumeMessages(ctx context.Context, handler sarama.ConsumerGroupHandler) error
}
