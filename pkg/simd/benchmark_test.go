package simd

import (
	"testing"
)

// Go benchmark 标准测试
func BenchmarkScalarDotProduct(b *testing.B) {
	ops := NewScalarOps()
	v1 := make([]float64, 8)
	v2 := make([]float64, 8)
	for i := 0; i < 8; i++ {
		v1[i] = float64(i) * 0.1
		v2[i] = float64(i) * 0.2
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ops.DotProduct(v1, v2)
	}
}

func BenchmarkBLASDotProduct(b *testing.B) {
	ops, err := NewBLASOps()
	if err != nil {
		b.Skipf("BLAS not available: %v", err)
		return
	}
	
	v1 := make([]float64, 8)
	v2 := make([]float64, 8)
	for i := 0; i < 8; i++ {
		v1[i] = float64(i) * 0.1
		v2[i] = float64(i) * 0.2
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ops.DotProduct(v1, v2)
	}
}

func BenchmarkScalarSumSquares(b *testing.B) {
	ops := NewScalarOps()
	v := make([]float64, 8)
	for i := 0; i < 8; i++ {
		v[i] = float64(i) * 0.1
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ops.SumSquares(v)
	}
}

func BenchmarkBLASSumSquares(b *testing.B) {
	ops, err := NewBLASOps()
	if err != nil {
		b.Skipf("BLAS not available: %v", err)
		return
	}
	
	v := make([]float64, 8)
	for i := 0; i < 8; i++ {
		v[i] = float64(i) * 0.1
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ops.SumSquares(v)
	}
}

