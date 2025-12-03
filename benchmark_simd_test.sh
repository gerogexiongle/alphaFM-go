#!/bin/bash

# alphaFM 真实数据集性能对比测试脚本
# C++ vs Go 性能基准测试

set -e

echo "=========================================================="
echo "  alphaFM Performance Benchmark: C++ vs Go"
echo "=========================================================="
echo

# ===== 配置参数 =====
CPP_DIR="/data/xiongle/alphaFM"
GO_DIR="/data/xiongle/alphaFM-go"
TRAIN_DATA_DIR="/data/xiongle/data/train/feature"
TEST_DATA_DIR="/data/xiongle/data/test/feature"

# 输出目录
BENCHMARK_DIR="$GO_DIR/benchmark_results"
mkdir -p "$BENCHMARK_DIR"

# 模型文件（使用文本格式，完全兼容）
CPP_MODEL="$BENCHMARK_DIR/cpp_model.txt"
GO_MODEL="$BENCHMARK_DIR/go_model.txt"

# 预测结果文件
CPP_PREDICTION="$BENCHMARK_DIR/cpp_prediction.txt"
GO_PREDICTION="$BENCHMARK_DIR/go_prediction.txt"

# ===== 可调参数（方便快速修改） =====
# 隐向量维度 (FM 二阶交互的向量维度)
FACTOR_DIM=64

# SIMD 模式: "blas" (使用 BLAS 库加速) 或 "scalar" (纯标量计算)
SIMD_MODE="scalar"

# 并行线程数
THREAD_NUM=4

# 性能日志文件
PERF_LOG="$BENCHMARK_DIR/performance_report_dim${FACTOR_DIM}_${SIMD_MODE}.txt"

# 公共参数（C++ 和 Go 共用的基础参数）
COMMON_TRAIN_PARAMS="-dim 1,1,${FACTOR_DIM} -core ${THREAD_NUM} -w_alpha 0.05 -w_beta 1.0 -w_l1 0.1 -w_l2 5.0 -v_alpha 0.05 -v_beta 1.0 -v_l1 0.1 -v_l2 5.0 -init_stdev 0.001 -mf txt"
COMMON_PREDICT_PARAMS="-dim ${FACTOR_DIM} -mf txt"

# C++ 训练/预测参数（不支持 -simd）
CPP_TRAIN_PARAMS="$COMMON_TRAIN_PARAMS"
CPP_PREDICT_PARAMS="$COMMON_PREDICT_PARAMS"

# Go 训练/预测参数（支持 -simd）
GO_TRAIN_PARAMS="$COMMON_TRAIN_PARAMS -simd ${SIMD_MODE}"
GO_PREDICT_PARAMS="$COMMON_PREDICT_PARAMS -simd ${SIMD_MODE}"

echo "Configuration:"
echo "  C++ Directory: $CPP_DIR"
echo "  Go Directory:  $GO_DIR"
echo "  Train Data:    $TRAIN_DATA_DIR"
echo "  Test Data:     $TEST_DATA_DIR"
echo "  Results Dir:   $BENCHMARK_DIR"
echo "  Factor Dim:    $FACTOR_DIM"
echo "  SIMD Mode:     $SIMD_MODE (Go only)"
echo "  Threads:       $THREAD_NUM"
echo

# ===== 函数定义 =====

# 检查命令是否存在
check_command() {
    if ! command -v $1 &> /dev/null; then
        echo "Error: $1 is not installed"
        exit 1
    fi
}

# 获取进程峰值内存 (KB)
get_peak_memory() {
    local pid=$1
    local max_mem=0
    while kill -0 $pid 2>/dev/null; do
        if [ -f /proc/$pid/status ]; then
            local mem=$(grep VmRSS /proc/$pid/status | awk '{print $2}')
            if [ -n "$mem" ] && [ $mem -gt $max_mem ]; then
                max_mem=$mem
            fi
        fi
        sleep 0.1
    done
    echo $max_mem
}

# 合并多个part文件并计算统计信息
merge_and_stats() {
    local data_dir=$1
    local desc=$2
    echo "  $desc:"
    local file_count=$(ls -1 $data_dir/part-*.txt 2>/dev/null | wc -l)
    local total_size=$(du -sh $data_dir 2>/dev/null | awk '{print $1}')
    local total_lines=$(cat $data_dir/part-*.txt 2>/dev/null | wc -l)
    echo "    Files: $file_count"
    echo "    Total Size: $total_size"
    echo "    Total Lines: $total_lines"
    echo
}

