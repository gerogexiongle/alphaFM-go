package simd

import (
	"fmt"
)

// VectorOpsType SIMD实现类型
type VectorOpsType int

const (
	// VectorOpsScalar 标量运算（默认，无SIMD）
	VectorOpsScalar VectorOpsType = iota
	// VectorOpsBLAS BLAS库优化
	VectorOpsBLAS
	// VectorOpsAVX2 AVX2手写汇编（未来扩展）
	VectorOpsAVX2
)

// String 返回类型名称
func (t VectorOpsType) String() string {
	switch t {
	case VectorOpsScalar:
		return "scalar"
	case VectorOpsBLAS:
		return "blas"
	case VectorOpsAVX2:
		return "avx2"
	default:
		return "unknown"
	}
}

// ParseVectorOpsType 从字符串解析类型
func ParseVectorOpsType(s string) (VectorOpsType, error) {
	switch s {
	case "scalar", "":
		return VectorOpsScalar, nil
	case "blas":
		return VectorOpsBLAS, nil
	case "avx2":
		return VectorOpsAVX2, nil
	default:
		return VectorOpsScalar, fmt.Errorf("unknown vector ops type: %s (available: scalar, blas, avx2)", s)
	}
}

// VectorOps 向量运算接口
type VectorOps interface {
	// DotProduct 计算两个向量的点积: v1 · v2
	DotProduct(v1, v2 []float64) float64
	
	// DotProductScaled 计算缩放点积: (v1 · v2) * scale
	DotProductScaled(v1, v2 []float64, scale float64) float64
	
	// SumSquares 计算向量元素的平方和: Σ(vi^2)
	SumSquares(v []float64) float64
	
	// ScaledSumSquares 计算缩放后的平方和: Σ((vi * scale)^2)
	ScaledSumSquares(v []float64, scale float64) float64
	
	// Axpy 计算 y = alpha*x + y (BLAS Level 1)
	Axpy(alpha float64, x, y []float64)
	
	// Scale 计算 x = alpha*x (BLAS Level 1)
	Scale(alpha float64, x []float64)
	
	// Sum 计算向量元素之和: Σ(vi)
	Sum(v []float64) float64
	
	// Type 返回实现类型
	Type() VectorOpsType
	
	// Name 返回实现名称
	Name() string
}

// NewVectorOps 创建向量运算实现
func NewVectorOps(opsType VectorOpsType) (VectorOps, error) {
	switch opsType {
	case VectorOpsScalar:
		return NewScalarOps(), nil
	case VectorOpsBLAS:
		return NewBLASOps()
	case VectorOpsAVX2:
		return nil, fmt.Errorf("AVX2 implementation not yet available")
	default:
		return nil, fmt.Errorf("unknown vector ops type: %v", opsType)
	}
}

// globalVectorOps 全局向量运算实例（默认为标量运算）
var globalVectorOps VectorOps = NewScalarOps()

// GetGlobalVectorOps 获取全局向量运算实例
func GetGlobalVectorOps() VectorOps {
	return globalVectorOps
}

// SetGlobalVectorOps 设置全局向量运算实例
func SetGlobalVectorOps(ops VectorOps) {
	globalVectorOps = ops
}

// InitGlobalVectorOps 初始化全局向量运算实例
func InitGlobalVectorOps(opsType VectorOpsType) error {
	ops, err := NewVectorOps(opsType)
	if err != nil {
		return err
	}
	globalVectorOps = ops
	return nil
}
