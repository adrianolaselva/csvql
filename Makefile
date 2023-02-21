PROJECT_NAME=csvql
VERSION=latest

ifndef release
override release = $(VERSION)
endif

docker-build:
	docker build --rm -f "Dockerfile" -t "adrianolaselva/$(PROJECT_NAME):$(release)" "." --build-arg VERSION=$(release)
build:
	go build -o $(PROJECT_NAME) -v ./
test:
	go test -count=1 -v ./...
run:
	go run run ./...
tidy:
	go mod tidy
deps:
	go get -d -v ./...
build-linux:
	GOOS=linux GOARCH=amd64 go build -o $(PROJECT_NAME) -v ./