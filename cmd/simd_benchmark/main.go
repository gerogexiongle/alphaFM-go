package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/xiongle/alphaFM-go/pkg/simd"
)

func main() {
	vectorSize := flag.Int("size", 8, "vector size for benchmark")
	iterations := flag.Int("iter", 100000, "number of iterations")
	
	flag.Parse()

	config := &simd.BenchmarkConfig{
		VectorSize: *vectorSize,
		Iterations: *iterations,
	}

	fmt.Printf("Running SIMD Performance Benchmark...\n")
	fmt.Printf("Vector Size: %d, Iterations: %d\n", config.VectorSize, config.Iterations)

	results, err := simd.BenchmarkVectorOps(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Benchmark failed: %v\n", err)
		os.Exit(1)
	}

	simd.PrintBenchmarkResults(results)
	
	// 分析结果
	if len(results) > 1 {
		fmt.Println("Analysis:")
		fmt.Println("- scalar: baseline implementation (no SIMD)")
		
		for _, r := range results {
			if r.OpsType != simd.VectorOpsScalar {
				improvement := (r.Speedup - 1.0) * 100
				fmt.Printf("- %s: %.1f%% %s than scalar\n",
					r.OpsType.String(),
					improvement,
					map[bool]string{true: "faster", false: "slower"}[improvement > 0])
			}
		}
		
		fmt.Println("\nRecommendation:")
		bestResult := results[0]
		for _, r := range results {
			if r.Speedup > bestResult.Speedup {
				bestResult = r
			}
		}
		
		if bestResult.OpsType == simd.VectorOpsScalar {
			fmt.Println("Use -simd=scalar (default) for best compatibility")
		} else {
			fmt.Printf("Use -simd=%s for %.2fx speedup over scalar\n", 
				bestResult.OpsType.String(), bestResult.Speedup)
		}
	}
}
