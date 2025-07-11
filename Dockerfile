FROM golang:1.24.4-alpine AS base

ARG VERSION
ARG PROJECT_NAME

WORKDIR /app

COPY . .

RUN go mod tidy

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-X 'github.com/backstagefood/video-processor-worker/internal/controller/handlers.Version=${VERSION}' -X 'github.com/backstagefood/video-processor-worker/internal/controller/handlers.ProjectName=${PROJECT_NAME}'" -o video-processor-worker ./cmd/app/.

FROM alpine

ENV GIN_MODE=release

# Instalar ffmpeg
#RUN apk add --no-cache ffmpeg
RUN apk add --no-cache ffmpeg-libs ffmpeg


# Criar diretório de trabalho
WORKDIR /app

COPY --from=base /app/video-processor-worker .

CMD ["/app/video-processor-worker"]