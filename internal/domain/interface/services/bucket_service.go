package services

import (
	"context"
)

type BucketService interface {
	DownloadFile(ctx context.Context, fileWithPath string) ([]byte, map[string]string, error)
}
