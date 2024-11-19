format:
	gofumpt -l -w .
run:
	gofumpt -l -w .
	go run .
build:
	go mod tidy
	go build -tags netgo -ldflags '-s -w' -o app ./cmd/doozip
