NAME=service

clean:
	rm ${NAME}

build:
	go build -o ${NAME}

rebuild: clean build

test:
	go test -v -count=1 ./...

.PHONY: cover

cover:
	go test -short -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out
	rm coverage.out