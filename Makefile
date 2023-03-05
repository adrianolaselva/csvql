PROJECT_NAME=csvql
VERSION=latest

ifndef release
override release = $(VERSION)
endif

all:
	git rev-parse HEAD
build:
	go build -o $(PROJECT_NAME) -v ./
test:
	go test -count=1 -v ./...
linter:
	golangci-lint run --out-format checkstyle
run:
	go run run ./...
tidy:
	go mod tidy
mod-download:
	go mod download
deps:
	go get -d -v ./...
build-linux:
	GOOS=linux GOARCH=amd64 go build -o $(PROJECT_NAME) -v ./
docker-build:
	docker build --rm -f "Dockerfile" -t "adrianolaselva/$(PROJECT_NAME):$(release)" "." --build-arg VERSION=$(release)