run:
	go run main.go -config ./config.yaml
build:
	mkdir bin
	go build -o ./bin/deepenc main.go
clear:
	rm -f ./bin/deepenc
	rmdir ./bin