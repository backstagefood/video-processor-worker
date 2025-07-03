VERSION:=$(shell cat build.yaml | sed -n 's/^ *version: \(.*\)/\1/p')
PROJECT_NAME:=$(shell cat build.yaml | sed -n 's/^ *name: \(.*\)/\1/p')
LD_FLAGS:=-ldflags "-X 'github.com/backstagefood/video-processor-worker/internal/controller/handlers.Version=${VERSION}' -X 'github.com/backstagefood/video-processor-worker/internal/controller/handlers.ProjectName=${PROJETC_NAME}'"

# load environment variables
ifneq (,$(wildcard .env))
    include .env
    export $(shell sed 's/=.*//' .env)
endif

all: update run

update: 
	@go get -u ./...
	@go mod tidy

run: swagger exec

exec:
	@echo "running "${PROJECT_NAME}" version "${VERSION}
	@go run ${LD_FLAGS} cmd/app/main.go

swagger:
# install -> go install github.com/swaggo/swag/cmd/swag@latest
	@swag init -g cmd/app/main.go -o docs/http

docker-build:
	@docker build --network=host --build-arg VERSION=$(VERSION) --build-arg PROJECT_NAME=$(PROJECT_NAME) -t $(PROJECT_NAME):$(VERSION) .

podman-build:
	@podman build --network=host --build-arg VERSION=$(VERSION) --build-arg PROJECT_NAME=$(PROJECT_NAME) -t $(PROJECT_NAME):$(VERSION) .

#  docker build --no-cache --build-arg VERSION=0.0.2 --build-arg PROJECT_NAME=video-processor-worker -t video-processor-worker:0.0.2 .
