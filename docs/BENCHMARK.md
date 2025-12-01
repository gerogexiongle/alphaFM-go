# alphaFM 基准测试脚本

## 简介

`benchmark_test.sh` 是一个全面的性能基准测试脚本，用于对比 C++ 版本和 Go 版本的 alphaFM 在真实大规模数据集上的性能表现。

## 功能特性

### 1. 训练性能测试
- ⏱️ **训练时间**：精确测量两个版本的训练耗时
- 💾 **内存占用**：实时监控训练过程的峰值内存使用
- 📈 **训练质量**：记录迭代次数和最终损失值

### 2. 预测性能测试  
- ⏱️ **预测时间**：测量预测阶段的时间消耗
- 💾 **内存占用**：监控预测过程的峰值内存
- 📊 **预测数量**：统计预测结果数量

### 3. 结果分析
- 📉 **统计对比**：计算两个版本预测结果的统计差异
- 📋 **性能比率**：Go vs C++ 的性能倍数
- 📄 **完整报告**：生成详细的性能分析报告

## 数据集配置

脚本默认使用以下数据集：
- **训练数据**: `/data/xiongle/data/train/feature/*.txt`
- **测试数据**: `/data/xiongle/data/test/feature/*.txt`

支持 Spark 输出的分区文件格式（part-*.txt）。

### 数据格式

原始数据格式（Spark 输出）：
```
userID itemID label feature1:value1 feature2:value2 ...
```

示例：
```
680595 762463709_1748668636 0 17179869290:1 21474836482:1 25769803779:1
18824965 18824965_1595485745 1 17179869408:1 21474836490:1 25769803776:1
```

脚本会自动将数据转换为 alphaFM 期望的格式：
```
label feature1:value1 feature2:value2 ...
```

转换后：
```
0 17179869290:1 21474836482:1 25769803779:1
1 17179869408:1 21474836490:1 25769803776:1
```

这个转换通过 `awk` 命令在管道中自动完成，无需手动预处理数据。

### 流式处理优化

为了减少内存压力，脚本采用**逐个文件流式处理**的方式：

```bash
find $TRAIN_DATA_DIR -type f -name "part-*.txt" | sort | while read file; do
    cat "$file" | awk '{printf "%s", $3; for(i=4; i<=NF; i++) printf " %s", $i; printf "\n"}'
done | ./bin/fm_train [参数...]
```

**优点**：
- ✅ 不需要一次性加载所有文件到内存
- ✅ 逐个文件读取和处理，内存占用更低
- ✅ 适合处理超大规模数据集（数十 GB）
- ✅ 使用 `sort` 确保文件按顺序处理（可重现性）

## 使用方法

### 基础使用

```bash
cd /data/xiongle/alphaFM-go
./benchmark_test.sh
```

### 前置条件

1. **编译环境**：确保 C++ 和 Go 版本都可以编译
2. **数据准备**：训练和测试数据集已准备好
3. **依赖工具**：需要安装 `bc` 命令（用于浮点数计算）

```bash
# CentOS/RHEL 安装 bc
sudo yum install -y bc
```

## 训练参数

脚本默认使用以下训练参数：

```bash
-dim 1,1,4           # 因子分解维度
-core 4              # 并行线程数
-w_alpha 0.05        # w 参数的 alpha
-w_beta 1.0          # w 参数的 beta
-w_l1 0.1            # w 参数的 L1 正则化
-w_l2 5.0            # w 参数的 L2 正则化
-v_alpha 0.05        # v 参数的 alpha
-v_beta 1.0          # v 参数的 beta
-v_l1 0.1            # v 参数的 L1 正则化
-v_l2 5.0            # v 参数的 L2 正则化
-init_stdev 0.001    # 初始化标准差
```

## 输出结果

所有结果保存在 `benchmark_results/` 目录：

