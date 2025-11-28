package model

import (
	"fmt"
	"math"
	"sync"

	"github.com/xiongle/alphaFM-go/pkg/lock"
	"github.com/xiongle/alphaFM-go/pkg/sample"
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
	}
}

// FTRLTrainer FTRL训练器
type FTRLTrainer struct {
	model        *FTRLModel
	lockPool     *lock.LockPool
	opt          *TrainerOption
}

// NewFTRLTrainer 创建训练器
func NewFTRLTrainer(opt *TrainerOption) *FTRLTrainer {
	return &FTRLTrainer{
		model:    NewFTRLModel(opt.FactorNum, opt.InitMean, opt.InitStdev),
		lockPool: lock.NewLockPool(),
		opt:      opt,
	}
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

	// 预测
	xForPredict := make([]struct{ Feature string; Value float64 }, xLen)
	for i := 0; i < xLen; i++ {
		xForPredict[i].Feature = x[i].Feature
		xForPredict[i].Value = x[i].Value
	}
	bias := thetaBias.Wi
	p := t.model.Predict(xForPredict, bias, theta)

	// 计算sum（用于v的梯度计算）
	sum := make([]float64, t.model.FactorNum)
	for f := 0; f < t.model.FactorNum; f++ {
		sumF := 0.0
		for i := 0; i < xLen; i++ {
			sumF += theta[i].Vi[f] * x[i].Value
		}
		sum[f] = sumF
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

	// 更新v_n, v_z
	for i := 0; i < xLen; i++ {
		mu := theta[i]
		xi := x[i].Value

		for f := 0; f < t.model.FactorNum; f++ {
			feaLocks[i].Lock()
			vGif := mult * (sum[f]*xi - mu.Vi[f]*xi*xi)
			vSif := (1.0 / t.opt.VAlpha) * (math.Sqrt(mu.VNi[f]+vGif*vGif) - math.Sqrt(mu.VNi[f]))
			mu.VZi[f] += vGif - vSif*mu.Vi[f]
			mu.VNi[f] += vGif * vGif

			// 处理只出现一次的特征
			if t.opt.ForceVSparse && mu.VNi[f] > 0 && mu.Wi == 0.0 {
				mu.Vi[f] = 0.0
			}
			feaLocks[i].Unlock()
		}
	}
}

