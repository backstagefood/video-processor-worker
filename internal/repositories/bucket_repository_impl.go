package repositories

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/backstagefood/video-processor-worker/internal/domain/interface/repositories"
	"github.com/backstagefood/video-processor-worker/pkg/adapter/bucketconfig"
	"io"
	"log"
	"mime/multipart"
	"path/filepath"
	"strings"
)

type bucketRepository struct {
	s3Conn     *s3.S3
	bucketName string
}

func NewBucketRepository(s3Conn *bucketconfig.ApplicationS3Bucket) repositories.BucketRepository {
	return &bucketRepository{
		s3Conn:     s3Conn.Client(),
		bucketName: s3Conn.BucketName(),
	}
}

func (v *bucketRepository) CreateFile(ctx context.Context, path string, filename string, file multipart.File) (string, error) {
	// Ensure clean path construction
	key := filepath.Join(path, filename)
	// For S3, we need forward slashes regardless of OS
	key = filepath.ToSlash(key)
	// Remove any leading/trailing slashes that might cause issues
	key = strings.Trim(key, "/")

	log.Printf("bucketRepository - upload file: %s", key)

	// CreateFile directly from the file reader without loading into memory
	_, err := v.s3Conn.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket: aws.String(v.bucketName),
		Key:    aws.String(key),
		Body:   file,
	})

	if err != nil {
		log.Printf("failed to upload file to S3: %v", err)
		return "", fmt.Errorf("failed to upload file to S3: %w", err)
	}

	return key, nil
}

func (v *bucketRepository) DownloadFile(ctx context.Context, fileWithPath string) ([]byte, map[string]string, error) {
	log.Println("downloading file: ", fileWithPath)
	result, err := v.s3Conn.GetObjectWithContext(ctx, &s3.GetObjectInput{
		Bucket: aws.String(v.bucketName),
		Key:    aws.String(fileWithPath),
	})
	if err != nil {
		log.Println("failed to get object from S3: ", err)
		return nil, nil, fmt.Errorf("failed to get object from S3: %w", err)
	}
	defer result.Body.Close()

	data, err := io.ReadAll(result.Body)
	if err != nil {
		log.Println("failed to get object data: ", err)
		return nil, nil, fmt.Errorf("failed to read object data: %w", err)
	}

	metadata := make(map[string]string)
	if result.Metadata != nil {
		for k, vl := range result.Metadata {
			if v != nil {
				metadata[k] = *vl
			} else {
				metadata[k] = ""
			}
		}
	}

	return data, metadata, nil
}
