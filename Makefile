.PHONY: all clean

# Build flags to reduce binary size
# -ldflags="-s -w": strip debug info and symbol table (reduces ~20-30%)
# -trimpath: remove file system paths for better reproducibility
# Note: Static linking and cross-platform functionality are preserved
LDFLAGS=-ldflags="-s -w" -trimpath

all:
	go build $(LDFLAGS) -o bin/fm_train cmd/fm_train/main.go
	go build $(LDFLAGS) -o bin/fm_predict cmd/fm_predict/main.go
	go build $(LDFLAGS) -o bin/model_bin_tool cmd/model_bin_tool/main.go

clean:
	rm -f bin/fm_train bin/fm_predict bin/model_bin_tool

test:
	go test -v ./pkg/...

fmt:
	go fmt ./...

.DEFAULT_GOAL := all

