build:
	@go build -o bin/bank-api

run:  build
	@./bin/bank-api

test: 
	@go test -v ./...