# 格式化内存显示
format_memory() {
    local mem_kb=$1
    if [ $mem_kb -lt 1024 ]; then
        echo "${mem_kb} KB"
    elif [ $mem_kb -lt 1048576 ]; then
        echo "$(echo "scale=2; $mem_kb/1024" | bc) MB"
    else
        echo "$(echo "scale=2; $mem_kb/1048576" | bc) GB"
    fi
}

# ===== 步骤 1: 环境检查 =====
echo "Step 1: Environment Check"
echo "----------------------------------------"

check_command bc
check_command time

# 检查编译状态
if [ ! -f "$CPP_DIR/bin/fm_train" ] || [ ! -f "$CPP_DIR/bin/fm_predict" ]; then
    echo "  Compiling C++ version..."
    cd $CPP_DIR && make clean && make
    echo "  ✓ C++ version compiled"
else
    echo "  ✓ C++ version ready"
fi

if [ ! -f "$GO_DIR/bin/fm_train" ] || [ ! -f "$GO_DIR/bin/fm_predict" ]; then
    echo "  Compiling Go version..."
    cd $GO_DIR && make clean && make
    echo "  ✓ Go version compiled"
else
    echo "  ✓ Go version ready"
fi

echo

# ===== 步骤 2: 数据集信息 =====
echo "Step 2: Dataset Information"
echo "----------------------------------------"
merge_and_stats "$TRAIN_DATA_DIR" "Training Data"
merge_and_stats "$TEST_DATA_DIR" "Test Data"

# 清空性能日志
echo "=========================================================="  > $PERF_LOG
echo "  alphaFM Performance Benchmark Report"                      >> $PERF_LOG
echo "  Generated: $(date)"                                         >> $PERF_LOG
echo "=========================================================="  >> $PERF_LOG
echo                                                                >> $PERF_LOG

# ===== 步骤 3: C++ 版本训练 =====
echo "Step 3: Training with C++ Version"
echo "----------------------------------------"

echo "  Streaming training data file by file (to reduce memory pressure)..."
echo "  (Preprocessing: converting format from 'userID itemID label features...' to 'label features...')"
cd $CPP_DIR

# 使用 /usr/bin/time 获取详细资源使用信息
# 数据预处理: 原始格式 "userID itemID label feature1:value1 ..."
#            转换为     "label feature1:value1 ..."
# 逐个文件流式处理，减少内存占用
CPP_TRAIN_START=$(date +%s)
(find $TRAIN_DATA_DIR -type f -name "part-*.txt" | sort | while read file; do
    cat "$file" | awk '{printf "%s", $3; for(i=4; i<=NF; i++) printf " %s", $i; printf "\n"}'
done | /usr/bin/time -v ./bin/fm_train $CPP_TRAIN_PARAMS -m $CPP_MODEL) 2>&1 | tee $BENCHMARK_DIR/cpp_train.log &
CPP_TRAIN_PID=$!

# 监控内存使用
CPP_TRAIN_MEM=$(get_peak_memory $CPP_TRAIN_PID)
wait $CPP_TRAIN_PID
CPP_TRAIN_EXIT=$?
CPP_TRAIN_END=$(date +%s)
CPP_TRAIN_TIME=$((CPP_TRAIN_END - CPP_TRAIN_START))

if [ $CPP_TRAIN_EXIT -eq 0 ]; then
    echo "  ✓ C++ training completed"
    echo "    Time: ${CPP_TRAIN_TIME}s"
    echo "    Peak Memory: $(format_memory $CPP_TRAIN_MEM)"
else
    echo "  ✗ C++ training failed"
    exit 1
fi

# 提取训练日志中的关键信息
CPP_ITERATIONS=$(grep -oP 'iter:\s*\K\d+' $BENCHMARK_DIR/cpp_train.log | tail -1)
CPP_FINAL_LOSS=$(grep -oP 'tr-loss:\s*\K[0-9.]+' $BENCHMARK_DIR/cpp_train.log | tail -1)

