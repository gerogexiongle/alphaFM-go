#!/bin/bash

# 测试：设置相同随机种子能否得到相同的模型和预测结果

set -e

echo "=========================================="
echo "  测试：随机种子对模型的影响"
echo "=========================================="
echo

GO_DIR="/data/xiongle/alphaFM-go"
TEST_DATA="$GO_DIR/test_data.txt"
PARAMS="-dim 1,1,4 -core 1 -w_l1 0.05 -v_l1 0.05 -init_stdev 0.1"

cd $GO_DIR

echo "实验1：不修改代码，多次训练（每次随机种子不同）"
echo "----------------------------------------"

# 第一次训练
cat $TEST_DATA | ./bin/fm_train $PARAMS -m model1.txt 2>&1 | tail -2
echo "✓ 第一次训练完成"

# 等待1秒，确保时间戳不同
sleep 1

# 第二次训练
cat $TEST_DATA | ./bin/fm_train $PARAMS -m model2.txt 2>&1 | tail -2
echo "✓ 第二次训练完成"
echo

echo "对比两个模型："
echo "模型1 前3行："
head -3 model1.txt
echo
echo "模型2 前3行："
head -3 model2.txt
echo

# 检查模型是否相同
if diff model1.txt model2.txt > /dev/null 2>&1; then
    echo "❌ 意外：两个模型完全相同！（可能是随机种子相同）"
else
    echo "✅ 预期：两个模型不同（随机种子导致初始化不同）"
fi
echo

# 预测对比
cat $TEST_DATA | ./bin/fm_predict -m model1.txt -dim 4 -out pred1.txt 2>&1 > /dev/null
cat $TEST_DATA | ./bin/fm_predict -m model2.txt -dim 4 -out pred2.txt 2>&1 > /dev/null

echo "预测结果对比："
echo "模型1预测："
head -3 pred1.txt
echo "模型2预测："
head -3 pred2.txt
echo

if diff pred1.txt pred2.txt > /dev/null 2>&1; then
    echo "❌ 预测结果相同（小数据集可能收敛到相同解）"
else
    echo "✅ 预测结果不同"
    echo "详细差异："
    paste pred1.txt pred2.txt | head -5 | awk '{printf "行%d: pred1=%s, pred2=%s, diff=%f\n", NR, $2, $4, $4-$2}'
fi

# 清理
rm -f model1.txt model2.txt pred1.txt pred2.txt

echo
echo "=========================================="
echo "  总结"
echo "=========================================="
echo
echo "当前实现："
echo "  - C++ 使用: srand(time(NULL))"
echo "  - Go  使用: rand.Seed(time.Now().UnixNano())"
echo
echo "结论："
echo "  ✅ 两个版本都使用时间戳作为随机种子"
echo "  ✅ 每次运行时间戳不同 → 随机种子不同 → 初始参数不同"
echo "  ✅ 这是标准做法，确保不同训练run之间独立"
echo
echo "如果要复现相同结果，需要："
echo "  1. 修改代码支持手动设置随机种子"
echo "  2. 添加命令行参数如: -seed 12345"
echo "  3. C++和Go都设置相同的种子值"
echo