```
benchmark_results/
├── performance_report.txt      # 完整性能报告（主要输出）
├── cpp_train.log              # C++ 训练详细日志
├── go_train.log               # Go 训练详细日志
├── cpp_predict.log            # C++ 预测详细日志
├── go_predict.log             # Go 预测详细日志
├── cpp_model.bin              # C++ 训练的模型文件
├── go_model.bin               # Go 训练的模型文件
├── cpp_prediction.txt         # C++ 预测结果
└── go_prediction.txt          # Go 预测结果
```

## 性能报告示例

```
==========================================================
  alphaFM Performance Benchmark Report
  Generated: Fri Nov 28 16:00:00 CST 2025
==========================================================

C++ Training Performance:
  Training Time: 1234s
  Peak Memory: 8.5 GB
  Iterations: 15
  Final Loss: 0.123456

Go Training Performance:
  Training Time: 1456s
  Peak Memory: 9.2 GB
  Iterations: 15
  Final Loss: 0.123458

Training Performance Comparison:
  Time Ratio (Go/C++): 1.18x
  Memory Ratio (Go/C++): 1.08x

C++ Prediction Performance:
  Prediction Time: 234s
  Peak Memory: 2.1 GB
  Predictions: 3418591

Go Prediction Performance:
  Prediction Time: 267s
  Peak Memory: 2.3 GB
  Predictions: 3418591

Prediction Performance Comparison:
  Time Ratio (Go/C++): 1.14x
  Memory Ratio (Go/C++): 1.10x

Prediction Results Comparison:
  Mean Difference: 0.000001
  Std Deviation: 0.000023
  Mean Absolute Error: 0.000015
  Total Comparisons: 3418591

==========================================================
Overall Summary
==========================================================

Training:
  C++: 1234s, 8.5 GB
  Go:  1456s, 9.2 GB
  Performance: Go is 1.18x in time, 1.08x in memory

Prediction:
  C++: 234s, 2.1 GB
  Go:  267s, 2.3 GB
  Performance: Go is 1.14x in time, 1.10x in memory

Model Files:
  C++ Model: 45M
  Go Model:  45M

Result Files:
  C++ Predictions: 3418591 lines
  Go Predictions:  3418591 lines
==========================================================
```

## 修改配置

如果需要自定义配置，可以编辑脚本中的以下变量：

```bash
# 目录配置
CPP_DIR="/data/xiongle/alphaFM"
GO_DIR="/data/xiongle/alphaFM-go"
TRAIN_DATA_DIR="/data/xiongle/data/train/feature"
TEST_DATA_DIR="/data/xiongle/data/test/feature"

# 训练参数
TRAIN_PARAMS="-dim 1,1,4 -core 4 ..."

# 预测参数
PREDICT_PARAMS="-dim 4"
```

## 注意事项

1. **数据量大**：真实数据集可能非常大（数GB），测试会需要较长时间
2. **内存要求**：确保系统有足够的内存运行测试
3. **磁盘空间**：确保有足够的磁盘空间存储模型和预测结果
4. **进程监控**：脚本会在后台监控进程内存使用，不会影响性能
5. **清理选项**：测试完成后可选择删除临时文件，性能报告会保留

## 故障排查

### 问题：训练或预测失败

检查日志文件：
```bash
cat benchmark_results/cpp_train.log
cat benchmark_results/go_train.log
```

### 问题：内存不足

减少 `-core` 参数值：
```bash
TRAIN_PARAMS="-dim 1,1,4 -core 2 ..."  # 从 4 改为 2
```

### 问题：bc 命令未找到

```bash
# CentOS/RHEL
sudo yum install -y bc

# Ubuntu/Debian
sudo apt-get install -y bc
```

## 与简单测试的区别

| 特性 | compare_test.sh | benchmark_test.sh |
|------|----------------|-------------------|
| 数据集 | 简单测试数据 | 真实大规模数据 |
| 性能监控 | 基础时间 | 详细时间+内存 |
| 内存监控 | ❌ | ✅ 实时监控峰值 |
| 统计分析 | 简单对比 | 详细统计差异 |
| 报告生成 | 终端输出 | 完整报告文件 |
| 日志保存 | ❌ | ✅ 完整日志 |

## 许可证

与主项目相同。

