test:
	@echo "Testing all modules"
	go test -v ./...

build:
	@echo "Building server and agent"
	go build -o server ./cmd/server/*.go
	go build -o agent ./cmd/agent/*.go

run_server:
	./server
# Добавить build и тест TestIteration<number>