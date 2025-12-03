package model

import (
	"fmt"
	"math"
	"sync"

	"github.com/xiongle/alphaFM-go/pkg/lock"
	"github.com/xiongle/alphaFM-go/pkg/sample"
	"github.com/xiongle/alphaFM-go/pkg/simd"
	"github.com/xiongle/alphaFM-go/pkg/utils"
)

// TrainerOption 训练选项
type TrainerOption struct {
	ModelPath           string
	ModelFormat         string
	InitModelPath       string
	InitialModelFormat  string
	ModelNumberType     string
	InitMean            float64
	InitStdev           float64
	WAlpha              float64
	WBeta               float64
	WL1                 float64
	WL2                 float64
	VAlpha              float64
	VBeta               float64
	VL1                 float64
	VL2                 float64
	ThreadsNum          int
	FactorNum           int
	K0                  bool
	K1                  bool
	BInit               bool
	ForceVSparse        bool
	SIMDType            simd.VectorOpsType // SIMD优化类型
}

// NewTrainerOption 创建默认训练选项
func NewTrainerOption() *TrainerOption {
	return &TrainerOption{
		K0:                 true,
		K1:                 true,
		FactorNum:          8,
		InitMean:           0.0,
		InitStdev:          0.1,
		WAlpha:             0.05,
		WBeta:              1.0,
		WL1:                0.1,
		WL2:                5.0,
		VAlpha:             0.05,
		VBeta:              1.0,
		VL1:                0.1,
		VL2:                5.0,
		ModelFormat:        "txt",
		InitialModelFormat: "txt",
		ThreadsNum:         1,
		BInit:              false,
		ForceVSparse:       false,
		ModelNumberType:    "double",
		SIMDType:           simd.VectorOpsScalar, // 默认不使用SIMD
	}
}

// FTRLTrainer FTRL训练器
type FTRLTrainer struct {
	model        *FTRLModel
	lockPool     *lock.LockPool
	opt          *TrainerOption
	simdOps      simd.VectorOps // SIMD运算实例
	useSIMD      bool           // 是否使用SIMD
	// 预分配的缓冲区（减少内存分配）
	sumBuf       []float64      // sum 缓冲区 [factorNum]
	viRowBuf     []float64      // Vi 行缓冲区 [factorNum]
}

// NewFTRLTrainer 创建训练器
func NewFTRLTrainer(opt *TrainerOption) *FTRLTrainer {
	t := &FTRLTrainer{
		model:    NewFTRLModel(opt.FactorNum, opt.InitMean, opt.InitStdev),
		lockPool: lock.NewLockPool(),
		opt:      opt,
		// 预分配缓冲区
		sumBuf:   make([]float64, opt.FactorNum),
		viRowBuf: make([]float64, opt.FactorNum),
	}
	
	// 初始化SIMD
	if opt.SIMDType != simd.VectorOpsScalar {
		ops, err := simd.NewVectorOps(opt.SIMDType)
		if err != nil {
			fmt.Printf("Warning: SIMD initialization failed, falling back to scalar: %v\n", err)
			t.simdOps = simd.NewScalarOps()
			t.useSIMD = false
		} else {
			t.simdOps = ops
			t.useSIMD = true
			fmt.Printf("SIMD enabled: %s\n", ops.Name())
		}
	} else {
		t.simdOps = simd.NewScalarOps()
		t.useSIMD = false
	}
	
	return t
}

// RunTask 处理一批数据
func (t *FTRLTrainer) RunTask(dataBuffer []string) error {
	for _, line := range dataBuffer {
		s, err := sample.ParseSample(line)
		if err != nil {
			fmt.Printf("Warning: skip invalid sample: %v\n", err)
			continue
		}
		t.train(s.Y, s.X)
	}
	return nil
}

// LoadModel 加载模型
func (t *FTRLTrainer) LoadModel(modelPath, modelFormat string) error {
	return t.model.LoadModel(modelPath, modelFormat)
}

// OutputModel 输出模型
func (t *FTRLTrainer) OutputModel(modelPath, modelFormat string) error {
	return t.model.OutputModel(modelPath, modelFormat)
}

