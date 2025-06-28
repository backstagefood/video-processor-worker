package repositories

import (
	"context"
	"mime/multipart"
)

type BucketRepository interface {
	DownloadFile(ctx context.Context, fileWithPath string) ([]byte, map[string]string, error)
	CreateFile(ctx context.Context, path string, filename string, file multipart.File) (string, error)
}
