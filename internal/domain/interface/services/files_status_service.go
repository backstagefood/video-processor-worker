package services

import (
	"github.com/backstagefood/video-processor-worker/internal/domain"
)

type FilesStatusService interface {
	ListFilesByEmail(userEmail string) ([]*domain.File, error)
}
