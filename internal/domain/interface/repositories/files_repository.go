package repositories

import (
	"github.com/backstagefood/video-processor-worker/internal/domain"
	"github.com/google/uuid"
)

type FilesRepository interface {
	CreateFile(file *domain.File) (*uuid.UUID, error)
	ListFilesByEmail(userEmail string) ([]*domain.File, error)
	UpdateFileStatus(id *uuid.UUID, fileProcessingResult *domain.FileProcessingResult) error
}
