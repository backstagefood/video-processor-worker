package usecase

import (
	"github.com/backstagefood/video-processor-worker/internal/domain"
	portRepositories "github.com/backstagefood/video-processor-worker/internal/domain/interface/repositories"
	portServices "github.com/backstagefood/video-processor-worker/internal/domain/interface/services"
)

type fileStatusService struct {
	filesRepository portRepositories.FilesRepository
}

func NewFilesStatusService(filesRepository portRepositories.FilesRepository) portServices.FilesStatusService {
	return &fileStatusService{
		filesRepository: filesRepository,
	}
}

func (f fileStatusService) ListFilesByEmail(userEmail string) ([]*domain.File, error) {
	return f.filesRepository.ListFilesByEmail(userEmail)
}
