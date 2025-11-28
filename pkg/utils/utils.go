package utils

import (
	"math"
	"math/rand"
	"strings"
)

const kPrecision = 0.0000000001

// SplitString 分割字符串
func SplitString(line string, delimiter byte) []string {
	return strings.Split(line, string(delimiter))
}

// Sgn 符号函数
func Sgn(x float64) int {
	if x > kPrecision {
		return 1
	}
	return -1
}

// Uniform 生成均匀分布随机数
func Uniform() float64 {
	return rand.Float64()
}

// Gaussian 生成标准正态分布随机数 (Box-Muller变换的极坐标形式)
func Gaussian() float64 {
	var u, v, x, y, Q float64
	for {
		for {
			u = Uniform()
			if u != 0.0 {
				break
			}
		}
		v = 1.7156 * (Uniform() - 0.5)
		x = u - 0.449871
		y = math.Abs(v) + 0.386595
		Q = x*x + y*(0.19600*y-0.25472*x)
		if Q < 0.27597 || (Q <= 0.27846 && v*v <= -4.0*u*u*math.Log(u)) {
			break
		}
	}
	return v / u
}

// GaussianWithParams 生成指定均值和标准差的正态分布随机数
func GaussianWithParams(mean, stdev float64) float64 {
	if stdev == 0.0 {
		return mean
	}
	return mean + stdev*Gaussian()
}

// Abs 绝对值
func Abs(x float64) float64 {
	return math.Abs(x)
}