echo >> $PERF_LOG
echo "C++ Training Performance:" >> $PERF_LOG
echo "  Training Time: ${CPP_TRAIN_TIME}s" >> $PERF_LOG
echo "  Peak Memory: $(format_memory $CPP_TRAIN_MEM)" >> $PERF_LOG
echo "  Iterations: $CPP_ITERATIONS" >> $PERF_LOG
echo "  Final Loss: $CPP_FINAL_LOSS" >> $PERF_LOG

echo

# ===== 步骤 4: Go 版本训练 =====
echo "Step 4: Training with Go Version"
echo "----------------------------------------"

echo "  Streaming training data file by file (to reduce memory pressure)..."
echo "  (Preprocessing: converting format from 'userID itemID label features...' to 'label features...')"
cd $GO_DIR

GO_TRAIN_START=$(date +%s)
(find $TRAIN_DATA_DIR -type f -name "part-*.txt" | sort | while read file; do
    cat "$file" | awk '{printf "%s", $3; for(i=4; i<=NF; i++) printf " %s", $i; printf "\n"}'
done | /usr/bin/time -v ./bin/fm_train $GO_TRAIN_PARAMS -m $GO_MODEL) 2>&1 | tee $BENCHMARK_DIR/go_train.log &
GO_TRAIN_PID=$!

# 监控内存使用
GO_TRAIN_MEM=$(get_peak_memory $GO_TRAIN_PID)
wait $GO_TRAIN_PID
GO_TRAIN_EXIT=$?
GO_TRAIN_END=$(date +%s)
GO_TRAIN_TIME=$((GO_TRAIN_END - GO_TRAIN_START))

if [ $GO_TRAIN_EXIT -eq 0 ]; then
    echo "  ✓ Go training completed"
    echo "    Time: ${GO_TRAIN_TIME}s"
    echo "    Peak Memory: $(format_memory $GO_TRAIN_MEM)"
else
    echo "  ✗ Go training failed"
    exit 1
fi

# 提取训练日志中的关键信息
GO_ITERATIONS=$(grep -oP 'iter:\s*\K\d+' $BENCHMARK_DIR/go_train.log | tail -1)
GO_FINAL_LOSS=$(grep -oP 'tr-loss:\s*\K[0-9.]+' $BENCHMARK_DIR/go_train.log | tail -1)

echo >> $PERF_LOG
echo "Go Training Performance:" >> $PERF_LOG
echo "  Training Time: ${GO_TRAIN_TIME}s" >> $PERF_LOG
echo "  Peak Memory: $(format_memory $GO_TRAIN_MEM)" >> $PERF_LOG
echo "  Iterations: $GO_ITERATIONS" >> $PERF_LOG
echo "  Final Loss: $GO_FINAL_LOSS" >> $PERF_LOG

echo

# ===== 步骤 5: 训练性能对比 =====
echo "Step 5: Training Performance Comparison"
echo "----------------------------------------"

# 计算性能比率
TRAIN_TIME_RATIO=$(echo "scale=2; $GO_TRAIN_TIME/$CPP_TRAIN_TIME" | bc)
TRAIN_MEM_RATIO=$(echo "scale=2; $GO_TRAIN_MEM/$CPP_TRAIN_MEM" | bc)

echo "  Training Time:"
echo "    C++: ${CPP_TRAIN_TIME}s"
echo "    Go:  ${GO_TRAIN_TIME}s"
echo "    Ratio (Go/C++): ${TRAIN_TIME_RATIO}x"
echo
echo "  Peak Memory:"
echo "    C++: $(format_memory $CPP_TRAIN_MEM)"
echo "    Go:  $(format_memory $GO_TRAIN_MEM)"
echo "    Ratio (Go/C++): ${TRAIN_MEM_RATIO}x"
echo
echo "  Model Quality:"
echo "    C++ Final Loss: $CPP_FINAL_LOSS (${CPP_ITERATIONS} iters)"
echo "    Go Final Loss:  $GO_FINAL_LOSS (${GO_ITERATIONS} iters)"
echo

echo >> $PERF_LOG
echo "Training Performance Comparison:" >> $PERF_LOG
echo "  Time Ratio (Go/C++): ${TRAIN_TIME_RATIO}x" >> $PERF_LOG
echo "  Memory Ratio (Go/C++): ${TRAIN_MEM_RATIO}x" >> $PERF_LOG
echo >> $PERF_LOG

# ===== 步骤 6: C++ 版本预测 =====
echo "Step 6: Prediction with C++ Version"
echo "----------------------------------------"

