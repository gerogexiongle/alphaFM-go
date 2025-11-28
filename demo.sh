#!/bin/bash

# alphaFM-go 项目演示脚本

echo "=================================================="
echo "   alphaFM-go: Go语言版本的FM算法实现"
echo "=================================================="
echo
echo "项目位置: /data/xiongle/alphaFM-go"
echo "原C++版本: /data/xiongle/alphaFM"
echo
echo "=================================================="
echo

# 显示项目信息
echo "📊 项目统计:"
echo "-------------------"
echo "Go代码总行数: $(find . -name "*.go" | xargs wc -l | tail -1 | awk '{print $1}')"
echo "源文件数量: $(find . -name "*.go" | wc -l)"
echo "可执行文件: $(ls bin/ | wc -l) 个"
echo

# 显示文件大小
echo "📦 编译产物:"
echo "-------------------"
ls -lh bin/ | grep -v total
echo

# 显示核心模块
echo "🏗️  核心模块:"
echo "-------------------"
echo "1. pkg/model/       - FM模型和FTRL算法"
echo "2. pkg/frame/       - 多线程框架"
echo "3. pkg/sample/      - 样本解析"
echo "4. pkg/lock/        - 锁管理"
echo "5. pkg/mem/         - 内存池"
echo "6. pkg/utils/       - 工具函数"
echo

# 显示可用命令
echo "🚀 可用命令:"
echo "-------------------"
echo "1. make             - 编译项目"
echo "2. make clean       - 清理编译产物"
echo "3. ./test.sh        - 运行测试"
echo "4. ./compare_test.sh - C++与Go版本对比"
echo

# 显示使用示例
echo "💡 使用示例:"
echo "-------------------"
echo "# 训练"
echo "cat data.txt | ./bin/fm_train -m model.txt -dim 1,1,8 -core 4"
echo
echo "# 预测"
echo "cat test.txt | ./bin/fm_predict -m model.txt -dim 8 -out result.txt"
echo
echo "# 查看帮助"
echo "./bin/fm_train -h"
echo

echo "=================================================="
echo "✅ 项目已完成并可以使用！"
echo "=================================================="
echo
echo "📚 查看文档:"
echo "  - README.md           用户文档"
echo "  - IMPLEMENTATION.md   实现说明"
echo "  - DELIVERY.md         交付文档"
echo
echo "🧪 运行测试:"
echo "  ./test.sh"
echo


