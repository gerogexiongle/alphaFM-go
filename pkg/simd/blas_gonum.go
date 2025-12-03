// +build !noblas

package simd

import (
	"gonum.org/v1/gonum/blas/blas64"
)

// gonumBLASImpl gonum BLAS实现
type gonumBLASImpl struct{}

func (g *gonumBLASImpl) ddot(n int, x []float64, incx int, y []float64, incy int) float64 {
	// gonum blas64 使用 Vector 类型
	xVec := blas64.Vector{N: n, Data: x, Inc: incx}
	yVec := blas64.Vector{N: n, Data: y, Inc: incy}
	return blas64.Dot(xVec, yVec)
}

func (g *gonumBLASImpl) daxpy(n int, alpha float64, x []float64, incx int, y []float64, incy int) {
	// y = alpha*x + y
	xVec := blas64.Vector{N: n, Data: x, Inc: incx}
	yVec := blas64.Vector{N: n, Data: y, Inc: incy}
	blas64.Axpy(alpha, xVec, yVec)
}

func (g *gonumBLASImpl) dscal(n int, alpha float64, x []float64, incx int) {
	// x = alpha*x
	xVec := blas64.Vector{N: n, Data: x, Inc: incx}
	blas64.Scal(alpha, xVec)
}

func (g *gonumBLASImpl) dasum(n int, x []float64, incx int) float64 {
	// sum of absolute values
	xVec := blas64.Vector{N: n, Data: x, Inc: incx}
	return blas64.Asum(xVec)
}

// initBLAS 初始化BLAS（gonum版本）
func initBLAS() (blasImpl, error) {
	return &gonumBLASImpl{}, nil
}
