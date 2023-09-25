NAME=service

${NAME}: build

clean:
	rm ${NAME}

build:
	go build -o ${NAME}

rebuild: clean build

test: ${NAME}
	go test -v -short -race -count=1 ./...

.PHONY: cover

cover: ${NAME}
	go test -v -short -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out
	rm coverage.out