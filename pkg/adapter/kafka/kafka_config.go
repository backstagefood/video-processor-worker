package kafka

import (
	"context"
	"github.com/IBM/sarama"
	"github.com/backstagefood/video-processor-worker/internal/domain/interface/adapters"
	"log/slog"
)

func newKafkaConfig() *sarama.Config {
	config := sarama.NewConfig()
	config.Version = sarama.V4_0_0_0
	config.Consumer.Return.Errors = true
	config.Producer.Return.Successes = true
	config.Producer.Retry.Max = 5
	config.Producer.RequiredAcks = sarama.WaitForAll
	return config
}

type Consumer struct {
	ConsumerGroup sarama.ConsumerGroup
	Topic         string
}

func NewConsumer(broker string, groupID string, topic string) (adapters.MessageConsumer, error) {
	consumerGroup, err := sarama.NewConsumerGroup([]string{broker}, groupID, newKafkaConfig())
	if err != nil {
		slog.Error("error creating kafka consumer group", slog.String("error", err.Error()))
		return nil, err
	}
	return &Consumer{ConsumerGroup: consumerGroup, Topic: topic}, nil
}

func (kc *Consumer) ConsumeMessages(ctx context.Context, handler sarama.ConsumerGroupHandler) error {
	for {
		err := kc.ConsumerGroup.Consume(ctx, []string{kc.Topic}, handler)
		if err != nil {
			slog.ErrorContext(ctx, "error consuming messages", slog.String("error", err.Error()))
			return err
		}
	}
}
