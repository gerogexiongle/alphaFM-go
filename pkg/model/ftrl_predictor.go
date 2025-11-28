package model

import (
	"bufio"
	"fmt"
	"os"
	"sync"

	"github.com/xiongle/alphaFM-go/pkg/sample"
)

// PredictorOption 预测选项
type PredictorOption struct {
	ModelPath       string
	ModelFormat     string
	PredictPath     string
	ModelNumberType string
	ThreadsNum      int
	FactorNum       int
}

// NewPredictorOption 创建默认预测选项
func NewPredictorOption() *PredictorOption {
	return &PredictorOption{
		FactorNum:       8,
		ThreadsNum:      1,
		ModelFormat:     "txt",
		ModelNumberType: "double",
	}
}

// FTRLPredictor FTRL预测器
type FTRLPredictor struct {
	model   *PredictModel
	opt     *PredictorOption
	outFile *os.File
	outMu   sync.Mutex
}

// NewFTRLPredictor 创建预测器
func NewFTRLPredictor(opt *PredictorOption) (*FTRLPredictor, error) {
	p := &FTRLPredictor{
		model: NewPredictModel(opt.FactorNum),
		opt:   opt,
	}

	// 加载模型
	fmt.Println("load model...")
	if err := p.model.LoadModel(opt.ModelPath, opt.ModelFormat); err != nil {
		return nil, fmt.Errorf("load model error: %v", err)
	}
	fmt.Println("model loading finished")

	// 打开输出文件
	f, err := os.Create(opt.PredictPath)
	if err != nil {
		return nil, fmt.Errorf("open output file error: %v", err)
	}
	p.outFile = f

	return p, nil
}

// RunTask 处理一批数据
func (p *FTRLPredictor) RunTask(dataBuffer []string) error {
	results := make([]string, len(dataBuffer))

	for i, line := range dataBuffer {
		s, err := sample.ParseSample(line)
		if err != nil {
			fmt.Printf("Warning: skip invalid sample: %v\n", err)
			continue
		}

		// 转换特征格式
		xForPredict := make([]struct{ Feature string; Value float64 }, len(s.X))
		for j := 0; j < len(s.X); j++ {
			xForPredict[j].Feature = s.X[j].Feature
			xForPredict[j].Value = s.X[j].Value
		}

		score := p.model.GetScore(xForPredict, p.model.MuBias.Wi)
		results[i] = fmt.Sprintf("%d %.6g", s.Y, score)
	}

	// 写入结果
	p.outMu.Lock()
	defer p.outMu.Unlock()

	writer := bufio.NewWriter(p.outFile)
	for _, result := range results {
		if result != "" {
			fmt.Fprintln(writer, result)
		}
	}
	writer.Flush()

	return nil
}

// Close 关闭预测器
func (p *FTRLPredictor) Close() error {
	if p.outFile != nil {
		return p.outFile.Close()
	}
	return nil
}

