# Scripts 工具脚本目录

本目录包含了 alphaFM-go 项目的各种测试和工具脚本。

## 📋 脚本列表

### 基础测试脚本

| 脚本名称 | 功能说明 | 用途 |
|---------|---------|------|
| `test.sh` | 快速功能测试 | 使用小数据集测试基本训练和预测功能 |
| `demo.sh` | 演示脚本 | 展示基本使用流程 |

### 对比测试脚本

| 脚本名称 | 功能说明 | 用途 |
|---------|---------|------|
| `compare_test.sh` | C++ vs Go 对比测试 | 在小数据集上对比两个版本的结果 |
| `compare_prediction.sh` | 预测结果对比 | 对比预测结果的差异 |

### 专项测试脚本

| 脚本名称 | 功能说明 | 用途 |
|---------|---------|------|
| `test_binary_model.sh` | 二进制模型测试 | 测试二进制模型的保存和加载 |
| `test_cpp_multithread_consistency.sh` | C++多线程一致性测试 | 验证C++版本多线程的一致性 |
| `test_go_multithread_consistency.sh` | Go多线程一致性测试 | 验证Go版本多线程的一致性 |
| `test_single_thread_predict.sh` | 单线程预测测试 | 测试单线程预测功能 |
| `test_seed.sh` | 随机种子测试 | 测试随机种子对结果的影响 |
| `test_data_preprocessing.sh` | 数据预处理测试 | 测试数据格式转换和流式处理 |

### 数据检查脚本

| 脚本名称 | 功能说明 | 用途 |
|---------|---------|------|
| `check_label_alignment.sh` | 标签对齐检查 | 检查预测结果中标签是否对齐 |

## 🚀 使用方法

### 从项目根目录运行

```bash
cd /data/xiongle/alphaFM-go

# 运行基础测试
./scripts/test.sh

# 运行演示
./scripts/demo.sh

# 对比测试
./scripts/compare_test.sh
```

### 直接在 scripts 目录运行

```bash
cd /data/xiongle/alphaFM-go/scripts

# 需要先回到根目录，因为脚本依赖根目录的相对路径
cd ..
./scripts/test.sh
```

## 📝 注意事项

1. **运行位置**: 大多数脚本需要在项目根目录下运行
2. **编译依赖**: 运行测试脚本前需要先编译项目 (`make`)
3. **数据依赖**: 部分脚本需要特定的测试数据文件
4. **基准测试**: 大规模基准测试请使用根目录的 `benchmark_test.sh`

## 🔗 相关文档

- 主 README: [../README.md](../README.md)
- 基准测试文档: [../docs/BENCHMARK.md](../docs/BENCHMARK.md)
- 测试报告: [../docs/TEST_REPORT.md](../docs/TEST_REPORT.md)

## 📧 维护

这些脚本是项目测试和开发的重要组成部分，请保持脚本的可维护性和文档的更新。


