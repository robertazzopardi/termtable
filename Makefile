run:
	go run .

tidy:
	go mod tidy

lint:
	golangci-lint run