// train 训练一个样本
func (t *FTRLTrainer) train(y int, x []sample.FeatureValue) {
	thetaBias := t.model.GetOrInitModelUnitBias()
	xLen := len(x)
	theta := make([]*FTRLModelUnit, xLen)
	feaLocks := make([]*sync.Mutex, xLen+1)

	// 获取模型单元和锁
	for i := 0; i < xLen; i++ {
		theta[i] = t.model.GetOrInitModelUnit(x[i].Feature)
		feaLocks[i] = t.lockPool.GetFeatureLock(x[i].Feature)
	}
	feaLocks[xLen] = t.lockPool.GetBiasLock()

	// 更新w（FTRL）
	for i := 0; i <= xLen; i++ {
		var mu *FTRLModelUnit
		if i < xLen {
			mu = theta[i]
		} else {
			mu = thetaBias
		}

		if (i < xLen && t.opt.K1) || (i == xLen && t.opt.K0) {
			feaLocks[i].Lock()
			if math.Abs(mu.WZi) <= t.opt.WL1 {
				mu.Wi = 0.0
			} else {
				if t.opt.ForceVSparse && mu.WNi > 0 && mu.Wi == 0.0 {
					mu.ReinitVi(t.model.InitMean, t.model.InitStdev)
				}
				mu.Wi = -1.0 * (1.0 / (t.opt.WL2 + (t.opt.WBeta+math.Sqrt(mu.WNi))/t.opt.WAlpha)) *
					(mu.WZi - float64(utils.Sgn(mu.WZi))*t.opt.WL1)
			}
			feaLocks[i].Unlock()
		}
	}

	// 更新v（FTRL）
	for i := 0; i < xLen; i++ {
		mu := theta[i]
		for f := 0; f < t.model.FactorNum; f++ {
			feaLocks[i].Lock()
			if mu.VNi[f] > 0 {
				if t.opt.ForceVSparse && mu.Wi == 0.0 {
					mu.Vi[f] = 0.0
				} else if math.Abs(mu.VZi[f]) <= t.opt.VL1 {
					mu.Vi[f] = 0.0
				} else {
					mu.Vi[f] = -1.0 * (1.0 / (t.opt.VL2 + (t.opt.VBeta+math.Sqrt(mu.VNi[f]))/t.opt.VAlpha)) *
						(mu.VZi[f] - float64(utils.Sgn(mu.VZi[f]))*t.opt.VL1)
				}
			}
			feaLocks[i].Unlock()
		}
	}

	// 预测和计算sum（使用SIMD优化）
	bias := thetaBias.Wi
	var p float64
	sum := t.sumBuf[:t.model.FactorNum] // 复用预分配的缓冲区
	
	if t.useSIMD && xLen > 0 {
		// SIMD 优化版本：同时计算预测值和 sum
		p, sum = t.predictAndSumSIMD(x, bias, theta)
	} else {
		// 标量版本
		p = t.predictScalar(x, bias, theta)
		// 计算sum
		for f := 0; f < t.model.FactorNum; f++ {
			sumF := 0.0
			for i := 0; i < xLen; i++ {
				sumF += theta[i].Vi[f] * x[i].Value
			}
			sum[f] = sumF
		}
	}

	// 计算梯度系数
	mult := float64(y) * (1.0/(1.0+math.Exp(-p*float64(y))) - 1.0)

	// 更新w_n, w_z
	for i := 0; i <= xLen; i++ {
		var mu *FTRLModelUnit
		var xi float64
		if i < xLen {
			mu = theta[i]
			xi = x[i].Value
		} else {
			mu = thetaBias
			xi = 1.0
		}

		if (i < xLen && t.opt.K1) || (i == xLen && t.opt.K0) {
			feaLocks[i].Lock()
			wGi := mult * xi
			wSi := (1.0 / t.opt.WAlpha) * (math.Sqrt(mu.WNi+wGi*wGi) - math.Sqrt(mu.WNi))
			mu.WZi += wGi - wSi*mu.Wi
			mu.WNi += wGi * wGi
			feaLocks[i].Unlock()
		}
	}

	// 更新v_n, v_z（使用SIMD优化）
	if t.useSIMD && xLen > 0 {
		t.updateVGradientsSIMD(theta, feaLocks, x, sum, mult)
	} else {
		// 标量版本
		for i := 0; i < xLen; i++ {
			mu := theta[i]
			xi := x[i].Value

			for f := 0; f < t.model.FactorNum; f++ {
				feaLocks[i].Lock()
				vGif := mult * (sum[f]*xi - mu.Vi[f]*xi*xi)
				vSif := (1.0 / t.opt.VAlpha) * (math.Sqrt(mu.VNi[f]+vGif*vGif) - math.Sqrt(mu.VNi[f]))
				mu.VZi[f] += vGif - vSif*mu.Vi[f]
				mu.VNi[f] += vGif * vGif

				if t.opt.ForceVSparse && mu.VNi[f] > 0 && mu.Wi == 0.0 {
					mu.Vi[f] = 0.0
				}
				feaLocks[i].Unlock()
			}
		}
	}
}

