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
	@swag init -g cmd/app/main.go -o docs/http

swagger-install:
	@go install github.com/swaggo/swag/cmd/swag@latest

docker-build:
	@docker build --network=host --build-arg VERSION=$(VERSION) --build-arg PROJECT_NAME=$(PROJECT_NAME) -t $(PROJECT_NAME):$(VERSION) .

podman-build:
	@podman build --network=host --build-arg VERSION=$(VERSION) --build-arg PROJECT_NAME=$(PROJECT_NAME) -t $(PROJECT_NAME):$(VERSION) .

mockery:
	@mockery

mockery-install:
	@go install github.com/vektra/mockery/v3@v3.2.3

mockery-ci: mockery-install mockery

install-ci: swagger-install
	@go mod download

test:
	@go test ./... -coverpkg=$(shell go list ./... | grep -v mocks | grep -v docs | grep -v adapter | grep -v cmd/app | tr '\n' ',') -coverprofile=coverage.out -covermode=count
	@go tool cover -func=coverage.out

# test-ci: mockery-ci test
test-ci: test