echo "  Streaming test data file by file (to reduce memory pressure)..."
echo "  (Preprocessing: converting format from 'userID itemID label features...' to 'label features...')"
cd $CPP_DIR

CPP_PRED_START=$(date +%s)
(find $TEST_DATA_DIR -type f -name "part-*.txt" | sort | while read file; do
    cat "$file" | awk '{printf "%s", $3; for(i=4; i<=NF; i++) printf " %s", $i; printf "\n"}'
done | /usr/bin/time -v ./bin/fm_predict $CPP_PREDICT_PARAMS -m $CPP_MODEL -out $CPP_PREDICTION) 2>&1 | tee $BENCHMARK_DIR/cpp_predict.log &
CPP_PRED_PID=$!

# 监控内存使用
CPP_PRED_MEM=$(get_peak_memory $CPP_PRED_PID)
wait $CPP_PRED_PID
CPP_PRED_EXIT=$?
CPP_PRED_END=$(date +%s)
CPP_PRED_TIME=$((CPP_PRED_END - CPP_PRED_START))

if [ $CPP_PRED_EXIT -eq 0 ]; then
    echo "  ✓ C++ prediction completed"
    echo "    Time: ${CPP_PRED_TIME}s"
    echo "    Peak Memory: $(format_memory $CPP_PRED_MEM)"
else
    echo "  ✗ C++ prediction failed"
    exit 1
fi

CPP_PRED_COUNT=$(wc -l < $CPP_PREDICTION)
echo "    Predictions: $CPP_PRED_COUNT"

echo >> $PERF_LOG
echo "C++ Prediction Performance:" >> $PERF_LOG
echo "  Prediction Time: ${CPP_PRED_TIME}s" >> $PERF_LOG
echo "  Peak Memory: $(format_memory $CPP_PRED_MEM)" >> $PERF_LOG
echo "  Predictions: $CPP_PRED_COUNT" >> $PERF_LOG

echo

# ===== 步骤 7: Go 版本预测 =====
echo "Step 7: Prediction with Go Version"
echo "----------------------------------------"

echo "  Streaming test data file by file (to reduce memory pressure)..."
echo "  (Preprocessing: converting format from 'userID itemID label features...' to 'label features...')"
cd $GO_DIR

GO_PRED_START=$(date +%s)
(find $TEST_DATA_DIR -type f -name "part-*.txt" | sort | while read file; do
    cat "$file" | awk '{printf "%s", $3; for(i=4; i<=NF; i++) printf " %s", $i; printf "\n"}'
done | /usr/bin/time -v ./bin/fm_predict $GO_PREDICT_PARAMS -m $GO_MODEL -out $GO_PREDICTION) 2>&1 | tee $BENCHMARK_DIR/go_predict.log &
GO_PRED_PID=$!

# 监控内存使用
GO_PRED_MEM=$(get_peak_memory $GO_PRED_PID)
wait $GO_PRED_PID
GO_PRED_EXIT=$?
GO_PRED_END=$(date +%s)
GO_PRED_TIME=$((GO_PRED_END - GO_PRED_START))

if [ $GO_PRED_EXIT -eq 0 ]; then
    echo "  ✓ Go prediction completed"
    echo "    Time: ${GO_PRED_TIME}s"
    echo "    Peak Memory: $(format_memory $GO_PRED_MEM)"
else
    echo "  ✗ Go prediction failed"
    exit 1
fi

GO_PRED_COUNT=$(wc -l < $GO_PREDICTION)
echo "    Predictions: $GO_PRED_COUNT"

echo >> $PERF_LOG
echo "Go Prediction Performance:" >> $PERF_LOG
echo "  Prediction Time: ${GO_PRED_TIME}s" >> $PERF_LOG
echo "  Peak Memory: $(format_memory $GO_PRED_MEM)" >> $PERF_LOG
echo "  Predictions: $GO_PRED_COUNT" >> $PERF_LOG

echo

# ===== 步骤 8: 预测性能对比 =====
echo "Step 8: Prediction Performance Comparison"
echo "----------------------------------------"

# 计算性能比率
PRED_TIME_RATIO=$(echo "scale=2; $GO_PRED_TIME/$CPP_PRED_TIME" | bc)
PRED_MEM_RATIO=$(echo "scale=2; $GO_PRED_MEM/$CPP_PRED_MEM" | bc)

