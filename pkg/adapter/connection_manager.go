package adapter

import (
	"github.com/backstagefood/video-processor-worker/internal/domain/interface/adapters"
	"github.com/backstagefood/video-processor-worker/pkg/adapter/bucketconfig"
	"github.com/backstagefood/video-processor-worker/pkg/adapter/kafka"
	databaseconnection "github.com/backstagefood/video-processor-worker/pkg/adapter/postgres"
	"log/slog"
	"os"
)

type ConnectionManager interface {
	GetBucketConn() *bucketconfig.ApplicationS3Bucket
	GetDBConn() *databaseconnection.ApplicationDatabase
	GetMessageConsumer() adapters.MessageConsumer
}

type connectionManagerImpl struct {
	bucketConn      *bucketconfig.ApplicationS3Bucket
	dbConn          *databaseconnection.ApplicationDatabase
	messageConsumer adapters.MessageConsumer
}

func NewConnectionManager() ConnectionManager {
	broker := os.Getenv("KAFKA_BROKER")
	groupId := os.Getenv("KAFKA_GROUP_ID")
	topic := os.Getenv("KAFKA_TOPIC")
	consumer, err := kafka.NewConsumer(broker, groupId, topic)
	if err != nil {
		slog.Error("não foi possível criar o consumidor do topico kafka", "error", err)
	}
	return &connectionManagerImpl{
		bucketConn:      bucketconfig.NewBucketConnection(),
		dbConn:          databaseconnection.NewDbConnection(),
		messageConsumer: consumer,
	}
}

func (c *connectionManagerImpl) GetBucketConn() *bucketconfig.ApplicationS3Bucket {
	return c.bucketConn
}
func (c *connectionManagerImpl) GetDBConn() *databaseconnection.ApplicationDatabase {
	return c.dbConn
}

func (c *connectionManagerImpl) GetMessageConsumer() adapters.MessageConsumer {
	return c.messageConsumer
}
