#!/bin/bash

# 测试C++多线程预测的一致性：同一程序多次运行，输出是否一致

set -e

echo "=========================================================="
echo "  C++ Multi-thread Prediction Consistency Test"
echo "=========================================================="
echo

CPP_DIR="/data/xiongle/alphaFM"
BENCHMARK_DIR="/data/xiongle/alphaFM-go/benchmark_results"
TEST_DATA_DIR="/data/xiongle/data/test/feature"

MODEL="$BENCHMARK_DIR/go_model.txt"

# 使用多线程模式 (-core 4)
CPP_PRED_RUN1="$BENCHMARK_DIR/cpp_pred_multithread_run1.txt"
CPP_PRED_RUN2="$BENCHMARK_DIR/cpp_pred_multithread_run2.txt"
CPP_PRED_RUN3="$BENCHMARK_DIR/cpp_pred_multithread_run3.txt"

PREDICT_PARAMS="-dim 4 -core 4 -mf txt"

echo "Configuration:"
echo "  Model:     $MODEL"
echo "  Threads:   4 (多线程模式)"
echo "  Test:      同一C++程序运行3次，检查输出是否一致"
echo

cd "$CPP_DIR"

# 第1次运行
echo "Run 1: C++ Prediction"
echo "----------------------------------------"
(find "$TEST_DATA_DIR" -type f -name "part-*.txt" | sort | while read file; do
    cat "$file" | awk '{printf "%s", $3; for(i=4; i<=NF; i++) printf " %s", $i; printf "\n"}'
done | ./bin/fm_predict $PREDICT_PARAMS -m "$MODEL" -out "$CPP_PRED_RUN1") 2>&1 | tail -3

COUNT1=$(wc -l < "$CPP_PRED_RUN1")
echo "  ✓ Run 1 completed: $COUNT1 predictions"
echo

# 第2次运行
echo "Run 2: C++ Prediction"
echo "----------------------------------------"
(find "$TEST_DATA_DIR" -type f -name "part-*.txt" | sort | while read file; do
    cat "$file" | awk '{printf "%s", $3; for(i=4; i<=NF; i++) printf " %s", $i; printf "\n"}'
done | ./bin/fm_predict $PREDICT_PARAMS -m "$MODEL" -out "$CPP_PRED_RUN2") 2>&1 | tail -3

COUNT2=$(wc -l < "$CPP_PRED_RUN2")
echo "  ✓ Run 2 completed: $COUNT2 predictions"
echo

# 第3次运行
echo "Run 3: C++ Prediction"
echo "----------------------------------------"
(find "$TEST_DATA_DIR" -type f -name "part-*.txt" | sort | while read file; do
    cat "$file" | awk '{printf "%s", $3; for(i=4; i<=NF; i++) printf " %s", $i; printf "\n"}'
done | ./bin/fm_predict $PREDICT_PARAMS -m "$MODEL" -out "$CPP_PRED_RUN3") 2>&1 | tail -3

COUNT3=$(wc -l < "$CPP_PRED_RUN3")
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
awk '{print $1}' "$CPP_PRED_RUN1" > /tmp/run1_labels_$$.txt
awk '{print $1}' "$CPP_PRED_RUN2" > /tmp/run2_labels_$$.txt
DIFF_12=$(paste /tmp/run1_labels_$$.txt /tmp/run2_labels_$$.txt | awk '$1 != $2 {count++} END {print count+0}')
echo "  Label mismatches: $DIFF_12"

if [ "$DIFF_12" -eq 0 ]; then
    echo "  ✅ Run1 和 Run2 完全一致"
else
    MATCH_PERCENT=$(echo "scale=2; ($COUNT1 - $DIFF_12) * 100 / $COUNT1" | bc)
    echo "  ⚠️  Label alignment: ${MATCH_PERCENT}%"
    
    echo "  Sample mismatches (first 10):"
    awk '{print NR, $1}' "$CPP_PRED_RUN1" > /tmp/run1_numbered_$$.txt
    paste -d' ' /tmp/run1_numbered_$$.txt /tmp/run2_labels_$$.txt | awk '$2 != $3 {
        printf "    Line %d: Run1=%s, Run2=%s\n", $1, $2, $3;
    }' | head -10
    rm -f /tmp/run1_numbered_$$.txt
