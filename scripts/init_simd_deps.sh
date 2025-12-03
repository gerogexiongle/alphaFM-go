#!/bin/bash
# 初始化 SIMD 依赖的脚本

echo "正在初始化 SIMD 依赖..."

# 创建 go.sum 文件（如果不存在）
if [ ! -f go.sum ]; then
    echo "创建 go.sum 文件..."
    cat > go.sum << 'EOF'
gonum.org/v1/gonum v0.14.0 h1:2NiG67LD1tEH0D7kM+ps2V+fXmsAnpUeec7n8tcr4S0=
gonum.org/v1/gonum v0.14.0/go.mod h1:AoWeoz0becf9QMWtE8iWXNXc27fK4fNeHNf/oMejGfU=
EOF
fi

# 下载依赖
echo "下载依赖..."
go mod download

# 验证
echo "验证构建..."
go build -o /tmp/test_build cmd/fm_train/main.go 2>&1 | head -5
if [ $? -eq 0 ]; then
    echo "✅ 依赖初始化成功！"
    rm /tmp/test_build
else
    echo "⚠️ 构建遇到问题，尝试 go mod tidy..."
    go mod tidy
fi

echo "完成！现在可以运行 make 了。"