echo "  Prediction Time:"
echo "    C++: ${CPP_PRED_TIME}s"
echo "    Go:  ${GO_PRED_TIME}s"
echo "    Ratio (Go/C++): ${PRED_TIME_RATIO}x"
echo
echo "  Peak Memory:"
echo "    C++: $(format_memory $CPP_PRED_MEM)"
echo "    Go:  $(format_memory $GO_PRED_MEM)"
echo "    Ratio (Go/C++): ${PRED_MEM_RATIO}x"
echo

echo >> $PERF_LOG
echo "Prediction Performance Comparison:" >> $PERF_LOG
echo "  Time Ratio (Go/C++): ${PRED_TIME_RATIO}x" >> $PERF_LOG
echo "  Memory Ratio (Go/C++): ${PRED_MEM_RATIO}x" >> $PERF_LOG
echo >> $PERF_LOG

# ===== 步骤 9: 预测结果对比 =====
echo "Step 9: Prediction Results Comparison"
echo "----------------------------------------"

echo "  Comparing first 10 predictions:"
echo
echo "  C++ predictions:"
head -10 $CPP_PREDICTION | nl
echo
echo "  Go predictions:"
head -10 $GO_PREDICTION | nl
echo

# 计算预测值的统计差异
echo "  Computing statistical differences..."

# 使用 awk 计算差异
awk 'NR==FNR{cpp[NR]=$0; next} {diff=$0-cpp[FNR]; sum+=diff; sumsq+=diff*diff; abssum+=(diff<0?-diff:diff); count++} END {
    if(count>0) {
        mean=sum/count;
        variance=sumsq/count-mean*mean;
        stddev=sqrt(variance);
        mae=abssum/count;
        printf "    Mean Difference: %.6f\n", mean;
        printf "    Std Deviation: %.6f\n", stddev;
        printf "    Mean Absolute Error: %.6f\n", mae;
        printf "    Total Comparisons: %d\n", count;
    }
}' $CPP_PREDICTION $GO_PREDICTION | tee -a $PERF_LOG

echo

# ===== 步骤 10: 生成最终报告 =====
echo "Step 10: Final Report"
echo "----------------------------------------"

cat >> $PERF_LOG << EOF

========================================================
Overall Summary
========================================================

Training:
  C++: ${CPP_TRAIN_TIME}s, $(format_memory $CPP_TRAIN_MEM)
  Go:  ${GO_TRAIN_TIME}s, $(format_memory $GO_TRAIN_MEM)
  Performance: Go is ${TRAIN_TIME_RATIO}x in time, ${TRAIN_MEM_RATIO}x in memory

Prediction:
  C++: ${CPP_PRED_TIME}s, $(format_memory $CPP_PRED_MEM)
  Go:  ${GO_PRED_TIME}s, $(format_memory $GO_PRED_MEM)
  Performance: Go is ${PRED_TIME_RATIO}x in time, ${PRED_MEM_RATIO}x in memory

Model Files:
  C++ Model: $(ls -lh $CPP_MODEL | awk '{print $5}')
  Go Model:  $(ls -lh $GO_MODEL | awk '{print $5}')

Result Files:
  C++ Predictions: $CPP_PRED_COUNT lines
  Go Predictions:  $GO_PRED_COUNT lines

========================================================
EOF

echo "  ✓ Performance report saved to: $PERF_LOG"
echo

# 显示报告摘要
cat $PERF_LOG

# ===== 清理选项 =====
echo
echo "=========================================================="
echo "Benchmark completed!"
echo "=========================================================="
echo
echo "Results saved in: $BENCHMARK_DIR"
echo "  - Performance report: performance_report.txt"
echo "  - Training logs: cpp_train.log, go_train.log"
echo "  - Prediction logs: cpp_predict.log, go_predict.log"
echo "  - Models: cpp_model.txt, go_model.txt (text format)"
echo "  - Predictions: cpp_prediction.txt, go_prediction.txt"
echo

read -p "Delete benchmark files? (y/n) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "Cleaning up..."
    rm -f $CPP_MODEL $GO_MODEL
    rm -f $CPP_PREDICTION $GO_PREDICTION
    rm -f $BENCHMARK_DIR/*.log
    echo "✓ Cleaned up (performance report kept)"
fi

echo
echo "Done!"

