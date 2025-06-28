package usecase

import (
	"context"
	portRepositories "github.com/backstagefood/video-processor-worker/internal/domain/interface/repositories"
	portServices "github.com/backstagefood/video-processor-worker/internal/domain/interface/services"
)

type bucketService struct {
	bucketRepository portRepositories.BucketRepository
}

func NewBucketService(bucketRepository portRepositories.BucketRepository) portServices.BucketService {
	return &bucketService{
		bucketRepository: bucketRepository,
	}
}

func (f *bucketService) DownloadFile(ctx context.Context, fileWithPath string) ([]byte, map[string]string, error) {
	return f.bucketRepository.DownloadFile(ctx, fileWithPath)
}
