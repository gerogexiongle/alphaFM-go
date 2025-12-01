#!/bin/bash

# 检查C++和Go预测结果文件的label列是否对齐

CPP_PRED="/data/xiongle/alphaFM-go/benchmark_results/cpp_pred_from_go_model.txt"
GO_PRED="/data/xiongle/alphaFM-go/benchmark_results/go_pred_from_go_model.txt"

echo "=========================================================="
echo "  Label Alignment Check"
echo "=========================================================="
echo
echo "Comparing:"
echo "  C++: $CPP_PRED"
echo "  Go:  $GO_PRED"
echo

# 检查文件是否存在
if [ ! -f "$CPP_PRED" ]; then
    echo "Error: C++ prediction file not found!"
    exit 1
fi

if [ ! -f "$GO_PRED" ]; then
    echo "Error: Go prediction file not found!"
    exit 1
fi

# 统计行数
CPP_LINES=$(wc -l < "$CPP_PRED")
GO_LINES=$(wc -l < "$GO_PRED")

echo "File Statistics:"
echo "  C++ lines: $CPP_LINES"
echo "  Go lines:  $GO_LINES"
echo

if [ $CPP_LINES -ne $GO_LINES ]; then
    echo "❌ ERROR: Different number of lines!"
    exit 1
fi

# 前10行对比
echo "First 10 lines label comparison:"
echo "  Line | C++ Label | Go Label | Match?"
echo "  -----|-----------|----------|--------"
awk '{print $1}' "$CPP_PRED" | head -10 > /tmp/cpp_labels_head_$$.txt
awk '{print $1}' "$GO_PRED" | head -10 > /tmp/go_labels_head_$$.txt
paste /tmp/cpp_labels_head_$$.txt /tmp/go_labels_head_$$.txt | awk '{

    match_symbol = ($1 == $2) ? "✓" : "✗";
    printf "  %4d | %9s | %8s | %s\n", NR, $1, $2, match_symbol;
}'
rm -f /tmp/cpp_labels_head_$$.txt /tmp/go_labels_head_$$.txt

echo

# 完整label对比
echo "Full label alignment check..."
awk '{print $1}' "$CPP_PRED" > /tmp/cpp_labels_full_$$.txt
awk '{print $1}' "$GO_PRED" > /tmp/go_labels_full_$$.txt
LABEL_DIFF_COUNT=$(paste /tmp/cpp_labels_full_$$.txt /tmp/go_labels_full_$$.txt | awk '$1 != $2 {count++} END {print count+0}')
rm -f /tmp/cpp_labels_full_$$.txt /tmp/go_labels_full_$$.txt

echo "Results:"
echo "  Total samples:     $CPP_LINES"
echo "  Label mismatches:  $LABEL_DIFF_COUNT"

if [ $LABEL_DIFF_COUNT -eq 0 ]; then
    echo "  ✅✅✅ All labels are perfectly aligned!"
    echo
    echo "Conclusion:"
    echo "  - Input data order is consistent"
    echo "  - C++ and Go processed the same test set in the same order"
else
    MATCH_COUNT=$((CPP_LINES - LABEL_DIFF_COUNT))
    MATCH_PERCENT=$(echo "scale=2; $MATCH_COUNT * 100 / $CPP_LINES" | bc)
    echo "  ⚠️  Label alignment: ${MATCH_PERCENT}%"
    echo
    echo "Sample mismatches (first 20):"
    awk '{print NR, $1}' "$CPP_PRED" > /tmp/cpp_labels_numbered_$$.txt
    awk '{print $1}' "$GO_PRED" > /tmp/go_labels_only_$$.txt
    paste -d' ' /tmp/cpp_labels_numbered_$$.txt /tmp/go_labels_only_$$.txt | awk '$2 != $3 {
        printf "    Line %d: C++ label=%s, Go label=%s\n", $1, $2, $3;
    }' | head -20
    rm -f /tmp/cpp_labels_numbered_$$.txt /tmp/go_labels_only_$$.txt
    echo
    echo "⚠️  WARNING: Labels are not aligned!"
    echo "This could mean:"
    echo "  1. Different input data order"
    echo "  2. Some samples were skipped/duplicated"
    echo "  3. Different data preprocessing"
fi

echo
echo "=========================================================="


