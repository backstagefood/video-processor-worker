package repositories

import (
	"database/sql"
	"errors"
	"log/slog"

	"github.com/backstagefood/video-processor-worker/internal/domain"
	"github.com/backstagefood/video-processor-worker/internal/domain/interface/repositories"
	databaseconnection "github.com/backstagefood/video-processor-worker/pkg/adapter/postgres"
)

type usersRepositoryImpl struct {
	dbClient *sql.DB
}

func NewUsersRepository(db *databaseconnection.ApplicationDatabase) repositories.UsersRepository {
	return &usersRepositoryImpl{
		dbClient: db.Client(),
	}
}

func (v *usersRepositoryImpl) FindUserByEmail(email string) (*domain.User, error) {
	query := `
        select id, name, email, created_at, updated_at 
        from users 
        where email = $1
    `
	var user domain.User
	err := v.dbClient.QueryRow(
		query,
		email,
	).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		slog.Error("usuário não localizado", "email", email, "error", err)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("usuário não localizado para o email " + email)
		}
		return nil, err
	}

	return &user, nil

}
