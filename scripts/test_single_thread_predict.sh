#!/bin/bash

# 使用单线程预测测试label对齐性

set -e

echo "=========================================================="
echo "  Single Thread Prediction Test"
echo "=========================================================="
echo

CPP_DIR="/data/xiongle/alphaFM"
GO_DIR="/data/xiongle/alphaFM-go"
BENCHMARK_DIR="$GO_DIR/benchmark_results"
TEST_DATA_DIR="/data/xiongle/data/test/feature"

MODEL="$BENCHMARK_DIR/go_model.txt"

# 使用单线程 (-core 1)
CPP_PRED="$BENCHMARK_DIR/cpp_pred_single_thread.txt"
GO_PRED="$BENCHMARK_DIR/go_pred_single_thread.txt"
PREDICT_PARAMS="-dim 4 -core 1 -mf txt"

echo "Configuration:"
echo "  Model:     $MODEL"
echo "  Threads:   1 (单线程模式)"
echo "  C++ Out:   $CPP_PRED"
echo "  Go Out:    $GO_PRED"
echo

# C++预测
echo "Step 1: C++ Prediction (single thread)"
echo "----------------------------------------"
cd "$CPP_DIR"
(find "$TEST_DATA_DIR" -type f -name "part-*.txt" | sort | while read file; do
    cat "$file" | awk '{printf "%s", $3; for(i=4; i<=NF; i++) printf " %s", $i; printf "\n"}'
done | ./bin/fm_predict $PREDICT_PARAMS -m "$MODEL" -out "$CPP_PRED") 2>&1 | tail -3

CPP_COUNT=$(wc -l < "$CPP_PRED")
echo "  ✓ C++ completed: $CPP_COUNT predictions"
echo

# Go预测
echo "Step 2: Go Prediction (single thread)"
echo "----------------------------------------"
cd "$GO_DIR"
(find "$TEST_DATA_DIR" -type f -name "part-*.txt" | sort | while read file; do
    cat "$file" | awk '{printf "%s", $3; for(i=4; i<=NF; i++) printf " %s", $i; printf "\n"}'
done | ./bin/fm_predict $PREDICT_PARAMS -m "$MODEL" -out "$GO_PRED") 2>&1 | tail -3

GO_COUNT=$(wc -l < "$GO_PRED")
echo "  ✓ Go completed: $GO_COUNT predictions"
echo

# Label对齐检查
echo "Step 3: Label Alignment Check"
echo "----------------------------------------"

if [ "$CPP_COUNT" -ne "$GO_COUNT" ]; then
    echo "  ❌ Different prediction counts!"
    exit 1
fi

# 检查label对齐
LABEL_DIFF_COUNT=$(awk '{print $1}' "$CPP_PRED" > /tmp/cpp_labels_$$.txt && \
    awk '{print $1}' "$GO_PRED" > /tmp/go_labels_$$.txt && \
    paste /tmp/cpp_labels_$$.txt /tmp/go_labels_$$.txt | awk '$1 != $2 {count++} END {print count+0}' && \
    rm -f /tmp/cpp_labels_$$.txt /tmp/go_labels_$$.txt)

echo "  Total samples:    $CPP_COUNT"
echo "  Label mismatches: $LABEL_DIFF_COUNT"

if [ "$LABEL_DIFF_COUNT" -eq 0 ]; then
    echo "  ✅✅✅ All labels perfectly aligned!"
    echo
    echo "Conclusion:"
    echo "  单线程模式下，C++和Go的label输出顺序完全一致"
    echo "  这证明多线程是导致label乱序的原因"
else
    MATCH_PERCENT=$(echo "scale=2; ($CPP_COUNT - $LABEL_DIFF_COUNT) * 100 / $CPP_COUNT" | bc)
    echo "  Label alignment: ${MATCH_PERCENT}%"
    
    echo
    echo "Sample mismatches (first 10):"
    awk '{print NR, $1}' "$CPP_PRED" > /tmp/cpp_pred_numbered_$$.txt
    awk '{print $1}' "$GO_PRED" > /tmp/go_pred_only_$$.txt
    paste -d' ' /tmp/cpp_pred_numbered_$$.txt /tmp/go_pred_only_$$.txt | awk '$2 != $3 {
        printf "    Line %d: C++ label=%s, Go label=%s\n", $1, $2, $3;
    }' | head -10
    rm -f /tmp/cpp_pred_numbered_$$.txt /tmp/go_pred_only_$$.txt
fi

echo
echo "=========================================================="

