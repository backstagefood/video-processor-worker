package domain

type FilePayload struct {
	UserName string `json:"user_name"`
	FilePath string `json:"file_path"`
	FileSize int64  `json:"file_size"`
}
