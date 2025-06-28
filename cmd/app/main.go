package main

import (
	"context"
	"fmt"
	"github.com/backstagefood/video-processor-worker/pkg/adapter"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	routes "github.com/backstagefood/video-processor-worker/internal/controller/router"
	"github.com/backstagefood/video-processor-worker/utils"
)

// @title Video Processor Worker
// @version 1.0
// @description API for video processing.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /
func main() {

	serverPort := utils.GetEnvVarOrDefault("SERVER_PORT", "8080")
	slog.Info(fmt.Sprintf("🎬 servidor iniciado na porta %s", serverPort))
	slog.Info(fmt.Sprintf("📂 acesse: http://localhost:%s\n", serverPort))

	router := routes.NewRouter(adapter.NewConnectionManager())

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", serverPort),
		Handler: router.Handler(),
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("falha ao iniciar servidor: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 21)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("parando servidor")

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("erro ao parar servidor", "err", err)
	}

	<-ctx.Done()
	slog.Info("tempo esgotado - 1 segundo")
	slog.Info("servidor parado com sucesso")
}
