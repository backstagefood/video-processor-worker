package repositories

import (
	"database/sql"
	"github.com/google/uuid"
	"log/slog"

	"github.com/backstagefood/video-processor-worker/internal/domain"
	"github.com/backstagefood/video-processor-worker/internal/domain/interface/repositories"
	databaseconnection "github.com/backstagefood/video-processor-worker/pkg/adapter/postgres"
)

type filesRepositoryImpl struct {
	dbClient *sql.DB
}

func NewFilesRepository(db *databaseconnection.ApplicationDatabase) repositories.FilesRepository {
	return &filesRepositoryImpl{
		dbClient: db.Client(),
	}
}

func (f *filesRepositoryImpl) CreateFile(file *domain.File) (*uuid.UUID, error) {
	query := `
        INSERT INTO files
        (user_id, video_file_path, video_file_size, status_id)
        VALUES ($1, $2, $3, $4)
        RETURNING id;
    `
	err := f.dbClient.QueryRow(
		query,
		file.UserID,
		file.VideoFilePath,
		file.VideoFileSize,
		file.FileStatus.ID,
	).Scan(&file.ID)

	if err != nil {
		slog.Error("não foi possível criar o arquivo", "error", err)
		return nil, err
	}

	return &file.ID, nil

}

func (f *filesRepositoryImpl) UpdateFileStatus(id *uuid.UUID, fileProcessingResult *domain.FileProcessingResult) error {
	slog.Info("atualiza status de processamento do arquivo", "fileProcessingResult", fileProcessingResult)
	query := `
        UPDATE files
		SET status_id=$2, zip_file_path=$3, zip_file_size=$4, processing_result=$5, updated_at=now()
		WHERE id=$1;
    `

	// Validade UUID fields
	stmt, err := f.dbClient.Prepare(query)
	defer stmt.Close()
	if err != nil {
		return err
	}
	_, err = stmt.Exec(
		id,
		fileProcessingResult.Status,
		fileProcessingResult.FilePath,
		fileProcessingResult.FileSize,
		fileProcessingResult.Message)
	if err != nil {
		return err
	}
	return nil
}

func (f *filesRepositoryImpl) ListFilesByEmail(userEmail string) ([]*domain.File, error) {
	query := `
       SELECT f.id, f.user_id, f.video_file_path, f.video_file_size, f.zip_file_path, f.zip_file_size, s.id, s.status, f.processing_result, f.created_at, f.updated_at
		FROM files f, users u, file_status s
		WHERE f.user_id = u.id
		  AND f.status_id = s.id
		  AND u.email = $1;
	`
	stmt, err := f.dbClient.Prepare(query)
	defer stmt.Close()
	if err != nil {
		return nil, err
	}
	rows, err := stmt.Query(userEmail)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	files := make([]*domain.File, 0)
	for rows.Next() {
		var file domain.File
		if err := rows.Scan(
			&file.ID,
			&file.UserID,
			&file.VideoFilePath,
			&file.VideoFileSize,
			&file.ZipFilePath,
			&file.ZipFileSize,
			&file.FileStatus.ID,
			&file.FileStatus.Status,
			&file.ProcessingResult,
			&file.CreatedAt,
			&file.UpdatedAt,
		); err != nil {
			return nil, err
		}
		files = append(files, &file)
	}
	return files, nil
}
