#!/bin/bash

# 数据预处理测试脚本

echo "========================================"
echo "  Data Preprocessing Test"
echo "========================================"
echo

TRAIN_DATA_DIR="/data/xiongle/data/train/feature"
TEST_FILE="/data/xiongle/alphaFM-go/test_data.txt"

echo "Step 1: Show original data format (from real dataset)"
echo "----------------------------------------"
head -3 $TRAIN_DATA_DIR/part-00275-39f0af10-858e-468b-88d6-e7938549d620-c000.txt
echo

echo "Step 2: Show converted format (for alphaFM)"
echo "----------------------------------------"
head -3 $TRAIN_DATA_DIR/part-00275-39f0af10-858e-468b-88d6-e7938549d620-c000.txt | awk '{printf "%s", $3; for(i=4; i<=NF; i++) printf " %s", $i; printf "\n"}'
echo

echo "Step 3: Test with Go fm_train (first 1000 lines)"
echo "----------------------------------------"
cd /data/xiongle/alphaFM-go

if [ ! -f bin/fm_train ]; then
    echo "Building fm_train..."
    make
fi

head -1000 $TRAIN_DATA_DIR/part-00275-39f0af10-858e-468b-88d6-e7938549d620-c000.txt | \
    awk '{printf "%s", $3; for(i=4; i<=NF; i++) printf " %s", $i; printf "\n"}' | \
    ./bin/fm_train -dim 1,1,4 -core 2 -w_l1 0.05 -v_l1 0.05 -init_stdev 0.001 -m /tmp/test_model.bin 2>&1 | tail -5

echo
echo "✓ Data preprocessing test completed!"
echo

# Clean up
rm -f /tmp/test_model.bin

