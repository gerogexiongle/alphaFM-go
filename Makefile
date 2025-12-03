.PHONY: all clean test fmt benchmark simd-benchmark deps

# Build flags to reduce binary size
# -ldflags="-s -w": strip debug info and symbol table (reduces ~20-30%)
# -trimpath: remove file system paths for better reproducibility
# Note: Static linking and cross-platform functionality are preserved
LDFLAGS=-ldflags="-s -w" -trimpath

# 下载依赖（go build 会自动下载，这里显式声明用于 CI/CD 缓存等场景）
deps:
	go mod download

all: deps
	go build $(LDFLAGS) -o bin/fm_train cmd/fm_train/main.go
	go build $(LDFLAGS) -o bin/fm_predict cmd/fm_predict/main.go
	go build $(LDFLAGS) -o bin/model_bin_tool cmd/model_bin_tool/main.go
	go build $(LDFLAGS) -o bin/simd_benchmark cmd/simd_benchmark/main.go

clean:
	rm -f bin/fm_train bin/fm_predict bin/model_bin_tool bin/simd_benchmark

test:
	go test -v ./pkg/...

# 运行SIMD单元测试
test-simd:
	go test -v ./pkg/simd/...

# 运行标准Go benchmark
benchmark:
	go test -bench=. -benchmem ./pkg/simd/...

# 运行自定义SIMD性能对比工具
simd-benchmark: all
	@echo "Running SIMD performance comparison..."
	./bin/simd_benchmark -size 8 -iter 100000

fmt:
	go fmt ./...

.DEFAULT_GOAL := all


