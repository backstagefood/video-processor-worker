package repositories

import (
	"github.com/backstagefood/video-processor-worker/internal/domain"
)

type UsersRepository interface {
	FindUserByEmail(email string) (*domain.User, error)
}
