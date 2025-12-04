# SIMD/BLAS 优化指南

## 📋 概述

alphaFM-go 支持 SIMD/BLAS 向量化优化，可显著提升高维度场景下的训练和预测性能。

### 核心特性
- ✅ **灵活开关**：通过 `-simd` 参数控制
- ✅ **向后兼容**：默认使用标量实现，行为不变
- ✅ **自动降级**：BLAS 不可用时自动降级到标量
- ✅ **结果一致**：所有实现结果完全相同

---

## 🚀 快速开始

### 编译

```bash
make clean && make
```

### 使用方法

```bash
# 标量模式（默认）
cat data.txt | ./bin/fm_train -m model.txt -dim 1,1,8

# BLAS 模式（高维度推荐）
cat data.txt | ./bin/fm_train -m model.txt -dim 1,1,64 -simd blas
cat data.txt | ./bin/fm_predict -m model.txt -dim 64 -out pred.txt -simd blas
```

### 性能测试

```bash
# 测试不同维度的最优配置
./bin/simd_benchmark -size 8    # 典型 FM
./bin/simd_benchmark -size 64   # 高维度
```

---

## 📊 性能测试结果 (dim=64)

### Go Scalar vs BLAS 模式

| 指标 | Scalar 模式 | BLAS 模式 | 提升 |
|------|-------------|-----------|------|
| **训练时间** | 882s | 653s | **26% 更快** ⬆️ |
| **预测时间** | 321s | 72s | **78% 更快** ⬆️⬆️ |
| **预测准确性** | ✅ | ✅ | 完全一致 |

### C++ vs Go (dim=64) 完整对比

| 指标 | C++ | Go (Scalar) | Go (BLAS) | 最佳 |
|------|-----|-------------|-----------|------|
| **训练** | 1050s | 882s (0.84x) | 653s (0.61x) | 🏆 Go BLAS |
| **预测** | 436s | 321s (0.73x) | 72s (0.16x) | 🏆 Go BLAS |
| **吞吐** | 7,841/s | 10,650/s | 47,480/s | 🏆 Go BLAS |

**结论**：Go BLAS 模式训练快 **38%**，预测快 **6 倍**！

---

## 🎯 使用建议

### 性能对比表

| 维度 (factor_num) | 推荐模式 | 说明 |
|-------------------|---------|------|
| **≤ 16** | `scalar` (默认) | 标量实现最优 |
| **16-32** | `scalar` | 接近临界点 |
| **≥ 64** | `blas` | BLAS 快 27%+ |
| **≥ 128** | `blas` | BLAS 快 50%+ |

### 快速决策

```bash
# 典型 FM (dim=4-16)：使用默认配置
./bin/fm_train -m model.txt -dim 1,1,8 -core 4

# 高维度 (dim≥64)：启用 BLAS
./bin/fm_train -m model.txt -dim 1,1,64 -core 4 -simd blas

# 不确定？先测试
./bin/simd_benchmark -size <你的factor_num>
```

---

## 🔧 技术架构

```
应用层 (Train/Predict)
    ↓
VectorOps 接口
    ↓
┌─────────────┬─────────────┐
│ ScalarOps   │  BLASOps    │
│ (默认)      │  (gonum)    │
└─────────────┴─────────────┘
```

### 优化的计算

```go
// FM 二阶交互项（最耗时）
sumF = Σ(vi[f] * xi)      // → DotProduct / Axpy
sumSqr = Σ((vi[f] * xi)²) // → SumSquares / ScaledSumSquares
```

### 核心文件

| 文件 | 功能 |
|------|------|
| `pkg/simd/simd.go` | 接口定义 |
| `pkg/simd/scalar_ops.go` | 标量实现 |
| `pkg/simd/blas_ops.go` | BLAS 实现 |
| `cmd/simd_benchmark/` | 性能测试工具 |

---

## ❓ 常见问题

### Q: BLAS 初始化失败？

```bash
# 下载依赖
go mod download

# 或使用标量模式（无依赖）
./bin/fm_train -simd scalar ...
```

### Q: 编译不包含 BLAS？

```bash
# 使用 noblas 标签
go build -tags noblas -o bin/fm_train cmd/fm_train/main.go
```

### Q: 如何验证结果一致性？

```bash
# 对比两种模式的预测结果
diff pred_scalar.txt pred_blas.txt && echo "✅ 完全一致"
```

---

## 📈 性能优化优先级

1. 🥇 **多线程** `-core N` — 最有效，线性扩展
2. 🥈 **SIMD/BLAS** `-simd blas` — 高维度场景有效
3. 🥉 **数据优化** — 特征工程、减少稀疏度

---

## 📝 总结

| 场景 | 配置 | 效果 |
|------|------|------|
| **典型 FM (dim≤16)** | 默认 (scalar) | ✅ 最优 |
| **高维度 (dim≥64)** | `-simd blas` | 🚀 训练快26%，预测快78% |
| **生产环境** | 根据维度选择 | ✅ 稳定可靠 |

**一句话**：低维度用默认，高维度用 BLAS！🎯
