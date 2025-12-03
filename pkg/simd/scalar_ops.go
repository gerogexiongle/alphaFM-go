package simd

// ScalarOps 标量运算实现（无SIMD优化）
type ScalarOps struct{}

// NewScalarOps 创建标量运算实例
func NewScalarOps() *ScalarOps {
	return &ScalarOps{}
}

// DotProduct 计算点积（标量实现）
func (s *ScalarOps) DotProduct(v1, v2 []float64) float64 {
	sum := 0.0
	n := len(v1)
	if len(v2) < n {
		n = len(v2)
	}
	for i := 0; i < n; i++ {
		sum += v1[i] * v2[i]
	}
	return sum
}

// DotProductScaled 计算缩放点积（标量实现）
func (s *ScalarOps) DotProductScaled(v1, v2 []float64, scale float64) float64 {
	sum := 0.0
	n := len(v1)
	if len(v2) < n {
		n = len(v2)
	}
	for i := 0; i < n; i++ {
		sum += v1[i] * v2[i]
	}
	return sum * scale
}

// SumSquares 计算平方和（标量实现）
func (s *ScalarOps) SumSquares(v []float64) float64 {
	sum := 0.0
	for i := 0; i < len(v); i++ {
		sum += v[i] * v[i]
	}
	return sum
}

// ScaledSumSquares 计算缩放后的平方和（标量实现）
func (s *ScalarOps) ScaledSumSquares(v []float64, scale float64) float64 {
	sum := 0.0
	scale2 := scale * scale
	for i := 0; i < len(v); i++ {
		sum += v[i] * v[i] * scale2
	}
	return sum
}

// Axpy 计算 y = alpha*x + y（标量实现）
func (s *ScalarOps) Axpy(alpha float64, x, y []float64) {
	n := len(x)
	if len(y) < n {
		n = len(y)
	}
	for i := 0; i < n; i++ {
		y[i] += alpha * x[i]
	}
}

// Scale 计算 x = alpha*x（标量实现）
func (s *ScalarOps) Scale(alpha float64, x []float64) {
	for i := 0; i < len(x); i++ {
		x[i] *= alpha
	}
}

// Sum 计算向量元素之和（标量实现）
func (s *ScalarOps) Sum(v []float64) float64 {
	sum := 0.0
	for i := 0; i < len(v); i++ {
		sum += v[i]
	}
	return sum
}

// Type 返回实现类型
func (s *ScalarOps) Type() VectorOpsType {
	return VectorOpsScalar
}

// Name 返回实现名称
func (s *ScalarOps) Name() string {
	return "Scalar (No SIMD)"
}
