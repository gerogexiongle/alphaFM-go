// +build noblas

package simd

import (
	"fmt"
)

// initBLAS 初始化BLAS（无BLAS版本）
func initBLAS() (blasImpl, error) {
	return nil, fmt.Errorf("BLAS support not compiled in (use 'go build' without -tags=noblas)")
}
