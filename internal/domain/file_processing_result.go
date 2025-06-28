package domain

type FileProcessingResult struct {
	FilePath *string
	FileSize *int64
	Status   int
	Message  string
}

func NewFileProcessingResultWithError(message string) *FileProcessingResult {
	return &FileProcessingResult{
		FilePath: nil,
		FileSize: nil,
		Status:   4,
		Message:  message,
	}
}
