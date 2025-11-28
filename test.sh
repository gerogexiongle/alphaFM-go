#!/bin/bash

# alphaFM-go 测试脚本

set -e

echo "=== alphaFM-go Test Script ==="
echo

# 编译项目
echo "1. Building project..."
make clean
make
echo "Build completed!"
echo

# 测试训练
echo "2. Testing training..."
cat test_data.txt | ./bin/fm_train \
    -m test_model.txt \
    -dim 1,1,4 \
    -core 2 \
    -w_l1 0.05 \
    -v_l1 0.05 \
    -init_stdev 0.001

if [ $? -eq 0 ]; then
    echo "Training completed successfully!"
    echo "Model saved to: test_model.txt"
else
    echo "Training failed!"
    exit 1
fi
echo

# 检查模型文件
if [ -f test_model.txt ]; then
    echo "3. Model file info:"
    wc -l test_model.txt
    echo "First 3 lines of model:"
    head -3 test_model.txt
else
    echo "Model file not found!"
    exit 1
fi
echo

# 测试预测
echo "4. Testing prediction..."
cat test_data.txt | ./bin/fm_predict \
    -m test_model.txt \
    -dim 4 \
    -out test_result.txt

if [ $? -eq 0 ]; then
    echo "Prediction completed successfully!"
    echo "Results saved to: test_result.txt"
else
    echo "Prediction failed!"
    exit 1
fi
echo

# 检查预测结果
if [ -f test_result.txt ]; then
    echo "5. Prediction results:"
    head -5 test_result.txt
else
    echo "Result file not found!"
    exit 1
fi
echo

echo "=== All tests passed! ==="
echo
echo "Cleanup test files? (y/n)"
read -r response
if [ "$response" = "y" ]; then
    rm -f test_model.txt test_result.txt
    echo "Test files cleaned up."
fi

