package domain

import (
	"github.com/backstagefood/video-processor-worker/utils"
	"time"

	"github.com/google/uuid"
)

type File struct {
	ID               uuid.UUID  `json:"id"`
	UserID           uuid.UUID  `json:"user_id"`
	VideoFilePath    string     `json:"video_file_path"`
	VideoFileSize    int64      `json:"video_file_size,omitempty"`
	ZipFilePath      *string    `json:"zip_file_path,omitempty"`
	ZipFileSize      *int64     `json:"zip_file_size,omitempty"`
	FileStatus       FileStatus `json:"file_status"`
	ProcessingResult *string    `json:"processing_result,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        *time.Time `json:"updated_at,omitempty"`
}

func (f *File) GetVideoFileName() string {
	return utils.GetFileName(f.VideoFilePath)
}

func (f *File) GetZipFileName() string {
	if f.ZipFilePath != nil {
		return utils.GetFileName(*f.ZipFilePath)
	}
	return ""
}
