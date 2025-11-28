.PHONY: all clean

all:
	go build -o bin/fm_train cmd/fm_train/main.go
	go build -o bin/fm_predict cmd/fm_predict/main.go
	go build -o bin/model_bin_tool cmd/model_bin_tool/main.go

clean:
	rm -f bin/fm_train bin/fm_predict bin/model_bin_tool

test:
	go test -v ./pkg/...

fmt:
	go fmt ./...

.DEFAULT_GOAL := all

