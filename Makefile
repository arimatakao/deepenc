run:
	go run main.go -config ./config.yaml
build:
	mkdir bin
	go build -o ./bin/deepenc main.go
test:
	go test -v ./...
clear:
	rm -rf ./bin