# alphaFM-go

> Go语言实现的Factorization Machines算法库，基于FTRL优化

[![Go Version](https://img.shields.io/badge/Go-1.18+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Build](https://img.shields.io/badge/Build-Passing-brightgreen.svg)]()

## 📖 简介

**alphaFM-go** 是对经典C++版本 [alphaFM](https://github.com/CastellanZhang/alphaFM) 的Go语言重构实现。

- 🎯 **用途**: 二分类问题（CTR预估、推荐系统等）
- 🧮 **算法**: Factorization Machines (FM)
- 🚀 **优化**: FTRL (Follow The Regularized Leader)
- 💻 **语言**: Go 1.18+
- 📊 **特点**: 单机多线程、流式处理、工业级性能

### 为什么选择Go版本？

**代码质量**：
- ✨ **代码更简洁** - 比C++版本减少30%代码量
- 🔒 **类型安全** - 编译时类型检查
- ⚡ **并发简单** - goroutine比pthread更易用
- 🛡️ **内存安全** - 自动垃圾回收，无内存泄漏
- 🌍 **跨平台** - 无需修改即可编译

**真实性能优势**（基于1640万训练+341万预测的生产数据）：
- 🚀 **预测速度快39%** - 117秒 vs 71秒
- ⚡ **预测吞吐高65%** - 29,211/s vs 48,150/s  
- ⚖️ **训练速度相同** - 316秒 vs 317秒（0.3%差异）
- ✅ **结果完全一致** - 341万预测零误差
- 📦 **模型格式统一** - 2.8MB vs 2.8MB（精度对齐后）
- 🎯 **输出完全确定** - 多线程下输出顺序100%可重现（C++版本存在乱序）

**适用场景**：
- 🎯 **在线预测服务** - 预测速度快40%，响应更快
- 📊 **高并发场景** - goroutine轻量级，支持更多并发
- 🔧 **快速迭代** - 代码简洁，易于维护和扩展
- 🌐 **云原生部署** - 容器友好，资源利用率高
- 🎲 **批量预测** - 多线程输出确定，便于结果关联和验证

## ✨ 核心特性

| 特性 | 说明 | 状态 |
|------|------|------|
| FM模型 | 二阶特征交互 | ✅ |
| FTRL优化 | 在线学习算法 | ✅ |
| 多线程训练 | Goroutine并行 | ✅ |
| 多线程预测 | 高性能推理 | ✅ |
| 增量训练 | 加载旧模型继续训练 | ✅ |
| 稀疏解 | L1正则+强制稀疏 | ✅ |
| 文本模型 | 可读的模型格式 | ✅ |
| 二进制模型 | 快速加载 | ⚠️ 基础实现 |
| 流式处理 | 管道输入，无需全部加载 | ✅ |
| **SIMD优化** | **向量化加速（可选）** | **✅ 新增** |

## 🚀 快速开始

### 安装依赖

```bash
# 需要 Go 1.18 或更高版本
go version
```

### 编译

```bash
cd /data/xiongle/alphaFM-go

# 编译所有工具（自动下载依赖）
make
```

编译后在 `bin/` 目录生成4个可执行文件：
- `fm_train` - 训练程序
- `fm_predict` - 预测程序  
- `model_bin_tool` - 模型工具
- `simd_benchmark` - SIMD性能测试工具

### 快速测试

```bash
# 运行测试脚本
./scripts/test.sh

# 手动测试
cat test_data.txt | ./bin/fm_train -m model.txt -dim 1,1,4 -core 2
cat test_data.txt | ./bin/fm_predict -m model.txt -dim 4 -out result.txt
```

## 📚 使用示例

### 基础训练

```bash
cat train.txt | ./bin/fm_train \
    -m model.txt \
    -dim 1,1,8 \
    -core 4
```

### 高级训练

```bash
cat train.txt | ./bin/fm_train \
    -m model.txt \
    -dim 1,1,8 \
    -core 10 \
    -w_alpha 0.01 \
    -v_alpha 0.01 \
    -w_l1 0.05 \
    -v_l1 0.05 \
    -init_stdev 0.001 \
    -fvs 1
```

### 增量训练

```bash
cat new_data.txt | ./bin/fm_train \
    -im old_model.txt \
    -m new_model.txt \
    -dim 1,1,8 \
    -core 4
```

### 预测

```bash
cat test.txt | ./bin/fm_predict \
    -m model.txt \
    -dim 8 \
    -out result.txt \
    -core 4
```

## 🎛️ 参数说明

### 训练参数 (fm_train)

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `-m` | 输出模型路径 | 必需 |
| `-mf` | 模型格式 (txt/bin) | txt |
| `-dim` | k0,k1,k2 (bias,一阶,二阶维度) | 1,1,8 |
| `-init_stdev` | 二阶参数初始化标准差 | 0.1 |
| `-w_alpha` | w的FTRL学习率α | 0.05 |
| `-w_beta` | w的FTRL学习率β | 1.0 |
| `-w_l1` | w的L1正则 | 0.1 |
| `-w_l2` | w的L2正则 | 5.0 |
| `-v_alpha` | v的FTRL学习率α | 0.05 |
| `-v_beta` | v的FTRL学习率β | 1.0 |
| `-v_l1` | v的L1正则 | 0.1 |
| `-v_l2` | v的L2正则 | 5.0 |
| `-core` | 线程数 | 1 |
| `-im` | 初始模型路径（增量训练） | - |
| `-fvs` | 强制稀疏 (0/1) | 0 |

### 预测参数 (fm_predict)

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `-m` | 模型路径 | 必需 |
| `-mf` | 模型格式 (txt/bin) | txt |
| `-dim` | 二阶维度 | 8 |
| `-core` | 线程数 | 1 |
| `-out` | 输出路径 | 必需 |

## 📊 数据格式

### 输入样本格式（类似libsvm）

```
label feature1:value1 feature2:value2 ...
```

**示例：**
```
1 sex:1 age:0.3 f1:1 f3:0.9
0 sex:0 age:0.7 f2:0.4 f5:0.8 f8:1
```

- `label`: 1/0 或 1/-1
- `feature`: 字符串或数字
- `value`: 浮点数（建议归一化）
- 值为0的特征可省略

### 预测结果格式

```
label score
```

- `label`: 真实标签 (1/-1)
- `score`: 预测为正样本的概率 [0, 1]

## 📈 性能对比

### 基准测试结果（真实生产数据集）

**测试数据集**：
- 训练数据：1,640万条样本，107个分区文件，总大小 14GB
- 测试数据：341万条样本，200+个分区文件，总大小 5GB
- 测试参数：`-dim 1,1,4 -core 4`

**测试环境**：
- CPU: Intel(R) Xeon(R) Platinum 8378C @ 2.80GHz
- CPU 核心：8核16线程
- 内存：30GB
- 操作系统：CentOS 7
- 数据处理：流式处理，逐文件加载（低内存占用）

**训练性能**：

| 指标 | C++版本 | Go版本 | 对比 |
|------|---------|--------|------|
| **训练时间** | 316秒 | 317秒 | ⚖️ **持平** (0.3%差异) |
| **吞吐量** | 51,899样本/秒 | 51,735样本/秒 | 几乎相同 |
| 模型大小 | 2.8 MB | 2.8 MB | ✅ 完全一致 |

**预测性能**：

| 指标 | C++版本 | Go版本 | 对比 |
|------|---------|--------|------|
| **预测时间** | 117秒 | 71秒 | 🚀 **Go快39%** |
| **吞吐量** | 29,211样本/秒 | 48,150样本/秒 | 🎯 **Go快65%** |
| 预测样本数 | 3,418,591 | 3,418,591 | 完全一致 |

**结果准确性**：

| 指标 | 值 | 说明 |
|------|-----|------|
| 平均差异 | 0.000000 | 完全一致 |
| 标准差 | 0.000000 | 无偏差 |
| 平均绝对误差 | 0.000000 | 341万预测完全相同 |

### 综合评价 (dim=4, 基础对比)

> **测试配置**：dim=4, 4线程, 341万样本

| 指标 | C++ | Go | 胜者 |
|------|-----|-----|------|
| **训练速度** | 317秒 | 316秒 | ⚖️ 平手 |
| **预测速度** | 111秒 | 70秒 | 🏆 **Go** (快37%) |
| **预测吞吐** | 30,798/s | 48,837/s | 🏆 **Go** (快59%) |
| **结果准确性** | ✅ | ✅ | ⚖️ 完全一致 |
| **模型大小** | 2.8MB | 2.8MB | ⚖️ 完全一致 |
| **多线程确定性** | ❌ 乱序 | ✅ 确定 | 🏆 **Go** |
| **代码可维护性** | - | - | 🏆 Go |

**测试结论 (dim=4)**：
- ✅ **训练性能持平** - Go版本与C++版本训练速度几乎相同（0.3%差异）
- 🚀 **预测性能卓越** - Go版本预测速度快37%，吞吐量高59%
- 🎯 **结果完全一致** - 341万预测结果与C++版本完全相同，零误差
- 📦 **模型格式统一** - 模型文件大小完全一致（2.8MB），格式兼容
- 💡 **推荐使用Go版本** - 特别适合在线预测服务和高并发场景

**性能优势分析**：
- 🚀 **预测快37%** 的原因：
  - 更高效的 goroutine 并发模型
  - 优化的内存分配和缓存利用
  - 更好的 I/O 处理
- ⚖️ **训练持平** 说明：Go 的重构没有性能损失
- 📊 **吞吐量高59%** 表明：Go 更适合高并发预测服务

### 高维度 SIMD/BLAS 优化测试 (dim=64)

在 **dim=64** 的高维度场景下，Go 版本支持 SIMD/BLAS 优化，进一步提升性能。

#### 测试环境
- 数据集：相同真实数据集（训练集 + 测试集 341 万样本）
- 隐向量维度：dim=64
- 线程数：4

#### Go Scalar vs BLAS 模式对比

| 指标 | Scalar 模式 | BLAS 模式 (优化后) | 提升 |
|------|-------------|-------------------|------|
| **Go 训练时间** | 882s | 653s | **26% 更快** ⬆️ |
| **Go 预测时间** | 321s | 72s | **78% 更快** ⬆️⬆️ |
| **Go 模型大小** | 34M | 34M | 完全一致 |
| **预测准确性** | ✅ | ✅ | 完全一致 |

#### C++ vs Go (dim=64) 对比

| 指标 | C++ | Go (Scalar) | Go (BLAS) | 最佳 |
|------|-----|-------------|-----------|------|
| **训练时间** | 1050s | 882s (0.84x) | 653s (0.61x) | 🏆 **Go BLAS** |
| **预测时间** | 436s | 321s (0.73x) | 72s (0.16x) | 🏆 **Go BLAS** |
| **预测吞吐** | 7,841/s | 10,650/s | 47,480/s | 🏆 **Go BLAS** |

#### 结论

- 🚀 **BLAS 优化显著**：在 dim=64 下，Go BLAS 模式比 Scalar 模式快 26%（训练）和 78%（预测）
- 🏆 **Go BLAS vs C++**：训练快 **38%**，预测快 **6 倍**！
- ✅ **结果一致**：所有模式的预测结果完全一致（Mean Difference = 0.000000）
- 💡 **推荐**：高维度场景（dim≥64）建议使用 `-simd blas` 参数启用 BLAS 优化

#### 使用方法

```bash
# Scalar 模式（默认，适合低维度 dim≤32）
./bin/fm_train -dim 1,1,64 -m model.txt ...

# BLAS 模式（推荐高维度 dim≥64）
./bin/fm_train -dim 1,1,64 -simd blas -m model.txt ...
./bin/fm_predict -dim 64 -simd blas -m model.txt -out result.txt
```

## 🏗️ 项目结构

```
alphaFM-go/
├── cmd/                    # 可执行程序
│   ├── fm_train/          # 训练
│   ├── fm_predict/        # 预测
│   └── model_bin_tool/    # 工具
├── pkg/                    # 核心库
│   ├── model/             # 模型和算法
│   ├── frame/             # 多线程框架
│   ├── sample/            # 样本解析
│   ├── lock/              # 锁管理
│   ├── mem/               # 内存池
│   └── utils/             # 工具函数
├── bin/                   # 编译输出
├── docs/                  # 文档目录
│   ├── IMPLEMENTATION.md
│   ├── DELIVERY.md
│   ├── TEST_REPORT.md
│   ├── BENCHMARK.md
│   └── ...
├── scripts/               # 工具脚本
│   ├── test.sh
│   ├── demo.sh
│   ├── compare_test.sh
│   └── ...
├── benchmark_test.sh      # 基准测试脚本（主要）
└── README.md              # 项目说明
```

## 📖 文档

- [IMPLEMENTATION.md](docs/IMPLEMENTATION.md) - 详细实现说明
- [DELIVERY.md](docs/DELIVERY.md) - 项目交付文档
- [TEST_REPORT.md](docs/TEST_REPORT.md) - 测试报告
- [BENCHMARK.md](docs/BENCHMARK.md) - 基准测试文档
- [SIMD_GUIDE.md](docs/SIMD_GUIDE.md) - **SIMD/BLAS 优化指南（重要）**
- [STREAMING_OPTIMIZATION.md](docs/STREAMING_OPTIMIZATION.md) - 流式处理优化
- [MULTITHREAD_DETERMINISM_ANALYSIS.md](docs/MULTITHREAD_DETERMINISM_ANALYSIS.md) - 多线程确定性分析
- [CPP_GO_DETAILED_COMPARISON_REPORT.md](docs/CPP_GO_DETAILED_COMPARISON_REPORT.md) - C++与Go详细对比报告
- [BINARY_MODEL_IMPLEMENTATION.md](docs/BINARY_MODEL_IMPLEMENTATION.md) - 二进制模型实现文档
- [原C++版本文档](https://github.com/CastellanZhang/alphaFM)

## 🧪 测试

### 基础测试

```bash
# 快速功能测试（小数据集）
./scripts/test.sh

# 对比C++和Go版本（小数据集）
./scripts/compare_test.sh

# 演示
./scripts/demo.sh
```

### 基准测试（大规模真实数据）

**完整基准测试**（包含训练+预测+性能对比）：

```bash
# 运行完整基准测试
./benchmark_test.sh

# 测试特点：
# - 逐文件流式处理（低内存占用）
# - C++ vs Go 性能对比
# - 训练和预测时间、内存监控
# - 预测结果准确性验证
# - 生成详细性能报告
```

**测试输出**：
- `benchmark_results/performance_report.txt` - 完整性能报告
- `benchmark_results/cpp_model.bin` / `go_model.bin` - 训练的模型
- `benchmark_results/cpp_prediction.txt` / `go_prediction.txt` - 预测结果
- 详细的训练和预测日志

**快速数据预处理测试**：

```bash
# 测试数据格式转换和流式处理
./scripts/test_data_preprocessing.sh
```

详细说明请参考：
- [BENCHMARK.md](docs/BENCHMARK.md) - 基准测试完整文档
- [STREAMING_OPTIMIZATION.md](docs/STREAMING_OPTIMIZATION.md) - 流式处理优化说明

## 🤝 与C++版本的兼容性

- ✅ **文本模型格式完全兼容** - 可以互换使用
- ✅ **命令行参数完全兼容** - 参数名称和含义一致
- ✅ **数据格式完全兼容** - 样本格式相同
- ✅ **算法完全一致** - 数值结果相同

## 🔧 开发

```bash
# 格式化代码
make fmt

# 清理编译产物
make clean

# 重新编译
make
```

## 📝 License

MIT License - 详见 [LICENSE](LICENSE) 文件

## 🙏 致谢

本项目基于 [alphaFM](https://github.com/CastellanZhang/alphaFM) 的Go语言重构实现，感谢原作者的开源贡献。

## 📧 联系方式

- 项目位置: `/data/xiongle/alphaFM-go`
- 原C++版本: `/data/xiongle/alphaFM`

---

⭐ 如果这个项目对你有帮助，欢迎 Star！

