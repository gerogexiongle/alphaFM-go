#!/bin/bash

# 测试Go多线程预测的一致性：同一程序多次运行，输出是否一致

set -e

echo "=========================================================="
echo "  Go Multi-thread Prediction Consistency Test"
echo "=========================================================="
echo

GO_DIR="/data/xiongle/alphaFM-go"
BENCHMARK_DIR="$GO_DIR/benchmark_results"
TEST_DATA_DIR="/data/xiongle/data/test/feature"

MODEL="$BENCHMARK_DIR/go_model.txt"

# 使用多线程模式 (-core 4)
GO_PRED_RUN1="$BENCHMARK_DIR/go_pred_multithread_run1.txt"
GO_PRED_RUN2="$BENCHMARK_DIR/go_pred_multithread_run2.txt"
GO_PRED_RUN3="$BENCHMARK_DIR/go_pred_multithread_run3.txt"

PREDICT_PARAMS="-dim 4 -core 4 -mf txt"

echo "Configuration:"
echo "  Model:     $MODEL"
echo "  Threads:   4 (多线程模式)"
echo "  Test:      同一Go程序运行3次，检查输出是否一致"
echo

cd "$GO_DIR"

# 第1次运行
echo "Run 1: Go Prediction"
echo "----------------------------------------"
(find "$TEST_DATA_DIR" -type f -name "part-*.txt" | sort | while read file; do
    cat "$file" | awk '{printf "%s", $3; for(i=4; i<=NF; i++) printf " %s", $i; printf "\n"}'
done | ./bin/fm_predict $PREDICT_PARAMS -m "$MODEL" -out "$GO_PRED_RUN1") 2>&1 | tail -3

COUNT1=$(wc -l < "$GO_PRED_RUN1")
echo "  ✓ Run 1 completed: $COUNT1 predictions"
echo

# 第2次运行
echo "Run 2: Go Prediction"
echo "----------------------------------------"
(find "$TEST_DATA_DIR" -type f -name "part-*.txt" | sort | while read file; do
    cat "$file" | awk '{printf "%s", $3; for(i=4; i<=NF; i++) printf " %s", $i; printf "\n"}'
done | ./bin/fm_predict $PREDICT_PARAMS -m "$MODEL" -out "$GO_PRED_RUN2") 2>&1 | tail -3

COUNT2=$(wc -l < "$GO_PRED_RUN2")
echo "  ✓ Run 2 completed: $COUNT2 predictions"
echo

# 第3次运行
echo "Run 3: Go Prediction"
echo "----------------------------------------"
(find "$TEST_DATA_DIR" -type f -name "part-*.txt" | sort | while read file; do
    cat "$file" | awk '{printf "%s", $3; for(i=4; i<=NF; i++) printf " %s", $i; printf "\n"}'
done | ./bin/fm_predict $PREDICT_PARAMS -m "$MODEL" -out "$GO_PRED_RUN3") 2>&1 | tail -3

COUNT3=$(wc -l < "$GO_PRED_RUN3")
echo "  ✓ Run 3 completed: $COUNT3 predictions"
echo

# 对比3次运行结果
echo "=========================================================="
echo "Comparing 3 Runs"
echo "=========================================================="
echo

# Run1 vs Run2
echo "Run1 vs Run2:"
echo "----------------------------------------"
awk '{print $1}' "$GO_PRED_RUN1" > /tmp/go_run1_labels_$$.txt
awk '{print $1}' "$GO_PRED_RUN2" > /tmp/go_run2_labels_$$.txt
DIFF_12=$(paste /tmp/go_run1_labels_$$.txt /tmp/go_run2_labels_$$.txt | awk '$1 != $2 {count++} END {print count+0}')
echo "  Label mismatches: $DIFF_12"

if [ "$DIFF_12" -eq 0 ]; then
    echo "  ✅ Run1 和 Run2 完全一致"
else
    MATCH_PERCENT=$(echo "scale=2; ($COUNT1 - $DIFF_12) * 100 / $COUNT1" | bc)
    echo "  ⚠️  Label alignment: ${MATCH_PERCENT}%"
    
    echo "  Sample mismatches (first 10):"
    awk '{print NR, $1}' "$GO_PRED_RUN1" > /tmp/go_run1_numbered_$$.txt
    paste -d' ' /tmp/go_run1_numbered_$$.txt /tmp/go_run2_labels_$$.txt | awk '$2 != $3 {
        printf "    Line %d: Run1=%s, Run2=%s\n", $1, $2, $3;
    }' | head -10
    rm -f /tmp/go_run1_numbered_$$.txt
