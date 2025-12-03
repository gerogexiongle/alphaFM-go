package simd

import (
	"testing"
)

func TestScalarOps(t *testing.T) {
	ops := NewScalarOps()
	
	v1 := []float64{1.0, 2.0, 3.0, 4.0}
	v2 := []float64{2.0, 3.0, 4.0, 5.0}
	
	// 测试点积: 1*2 + 2*3 + 3*4 + 4*5 = 2 + 6 + 12 + 20 = 40
	dot := ops.DotProduct(v1, v2)
	expected := 40.0
	if dot != expected {
		t.Errorf("DotProduct failed: got %f, want %f", dot, expected)
	}
	
	// 测试缩放点积
	scaled := ops.DotProductScaled(v1, v2, 2.0)
	if scaled != expected*2.0 {
		t.Errorf("DotProductScaled failed: got %f, want %f", scaled, expected*2.0)
	}
	
	// 测试平方和: 1^2 + 2^2 + 3^2 + 4^2 = 1 + 4 + 9 + 16 = 30
	sumSq := ops.SumSquares(v1)
	expectedSumSq := 30.0
	if sumSq != expectedSumSq {
		t.Errorf("SumSquares failed: got %f, want %f", sumSq, expectedSumSq)
	}
}

func TestBLASOps(t *testing.T) {
	ops, err := NewBLASOps()
	if err != nil {
		t.Skipf("BLAS not available: %v", err)
		return
	}
	
	v1 := []float64{1.0, 2.0, 3.0, 4.0}
	v2 := []float64{2.0, 3.0, 4.0, 5.0}
	
	// 测试点积
	dot := ops.DotProduct(v1, v2)
	expected := 40.0
	if dot != expected {
		t.Errorf("BLAS DotProduct failed: got %f, want %f", dot, expected)
	}
	
	// 测试平方和
	sumSq := ops.SumSquares(v1)
	expectedSumSq := 30.0
	if sumSq != expectedSumSq {
		t.Errorf("BLAS SumSquares failed: got %f, want %f", sumSq, expectedSumSq)
	}
}

func TestVectorOpsConsistency(t *testing.T) {
	// 测试不同实现的结果一致性
	v1 := []float64{1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0, 8.0}
	v2 := []float64{0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8}
	
	scalarOps := NewScalarOps()
	scalarDot := scalarOps.DotProduct(v1, v2)
	scalarSumSq := scalarOps.SumSquares(v1)
	
	blasOps, err := NewBLASOps()
	if err != nil {
		t.Skipf("BLAS not available: %v", err)
		return
	}
	
	blasDot := blasOps.DotProduct(v1, v2)
	blasSumSq := blasOps.SumSquares(v1)
	
	// 允许浮点误差
	const epsilon = 1e-10
	
	if diff := scalarDot - blasDot; diff > epsilon || diff < -epsilon {
		t.Errorf("DotProduct inconsistency: scalar=%f, blas=%f, diff=%f", scalarDot, blasDot, diff)
	}
	
	if diff := scalarSumSq - blasSumSq; diff > epsilon || diff < -epsilon {
		t.Errorf("SumSquares inconsistency: scalar=%f, blas=%f, diff=%f", scalarSumSq, blasSumSq, diff)
	}
}

func TestParseVectorOpsType(t *testing.T) {
	tests := []struct {
		input    string
		expected VectorOpsType
		hasError bool
	}{
		{"scalar", VectorOpsScalar, false},
		{"", VectorOpsScalar, false},
		{"blas", VectorOpsBLAS, false},
		{"avx2", VectorOpsAVX2, false},
		{"invalid", VectorOpsScalar, true},
	}
	
	for _, tt := range tests {
		result, err := ParseVectorOpsType(tt.input)
		if tt.hasError && err == nil {
			t.Errorf("Expected error for input %q, got none", tt.input)
		}
		if !tt.hasError && err != nil {
			t.Errorf("Unexpected error for input %q: %v", tt.input, err)
		}
		if !tt.hasError && result != tt.expected {
			t.Errorf("For input %q: got %v, want %v", tt.input, result, tt.expected)
		}
	}
}

