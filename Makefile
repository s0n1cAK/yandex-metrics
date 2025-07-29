test:
	@echo "Testing all modules"
	go test -v ./...

build:
	@echo "Building server and agent"
	go build -o server ./cmd/server/*.go
	go build -o agent ./cmd/agent/*.go

run_server:
	./server


# Добавить build и тест TestIteration<number> \
./metricstest -test.v -test.run=^TestIteration5$ -agent-binary-path=./agent -binary-path=./server -server-port=8080 -source-path=.
