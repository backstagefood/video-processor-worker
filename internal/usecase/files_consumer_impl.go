package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/IBM/sarama"
	"github.com/backstagefood/video-processor-worker/internal/domain"
	portRepositories "github.com/backstagefood/video-processor-worker/internal/domain/interface/repositories"
	"github.com/backstagefood/video-processor-worker/internal/repositories"
	"github.com/backstagefood/video-processor-worker/pkg/adapter/bucketconfig"
	databaseconnection "github.com/backstagefood/video-processor-worker/pkg/adapter/postgres"
	"github.com/backstagefood/video-processor-worker/utils"
	"github.com/google/uuid"
	"log/slog"
	"mime/multipart"
	"path/filepath"
	"time"
)

func NewFileConsumer(bucketConn *bucketconfig.ApplicationS3Bucket, dbClient *databaseconnection.ApplicationDatabase) sarama.ConsumerGroupHandler {
	usersRepository := repositories.NewUsersRepository(dbClient)
	filesRepository := repositories.NewFilesRepository(dbClient)
	bucketRepository := repositories.NewBucketRepository(bucketConn)
	return &fileConsumer{
		usersRepository:  usersRepository,
		filesRepository:  filesRepository,
		bucketRepository: bucketRepository,
	}
}

type fileConsumer struct {
	usersRepository  portRepositories.UsersRepository
	filesRepository  portRepositories.FilesRepository
	bucketRepository portRepositories.BucketRepository
}

func (f *fileConsumer) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (f *fileConsumer) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (f *fileConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	// Cria um semáforo com capacidade para 2 goroutines
	sem := make(chan struct{}, 2) // processa 2 videos por vez

	for message := range claim.Messages() {
		// Espera até que haja espaço no semáforo
		sem <- struct{}{}

		slog.Info("recebendo nova mensagem",
			slog.String("topic", message.Topic),
			slog.String("key", string(message.Key)),
			slog.String("value", string(message.Value)),
		)

		var filePayload domain.FilePayload
		if err := json.Unmarshal(message.Value, &filePayload); err != nil {
			slog.Error("não foi possível receber a mensagem do topico kafka", slog.String("error", err.Error()))
			session.MarkMessage(message, "")
			<-sem // Libera o slot no semáforo em caso de erro
			continue
		}

		slog.Info("video recebido com sucesso", "filePayload", filePayload)
		session.MarkMessage(message, "")

		// Gravar arquivo
		fileId := f.insertFile(&filePayload)

		if fileId != nil {
			go func(msg *sarama.ConsumerMessage, id *uuid.UUID, payload domain.FilePayload) {
				defer func() {
					<-sem // Libera o slot no semáforo quando terminar
				}()

				f.atualizaStatus(id, &domain.FileProcessingResult{
					FilePath: nil,
					FileSize: nil,
					Status:   2,
					Message:  "em processamento",
				})
				// Sleep para simular processamento demorado (10 segundos)
				time.Sleep(10 * time.Second)
				processingResult := f.processFile(context.Background(), payload.FilePath, payload.UserName)
				f.atualizaStatus(id, processingResult)

			}(message, fileId, filePayload)
		} else {
			<-sem // Libera o slot se fileId for nil
		}
	}

	// Espera todas as goroutines terminarem antes de retornar
	for i := 0; i < cap(sem); i++ {
		sem <- struct{}{}
	}

	return nil
}

func (f *fileConsumer) atualizaStatus(fileId *uuid.UUID, processingResult *domain.FileProcessingResult) {
	err := f.filesRepository.UpdateFileStatus(fileId, processingResult)
	if err != nil {
		slog.Error("não foi possível atualizar o status do arquivo", "error", err, "processingResulta", processingResult)
	}
}

func (f *fileConsumer) insertFile(payload *domain.FilePayload) *uuid.UUID {
	user, err := f.usersRepository.FindUserByEmail(payload.UserName)
	if err != nil {
		slog.Error("não foi possível obter o usuário", "error", err)
		return nil
	}
	slog.Info("usuário encontrado", "user", user)
	fileEntity := &domain.File{UserID: user.ID, VideoFilePath: payload.FilePath, VideoFileSize: payload.FileSize, FileStatus: domain.FileStatus{ID: 1, Status: ""}}
	fileId, err := f.filesRepository.CreateFile(fileEntity)
	if err != nil {
		slog.Error("não foi possível gravar o arquivo na base de dados", "error", err)
		return nil
	}
	slog.Info("id do novo arquivo na base", "fileId", fileId)
	return fileId

}

func (f *fileConsumer) processFile(ctx context.Context, fileFullPath, userEmail string) *domain.FileProcessingResult {
	videoData, _, err := f.bucketRepository.DownloadFile(ctx, fileFullPath)
	if err != nil {
		return domain.NewFileProcessingResultWithError("não foi possível processar o arquivo de video - " + err.Error())
	}
	frames, err := utils.ExtractFrames(videoData, 1.0)
	if err != nil || len(frames) == 0 {
		return domain.NewFileProcessingResultWithError("não foi possível processar o arquivo de video - " + err.Error())
	}
	slog.Info(fmt.Sprintf("📸 extraídos %d frames\n", len(frames)))
	fileName := utils.GetBaseFilename(fileFullPath)
	zipFilename := fmt.Sprintf("frames_%s.zip", fileName)

	// cria arquivo na memoria para guardar no bucket
	arquivoZip, err := utils.CreateImageZipInMemory(frames)
	if err != nil {
		return domain.NewFileProcessingResultWithError("não foi possível criar o arquivo ZIP em memória - " + err.Error())
	}
	// gravar no bucket
	zipFileSize, zipFilePath, err := f.createFile(ctx, arquivoZip, zipFilename, userEmail)
	if err != nil {
		slog.Error("não foi possível gravar o arquivo zip no bucket", "fileName", fileName, "error", err)
		return domain.NewFileProcessingResultWithError("não foi possível criar o arquivo ZIP no bucket - " + err.Error())
	}
	return &domain.FileProcessingResult{FilePath: &zipFilePath, FileSize: &zipFileSize, Status: 3, Message: fmt.Sprintf("%d frames extraídos", len(frames))}
}

func (f *fileConsumer) createFile(ctx context.Context, file multipart.File, fileName, userEmail string) (int64, string, error) {
	slog.Info("fileConsumer - create file", "userEmail", userEmail, "fileName", fileName)
	// junta nome do usuario com caminho
	path := filepath.Join(utils.SanitizeEmailForPath(userEmail), "zip_files")

	// grava no bucket
	fileFullPath, err := f.bucketRepository.CreateFile(ctx, path, fileName, file)
	if err != nil {
		return 0, "", err
	}

	fileSize, err := utils.GetFileSize(file)
	if err != nil {
		return 0, "", err
	}
	slog.Info("arquivo gravado com sucesso", "fileName", fileName, "filesize", fileSize, "fileFullPath", fileFullPath)
	return fileSize, fileFullPath, nil
}
