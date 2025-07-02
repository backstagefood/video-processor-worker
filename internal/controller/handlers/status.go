package handlers

import (
	portServices "github.com/backstagefood/video-processor-worker/internal/domain/interface/services"
	"github.com/backstagefood/video-processor-worker/internal/repositories"
	"github.com/backstagefood/video-processor-worker/internal/usecase"
	databaseconnection "github.com/backstagefood/video-processor-worker/pkg/adapter/postgres"
	"github.com/gin-gonic/gin"
	"log/slog"
)

type StatusHandler struct {
	filesStatusService portServices.FilesStatusService
}

// @BasePath /v1/status
// PingExample godoc
// @Summary List all files
// @Schemes
// @Description List all files
// @Tags status
// @Produce application/json
// @Success 200 {object} object{files=[]object{filename=string,size=number,statusId=integer,processingResult=object,created_at=string},total=integer} "success response"
// @Failure 500 {object} object{error=string} "generic error response"
// @Router /v1/status [get]
func NewStatusHandler(dbClient *databaseconnection.ApplicationDatabase) *StatusHandler {
	filesRepository := repositories.NewFilesRepository(dbClient)
	return &StatusHandler{
		filesStatusService: usecase.NewFilesStatusService(filesRepository),
	}
}

func (f *StatusHandler) HandleStatus(c *gin.Context) {
	userEmail := c.MustGet("user_email").(string)
	slog.Info("obtem userEmail em handleStatus", "userEmail", userEmail)

	files, err := f.filesStatusService.ListFilesByEmail(userEmail)
	if err != nil {
		slog.Info("não foi possível obter a lista de arquivos", "error", err)
		c.JSON(500, gin.H{"error": "Erro ao listar arquivos"})
		return
	}
	var results []map[string]interface{}
	for _, file := range files {
		results = append(results, map[string]interface{}{
			"filename":         file.GetZipFileName(),
			"size":             file.ZipFileSize,
			"statusId":         file.FileStatus.ID,
			"status":           file.FileStatus.Status,
			"processingResult": file.ProcessingResult,
			"created_at":       file.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	c.JSON(200, gin.H{
		"files": results,
		"total": len(results),
	})
}