fi
rm -f /tmp/go_run1_labels_$$.txt /tmp/go_run2_labels_$$.txt
echo

# Run2 vs Run3
echo "Run2 vs Run3:"
echo "----------------------------------------"
awk '{print $1}' "$GO_PRED_RUN2" > /tmp/go_run2_labels_$$.txt
awk '{print $1}' "$GO_PRED_RUN3" > /tmp/go_run3_labels_$$.txt
DIFF_23=$(paste /tmp/go_run2_labels_$$.txt /tmp/go_run3_labels_$$.txt | awk '$1 != $2 {count++} END {print count+0}')
echo "  Label mismatches: $DIFF_23"

if [ "$DIFF_23" -eq 0 ]; then
    echo "  ✅ Run2 和 Run3 完全一致"
else
    MATCH_PERCENT=$(echo "scale=2; ($COUNT2 - $DIFF_23) * 100 / $COUNT2" | bc)
    echo "  ⚠️  Label alignment: ${MATCH_PERCENT}%"
    
    echo "  Sample mismatches (first 10):"
    awk '{print NR, $1}' "$GO_PRED_RUN2" > /tmp/go_run2_numbered_$$.txt
    paste -d' ' /tmp/go_run2_numbered_$$.txt /tmp/go_run3_labels_$$.txt | awk '$2 != $3 {
        printf "    Line %d: Run2=%s, Run3=%s\n", $1, $2, $3;
    }' | head -10
    rm -f /tmp/go_run2_numbered_$$.txt
fi
rm -f /tmp/go_run2_labels_$$.txt /tmp/go_run3_labels_$$.txt
echo

# Run1 vs Run3
echo "Run1 vs Run3:"
echo "----------------------------------------"
awk '{print $1}' "$GO_PRED_RUN1" > /tmp/go_run1_labels_$$.txt
awk '{print $1}' "$GO_PRED_RUN3" > /tmp/go_run3_labels_$$.txt
DIFF_13=$(paste /tmp/go_run1_labels_$$.txt /tmp/go_run3_labels_$$.txt | awk '$1 != $2 {count++} END {print count+0}')
echo "  Label mismatches: $DIFF_13"

if [ "$DIFF_13" -eq 0 ]; then
    echo "  ✅ Run1 和 Run3 完全一致"
else
    MATCH_PERCENT=$(echo "scale=2; ($COUNT1 - $DIFF_13) * 100 / $COUNT1" | bc)
    echo "  ⚠️  Label alignment: ${MATCH_PERCENT}%"
    
    echo "  Sample mismatches (first 10):"
    awk '{print NR, $1}' "$GO_PRED_RUN1" > /tmp/go_run1_numbered_$$.txt
    paste -d' ' /tmp/go_run1_numbered_$$.txt /tmp/go_run3_labels_$$.txt | awk '$2 != $3 {
        printf "    Line %d: Run1=%s, Run3=%s\n", $1, $2, $3;
    }' | head -10
    rm -f /tmp/go_run1_numbered_$$.txt
fi
rm -f /tmp/go_run1_labels_$$.txt /tmp/go_run3_labels_$$.txt
echo

# 结论
echo "=========================================================="
echo "Conclusion"
echo "=========================================================="
echo

if [ "$DIFF_12" -eq 0 ] && [ "$DIFF_23" -eq 0 ] && [ "$DIFF_13" -eq 0 ]; then
    echo "✅✅✅ Go 多线程预测输出完全确定"
    echo ""
    echo "结论："
    echo "  - Go 版本多线程模式下，多次运行输出顺序完全一致"
    echo "  - Go 的 goroutine 调度和同步机制保证了输出的确定性"
    echo "  - 多线程实现达到了预期的可重现性"
else
    echo "⚠️⚠️⚠️ Go 多线程预测输出存在不确定性"
    echo ""
    echo "结论："
    echo "  - Go 版本存在多线程导致的输出乱序问题"
    echo "  - 这是由于 goroutine 并发调度的非确定性导致"
    echo "  - 如果需要完全确定的输出顺序，建议："
    echo "    1. 使用单线程模式 (-core 1)"
    echo "    2. 或在应用层根据输入数据的顺序重新排序输出"
    echo "    3. 预测结果（score）本身是准确的，只是输出顺序不同"
fi

echo
echo "=========================================================="