fi
rm -f /tmp/run1_labels_$$.txt /tmp/run2_labels_$$.txt
echo

# Run2 vs Run3
echo "Run2 vs Run3:"
echo "----------------------------------------"
awk '{print $1}' "$CPP_PRED_RUN2" > /tmp/run2_labels_$$.txt
awk '{print $1}' "$CPP_PRED_RUN3" > /tmp/run3_labels_$$.txt
DIFF_23=$(paste /tmp/run2_labels_$$.txt /tmp/run3_labels_$$.txt | awk '$1 != $2 {count++} END {print count+0}')
echo "  Label mismatches: $DIFF_23"

if [ "$DIFF_23" -eq 0 ]; then
    echo "  ✅ Run2 和 Run3 完全一致"
else
    MATCH_PERCENT=$(echo "scale=2; ($COUNT2 - $DIFF_23) * 100 / $COUNT2" | bc)
    echo "  ⚠️  Label alignment: ${MATCH_PERCENT}%"
    
    echo "  Sample mismatches (first 10):"
    awk '{print NR, $1}' "$CPP_PRED_RUN2" > /tmp/run2_numbered_$$.txt
    paste -d' ' /tmp/run2_numbered_$$.txt /tmp/run3_labels_$$.txt | awk '$2 != $3 {
        printf "    Line %d: Run2=%s, Run3=%s\n", $1, $2, $3;
    }' | head -10
    rm -f /tmp/run2_numbered_$$.txt
fi
rm -f /tmp/run2_labels_$$.txt /tmp/run3_labels_$$.txt
echo

# Run1 vs Run3
echo "Run1 vs Run3:"
echo "----------------------------------------"
awk '{print $1}' "$CPP_PRED_RUN1" > /tmp/run1_labels_$$.txt
awk '{print $1}' "$CPP_PRED_RUN3" > /tmp/run3_labels_$$.txt
DIFF_13=$(paste /tmp/run1_labels_$$.txt /tmp/run3_labels_$$.txt | awk '$1 != $2 {count++} END {print count+0}')
echo "  Label mismatches: $DIFF_13"

if [ "$DIFF_13" -eq 0 ]; then
    echo "  ✅ Run1 和 Run3 完全一致"
else
    MATCH_PERCENT=$(echo "scale=2; ($COUNT1 - $DIFF_13) * 100 / $COUNT1" | bc)
    echo "  ⚠️  Label alignment: ${MATCH_PERCENT}%"
    
    echo "  Sample mismatches (first 10):"
    awk '{print NR, $1}' "$CPP_PRED_RUN1" > /tmp/run1_numbered_$$.txt
    paste -d' ' /tmp/run1_numbered_$$.txt /tmp/run3_labels_$$.txt | awk '$2 != $3 {
        printf "    Line %d: Run1=%s, Run3=%s\n", $1, $2, $3;
    }' | head -10
    rm -f /tmp/run1_numbered_$$.txt
fi
rm -f /tmp/run1_labels_$$.txt /tmp/run3_labels_$$.txt
echo

# 结论
echo "=========================================================="
echo "Conclusion"
echo "=========================================================="
echo

if [ "$DIFF_12" -eq 0 ] && [ "$DIFF_23" -eq 0 ] && [ "$DIFF_13" -eq 0 ]; then
    echo "✅✅✅ C++ 多线程预测输出完全确定"
    echo ""
    echo "结论："
    echo "  - C++ 版本多线程模式下，多次运行输出顺序完全一致"
    echo "  - C++ 的线程同步机制保证了输出的确定性"
    echo "  - Go 和 C++ 的多线程实现机制存在差异"
else
    echo "⚠️⚠️⚠️ C++ 多线程预测输出存在不确定性"
    echo ""
    echo "结论："
    echo "  - C++ 版本也存在多线程导致的输出乱序问题"
    echo "  - 这是多线程并发的普遍现象，不是 Go 特有的"
fi

echo
echo "=========================================================="