// predictAndSumSIMD 使用SIMD同时计算预测值和sum
func (t *FTRLTrainer) predictAndSumSIMD(x []sample.FeatureValue, bias float64, theta []*FTRLModelUnit) (float64, []float64) {
	xLen := len(x)
	factorNum := t.model.FactorNum
	
	result := bias
	sum := t.sumBuf[:factorNum]
	
	// 清零 sum
	for f := 0; f < factorNum; f++ {
		sum[f] = 0.0
	}

	// 一阶项
	for i := 0; i < xLen; i++ {
		result += theta[i].Wi * x[i].Value
	}

	// 二阶交互项 - 使用SIMD优化
	// 对于每个特征，将其 Vi 向量乘以 x[i].Value 累加到 sum
	for i := 0; i < xLen; i++ {
		xi := x[i].Value
		vi := theta[i].Vi
		// sum += xi * vi (使用 BLAS Axpy: y = alpha*x + y)
		t.simdOps.Axpy(xi, vi, sum)
	}

	// 计算 result += 0.5 * (sum[f]^2 - sumSqr[f])
	// sumSqr[f] = Σ (vi[f] * xi)^2
	sumSqrTotal := 0.0
	for i := 0; i < xLen; i++ {
		xi := x[i].Value
		vi := theta[i].Vi
		// 使用 SIMD 计算 (vi * xi)^2 的和
		sumSqrTotal += t.simdOps.ScaledSumSquares(vi, xi)
	}
	
	// sum[f]^2 的总和
	sumTotal := t.simdOps.SumSquares(sum)
	
	result += 0.5 * (sumTotal - sumSqrTotal)

	return result, sum
}

// predictScalar 标量版本的预测
func (t *FTRLTrainer) predictScalar(x []sample.FeatureValue, bias float64, theta []*FTRLModelUnit) float64 {
	xLen := len(x)
	factorNum := t.model.FactorNum
	
	result := bias

	// 一阶项
	for i := 0; i < xLen; i++ {
		result += theta[i].Wi * x[i].Value
	}

	// 二阶交互项
	for f := 0; f < factorNum; f++ {
		sumF := 0.0
		sumSqr := 0.0
		for i := 0; i < xLen; i++ {
			d := theta[i].Vi[f] * x[i].Value
			sumF += d
			sumSqr += d * d
		}
		result += 0.5 * (sumF*sumF - sumSqr)
	}

	return result
}

// updateVGradientsSIMD 使用SIMD更新v的梯度
func (t *FTRLTrainer) updateVGradientsSIMD(theta []*FTRLModelUnit, feaLocks []*sync.Mutex, 
	x []sample.FeatureValue, sum []float64, mult float64) {
	
	xLen := len(x)
	factorNum := t.model.FactorNum
	
	// 预计算常用值
	invVAlpha := 1.0 / t.opt.VAlpha
	
	for i := 0; i < xLen; i++ {
		mu := theta[i]
		xi := x[i].Value
		xixi := xi * xi

		feaLocks[i].Lock()
		
		// 对于每个维度，使用向量化计算
		for f := 0; f < factorNum; f++ {
			vGif := mult * (sum[f]*xi - mu.Vi[f]*xixi)
			vGifSqr := vGif * vGif
			vSif := invVAlpha * (math.Sqrt(mu.VNi[f]+vGifSqr) - math.Sqrt(mu.VNi[f]))
			mu.VZi[f] += vGif - vSif*mu.Vi[f]
			mu.VNi[f] += vGifSqr

			if t.opt.ForceVSparse && mu.VNi[f] > 0 && mu.Wi == 0.0 {
				mu.Vi[f] = 0.0
			}
		}
		
		feaLocks[i].Unlock()
	}
}

