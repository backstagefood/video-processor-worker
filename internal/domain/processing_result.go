package domain

type ProcessingResult struct {
	Success bool   `json:"success"`
	Code    int    `json:"-"`
	Message string `json:"message"`
}
