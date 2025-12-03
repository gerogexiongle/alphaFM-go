package simd

import (
	"fmt"
	"time"
)

// BenchmarkConfig 基准测试配置
type BenchmarkConfig struct {
	VectorSize int
	Iterations int
}

// DefaultBenchmarkConfig 默认基准测试配置
func DefaultBenchmarkConfig() *BenchmarkConfig {
	return &BenchmarkConfig{
		VectorSize: 8,      // 模拟FM的factor_num=8
		Iterations: 100000, // 10万次迭代
	}
}

// BenchmarkResult 基准测试结果
type BenchmarkResult struct {
	Name        string
	OpsType     VectorOpsType
	Duration    time.Duration
	OpsPerSec   float64
	Speedup     float64 // 相对于标量版本的加速比
}

// BenchmarkVectorOps 对比不同SIMD实现的性能
func BenchmarkVectorOps(config *BenchmarkConfig) ([]*BenchmarkResult, error) {
	if config == nil {
		config = DefaultBenchmarkConfig()
	}

	results := make([]*BenchmarkResult, 0)

	// 准备测试数据
	v1 := make([]float64, config.VectorSize)
	v2 := make([]float64, config.VectorSize)
	for i := 0; i < config.VectorSize; i++ {
		v1[i] = float64(i) * 0.1
		v2[i] = float64(i) * 0.2
	}

	// 测试所有可用的实现
	opsTypes := []VectorOpsType{VectorOpsScalar, VectorOpsBLAS}
	var baselineOpsPerSec float64

	for _, opsType := range opsTypes {
		ops, err := NewVectorOps(opsType)
		if err != nil {
			fmt.Printf("Skipping %s: %v\n", opsType.String(), err)
			continue
		}

		// 预热
		for i := 0; i < 1000; i++ {
			ops.DotProduct(v1, v2)
		}

		// 正式测试
		start := time.Now()
		for i := 0; i < config.Iterations; i++ {
			_ = ops.DotProduct(v1, v2)
			_ = ops.SumSquares(v1)
		}
		duration := time.Since(start)

		opsPerSec := float64(config.Iterations*2) / duration.Seconds()
		
		result := &BenchmarkResult{
			Name:      ops.Name(),
			OpsType:   opsType,
			Duration:  duration,
			OpsPerSec: opsPerSec,
		}

		// 计算加速比（相对于标量版本）
		if opsType == VectorOpsScalar {
			baselineOpsPerSec = opsPerSec
			result.Speedup = 1.0
		} else {
			result.Speedup = opsPerSec / baselineOpsPerSec
		}

		results = append(results, result)
	}

	return results, nil
}

// PrintBenchmarkResults 打印基准测试结果
func PrintBenchmarkResults(results []*BenchmarkResult) {
	fmt.Println("\n========== SIMD Performance Benchmark ==========")
	fmt.Printf("%-30s %-15s %-15s %-10s\n", "Implementation", "Duration", "Ops/sec", "Speedup")
	fmt.Println(string(make([]byte, 75)))
	
	for _, r := range results {
		fmt.Printf("%-30s %-15s %-15.2f %-10.2fx\n",
			r.Name,
			r.Duration.String(),
			r.OpsPerSec,
			r.Speedup,
		)
	}
	fmt.Println("================================================\n")
}
