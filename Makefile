
GO_BIN_NAME = out

build:
	go build -o $(GO_BIN_NAME) src/main.go
run: build
	./$(GO_BIN_NAME)
watch: build
	fswatch -r ./src | xargs -n1 -I{} make build
