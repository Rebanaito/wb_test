NAME=service

clean:
	rm ${NAME}

build:
	go build -o ${NAME}

rebuild: clean build

test:
	go test -v -short -race -count=1 ./...

.PHONY: cover

cover:
	go test -v -short -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out
	rm coverage.out