package bucketconfig

import (
	"log"
	"log/slog"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type ApplicationS3BucketInterface interface {
	Client() *s3.S3
	BucketName() string
}

type ApplicationS3Bucket struct {
	s3Client   *s3.S3
	bucketName string
}

func NewBucketConnection() *ApplicationS3Bucket {

	endpoint := os.Getenv("S3_ENDPOINT")
	bucketName := os.Getenv("S3_BUCKET_NAME")
	region := os.Getenv("AWS_REGION") // AWS_REGION=us-east-1
	s3User := os.Getenv("S3_USER")
	s3Pass := os.Getenv("S3_PASS")

	bucketConnection := &aws.Config{
		Credentials:      credentials.NewStaticCredentials(s3User, s3Pass, ""),
		Endpoint:         aws.String(endpoint),
		Region:           aws.String(region),
		DisableSSL:       aws.Bool(true),
		S3ForcePathStyle: aws.Bool(true),
	}

	newSession, err := session.NewSession(bucketConnection)
	if err != nil {
		log.Fatalf("falha ao criar sessao do bucket s3: %v", err)
	}

	slog.Info("bucket s3 conectado com sucesso!", "endpoint", endpoint, "region", region, "bucketName", bucketName)

	return &ApplicationS3Bucket{s3Client: s3.New(newSession), bucketName: bucketName}
}

func (s *ApplicationS3Bucket) Client() *s3.S3 {
	return s.s3Client
}

func (s *ApplicationS3Bucket) BucketName() string {
	return s.bucketName
}
