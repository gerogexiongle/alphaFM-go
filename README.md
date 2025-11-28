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
- 🚀 **预测速度快40%** - 118秒 vs 71秒
- ⚡ **预测吞吐高66%** - 28,954/s vs 48,150/s  
- ⚖️ **训练速度相同** - 317秒 vs 318秒（0.3%差异）
- ✅ **结果完全一致** - 341万预测零误差

**适用场景**：
- 🎯 **在线预测服务** - 预测速度快40%，响应更快
- 📊 **高并发场景** - goroutine轻量级，支持更多并发
- 🔧 **快速迭代** - 代码简洁，易于维护和扩展
- 🌐 **云原生部署** - 容器友好，资源利用率高

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

## 🚀 快速开始

### 安装依赖

```bash
# 需要 Go 1.18 或更高版本
go version
```

### 编译

```bash
cd /data/xiongle/alphaFM-go
make
```

编译后在 `bin/` 目录生成3个可执行文件：
- `fm_train` - 训练程序
- `fm_predict` - 预测程序  
- `model_bin_tool` - 模型工具

### 快速测试

```bash
# 运行测试脚本
./test.sh

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

### 基准测试结果（真实数据集）

**测试数据集**：
- 训练数据：1,640万条样本，107个分区文件
- 测试数据：341万条样本，200+个分区文件
- 测试参数：`-dim 1,1,4 -core 4`

**训练性能**：

| 指标 | C++版本 | Go版本 | 对比 |
|------|---------|--------|------|
| **训练时间** | 317秒 | 318秒 | ⚖️ **持平** (0.3%差异) |
| **吞吐量** | 51,735样本/秒 | 51,572样本/秒 | 几乎相同 |
| 模型大小 | 2.8 MB | 5.6 MB | Go大2倍 |

**预测性能**：

| 指标 | C++版本 | Go版本 | 对比 |
|------|---------|--------|------|
| **预测时间** | 118秒 | 71秒 | 🚀 **Go快40%** |
| **吞吐量** | 28,954样本/秒 | 48,150样本/秒 | 🎯 **Go快66%** |
| 预测样本数 | 3,418,591 | 3,418,591 | 完全一致 |

**结果准确性**：

| 指标 | 值 | 说明 |
|------|-----|------|
| 平均差异 | 0.000000 | 完全一致 |
| 标准差 | 0.000000 | 无偏差 |
| 平均绝对误差 | 0.000000 | 341万预测完全相同 |

### 综合评价

| 维度 | C++ | Go | 胜者 |
|------|-----|-----|------|
| **训练速度** | 317秒 | 318秒 | ⚖️ 平手 |
| **预测速度** | 118秒 | 71秒 | 🏆 **Go** (快40%) |
| **预测吞吐** | 28,954/s | 48,150/s | 🏆 **Go** (快66%) |
| **结果准确性** | ✅ | ✅ | ⚖️ 完全一致 |
| **模型大小** | 2.8MB | 5.6MB | 🏆 C++ |
| **代码可维护性** | - | - | 🏆 Go |

**测试结论**：
- ✅ **训练性能持平** - Go版本与C++版本训练速度几乎相同
- 🚀 **预测性能优异** - Go版本预测速度快40%，适合在线服务
- 🎯 **结果完全一致** - 341万预测结果与C++版本完全相同，零误差
- 💡 **推荐使用Go版本** - 特别是在需要高性能预测的生产环境

**测试环境**：
- CPU: Intel Xeon
- 内存: 充足
- 数据：生产级真实数据集
- 优化：流式处理，逐文件加载（低内存占用）

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
└── bin/                   # 编译输出
```

## 📖 文档

- [IMPLEMENTATION.md](IMPLEMENTATION.md) - 详细实现说明
- [DELIVERY.md](DELIVERY.md) - 项目交付文档
- [TEST_REPORT.md](TEST_REPORT.md) - 测试报告
- [BENCHMARK.md](BENCHMARK.md) - 基准测试文档（新）
- [STREAMING_OPTIMIZATION.md](STREAMING_OPTIMIZATION.md) - 流式处理优化（新）
- [原C++版本文档](https://github.com/CastellanZhang/alphaFM)

## 🧪 测试

### 基础测试

```bash
# 快速功能测试（小数据集）
./test.sh

# 对比C++和Go版本（小数据集）
./compare_test.sh

# 演示
./demo.sh
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
./test_data_preprocessing.sh
```

详细说明请参考：
- [BENCHMARK.md](BENCHMARK.md) - 基准测试完整文档
- [STREAMING_OPTIMIZATION.md](STREAMING_OPTIMIZATION.md) - 流式处理优化说明

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

