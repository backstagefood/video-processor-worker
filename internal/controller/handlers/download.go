package handlers

import (
	portServices "github.com/backstagefood/video-processor-worker/internal/domain/interface/services"
	"github.com/backstagefood/video-processor-worker/internal/repositories"
	"github.com/backstagefood/video-processor-worker/internal/usecase"
	"github.com/backstagefood/video-processor-worker/pkg/adapter/bucketconfig"
	"github.com/backstagefood/video-processor-worker/utils"
	"github.com/gin-gonic/gin"
	"log/slog"
	"net/http"
	"path/filepath"
)

type DownloadHandler struct {
	bucketService portServices.BucketService
}

func NewDownloadHandler(s3Conn *bucketconfig.ApplicationS3Bucket) *DownloadHandler {
	bucketRepository := repositories.NewBucketRepository(s3Conn)
	return &DownloadHandler{
		bucketService: usecase.NewBucketService(bucketRepository),
	}
}

// @BasePath /v1/download/:filename
// PingExample godoc
// @Summary Download zip file
// @Schemes
// @Description Download zip file with screenshots of the video
// @Tags download
// @Produce application/zip
// @Param filename path string true "Filename"
// @Success 200 {file} file "ZIP file"
// @Failure 500 {object} object{error=string} "generic error response"
// @Router /v1/download/{filename} [get]
func (h *DownloadHandler) HandleDownload(c *gin.Context) {
	userEmail := c.MustGet("user_email").(string)
	slog.Info("obtem userEmail em handleDownload", "userEmail", userEmail)

	filename := c.Param("filename")
	filePath := filepath.Join(utils.SanitizeEmailForPath(userEmail), "zip_files", filename)
	file, _, err := h.bucketService.DownloadFile(c, filePath)
	if err != nil {
		c.JSON(404, gin.H{"error": "Arquivo não encontrado"})
		return
	}

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "application/zip")
	c.Data(http.StatusOK, "application/octet-stream", file)
}
