format:
	gofumpt -l -w .
run:
	gofumpt -l -w .
	go run .
build:
	gofumpt -l -w .
	go build -o doozip .
