package simd

import (
	"fmt"
	"runtime"
)

// BLASOps BLAS库优化实现
// 注意：需要安装 gonum 库
type BLASOps struct {
	available bool
	impl      blasImpl
}

// blasImpl BLAS实现接口（用于解耦实际BLAS库）
type blasImpl interface {
	ddot(n int, x []float64, incx int, y []float64, incy int) float64
	daxpy(n int, alpha float64, x []float64, incx int, y []float64, incy int)
	dscal(n int, alpha float64, x []float64, incx int)
	dasum(n int, x []float64, incx int) float64
}

// NewBLASOps 创建BLAS运算实例
func NewBLASOps() (*BLASOps, error) {
	ops := &BLASOps{}
	
	// 尝试初始化BLAS
	impl, err := initBLAS()
	if err != nil {
		return nil, fmt.Errorf("BLAS not available: %v", err)
	}
	
	ops.available = true
	ops.impl = impl
	return ops, nil
}

// DotProduct 计算点积（BLAS实现）
func (b *BLASOps) DotProduct(v1, v2 []float64) float64 {
	if !b.available || len(v1) == 0 || len(v2) == 0 {
		// 降级到标量实现
		return (&ScalarOps{}).DotProduct(v1, v2)
	}
	
	n := len(v1)
	if len(v2) < n {
		n = len(v2)
	}
	
	return b.impl.ddot(n, v1, 1, v2, 1)
}

// DotProductScaled 计算缩放点积（BLAS实现）
func (b *BLASOps) DotProductScaled(v1, v2 []float64, scale float64) float64 {
	return b.DotProduct(v1, v2) * scale
}

// SumSquares 计算平方和（BLAS实现）
func (b *BLASOps) SumSquares(v []float64) float64 {
	if !b.available || len(v) == 0 {
		return (&ScalarOps{}).SumSquares(v)
	}
	
	// 使用 ddot(v, v) 计算 v·v
	return b.impl.ddot(len(v), v, 1, v, 1)
}

// ScaledSumSquares 计算缩放后的平方和
func (b *BLASOps) ScaledSumSquares(v []float64, scale float64) float64 {
	return b.SumSquares(v) * scale * scale
}

// Axpy 计算 y = alpha*x + y（BLAS实现）
func (b *BLASOps) Axpy(alpha float64, x, y []float64) {
	if !b.available || len(x) == 0 || len(y) == 0 {
		(&ScalarOps{}).Axpy(alpha, x, y)
		return
	}
	n := len(x)
	if len(y) < n {
		n = len(y)
	}
	b.impl.daxpy(n, alpha, x, 1, y, 1)
}

// Scale 计算 x = alpha*x（BLAS实现）
func (b *BLASOps) Scale(alpha float64, x []float64) {
	if !b.available || len(x) == 0 {
		(&ScalarOps{}).Scale(alpha, x)
		return
	}
	b.impl.dscal(len(x), alpha, x, 1)
}

// Sum 计算向量元素之和（BLAS实现，使用 dasum 绝对值和的变体）
func (b *BLASOps) Sum(v []float64) float64 {
	// 注意：BLAS 没有直接的 sum，使用标量实现
	// dasum 计算的是绝对值之和，不能直接用
	return (&ScalarOps{}).Sum(v)
}

// Type 返回实现类型
func (b *BLASOps) Type() VectorOpsType {
	return VectorOpsBLAS
}

// Name 返回实现名称
func (b *BLASOps) Name() string {
	if !b.available {
		return "BLAS (unavailable, using scalar)"
	}
	return fmt.Sprintf("BLAS (gonum, %s)", runtime.GOARCH)
}

// IsBLASAvailable 检查BLAS是否可用
func IsBLASAvailable() bool {
	_, err := initBLAS()
	return err == nil
}
