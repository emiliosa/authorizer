BINARY_NAME=authorize

all: build test

compile:
	echo "Compiling for OS [linux, freebsd] and Platform [arm64, amd64, 386]"
	GOOS=linux GOARCH=arm64 go build -o bin/${BINARY_NAME}-linux-arm64 main.go
	GOOS=linux GOARCH=amd64 go build -o bin/${BINARY_NAME}-linux-amd64 main.go
	GOOS=freebsd GOARCH=386 go build -o bin/${BINARY_NAME}-freebsd-386 main.go

build:
	go build -o bin/${BINARY_NAME} main.go

test:
	go test -v

run:
	go run main.go < operations.txt

clean:
	go clean
	rm bin/${BINARY_NAME}*

# lint:
#	golangci-lint run --enable-all