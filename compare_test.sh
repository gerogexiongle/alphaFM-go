#!/bin/bash

# C++版本和Go版本对比测试脚本

set -e

echo "======================================"
echo "  alphaFM: C++ vs Go Comparison Test"
echo "======================================"
echo

CPP_DIR="/data/xiongle/alphaFM"
GO_DIR="/data/xiongle/alphaFM-go"
TEST_DATA="$GO_DIR/test_data.txt"

# 确保两个版本都已编译
echo "Step 1: Checking if both versions are compiled..."

if [ ! -f "$CPP_DIR/bin/fm_train" ]; then
    echo "C++ version not found. Compiling..."
    cd $CPP_DIR && make
fi

if [ ! -f "$GO_DIR/bin/fm_train" ]; then
    echo "Go version not found. Compiling..."
    cd $GO_DIR && make
fi

echo "✓ Both versions are ready"
echo

# 训练参数
PARAMS="-dim 1,1,4 -core 2 -w_l1 0.05 -v_l1 0.05 -init_stdev 0.001"

# C++版本训练
echo "Step 2: Training with C++ version..."
cd $CPP_DIR
time cat $TEST_DATA | ./bin/fm_train $PARAMS -m cpp_model.txt 2>&1 | tail -3
echo "✓ C++ training completed"
echo

# Go版本训练
echo "Step 3: Training with Go version..."
cd $GO_DIR
time cat $TEST_DATA | ./bin/fm_train $PARAMS -m go_model.txt 2>&1 | tail -3
echo "✓ Go training completed"
echo

# 对比模型文件
echo "Step 4: Comparing model files..."
echo "C++ model lines: $(wc -l < $CPP_DIR/cpp_model.txt)"
echo "Go model lines:  $(wc -l < $GO_DIR/go_model.txt)"
echo

# C++版本预测
echo "Step 5: Prediction with C++ version..."
cd $CPP_DIR
cat $TEST_DATA | ./bin/fm_predict -m cpp_model.txt -dim 4 -out cpp_result.txt 2>&1 | tail -2
echo "✓ C++ prediction completed"
echo

# Go版本预测
echo "Step 6: Prediction with Go version..."
cd $GO_DIR
cat $TEST_DATA | ./bin/fm_predict -m go_model.txt -dim 4 -out go_result.txt 2>&1 | tail -2
echo "✓ Go prediction completed"
echo

# 对比预测结果
echo "Step 7: Comparing predictions..."
echo
echo "First 5 predictions from C++ version:"
head -5 $CPP_DIR/cpp_result.txt
echo
echo "First 5 predictions from Go version:"
head -5 $GO_DIR/go_result.txt
echo

# 清理测试文件
echo "======================================"
echo "Cleanup (delete test files)? (y/n)"
read -r response
if [ "$response" = "y" ]; then
    rm -f $CPP_DIR/cpp_model.txt $CPP_DIR/cpp_result.txt
    rm -f $GO_DIR/go_model.txt $GO_DIR/go_result.txt
    echo "✓ Test files cleaned up"
fi

echo
echo "======================================"
echo "  Comparison Test Completed!"
echo "======================================"

