package model

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/xiongle/alphaFM-go/pkg/utils"
)

const BiasFeatureName = "bias"

// FTRLModelUnit FTRL模型单元（针对每个特征）
type FTRLModelUnit struct {
	Wi   float64   // 一阶权重
	WNi  float64   // w的n参数（累积梯度平方和）
	WZi  float64   // w的z参数
	Vi   []float64 // 隐向量
	VNi  []float64 // v的n参数
	VZi  []float64 // v的z参数
}

// NewFTRLModelUnit 创建模型单元
func NewFTRLModelUnit(factorNum int, mean, stdev float64) *FTRLModelUnit {
	unit := &FTRLModelUnit{
		Wi:  0.0,
		WNi: 0.0,
		WZi: 0.0,
		Vi:  make([]float64, factorNum),
		VNi: make([]float64, factorNum),
		VZi: make([]float64, factorNum),
	}

	// 初始化隐向量
	for f := 0; f < factorNum; f++ {
		unit.Vi[f] = utils.GaussianWithParams(mean, stdev)
		unit.VNi[f] = 0.0
		unit.VZi[f] = 0.0
	}

	return unit
}

// NewFTRLModelUnitFromLine 从模型文件行创建
func NewFTRLModelUnitFromLine(factorNum int, parts []string) (*FTRLModelUnit, error) {
	if len(parts) != 3*factorNum+4 {
		return nil, fmt.Errorf("invalid model line format")
	}

	unit := &FTRLModelUnit{
		Vi:  make([]float64, factorNum),
		VNi: make([]float64, factorNum),
		VZi: make([]float64, factorNum),
	}

	var err error
	unit.Wi, err = strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return nil, err
	}

	// 解析v
	for f := 0; f < factorNum; f++ {
		unit.Vi[f], err = strconv.ParseFloat(parts[2+f], 64)
		if err != nil {
			return nil, err
		}
	}

	// w_n, w_z
	unit.WNi, err = strconv.ParseFloat(parts[2+factorNum], 64)
	if err != nil {
		return nil, err
	}
	unit.WZi, err = strconv.ParseFloat(parts[3+factorNum], 64)
	if err != nil {
		return nil, err
	}

	// v_n
	for f := 0; f < factorNum; f++ {
		unit.VNi[f], err = strconv.ParseFloat(parts[4+factorNum+f], 64)
		if err != nil {
			return nil, err
		}
	}

	// v_z
	for f := 0; f < factorNum; f++ {
		unit.VZi[f], err = strconv.ParseFloat(parts[4+2*factorNum+f], 64)
		if err != nil {
			return nil, err
		}
	}

	return unit, nil
}

// ReinitVi 重新初始化隐向量
func (u *FTRLModelUnit) ReinitVi(mean, stdev float64) {
	for f := 0; f < len(u.Vi); f++ {
		u.Vi[f] = utils.GaussianWithParams(mean, stdev)
	}
}

// IsNonZero 判断是否非零
func (u *FTRLModelUnit) IsNonZero() bool {
	if u.Wi != 0.0 {
		return true
	}
	for _, v := range u.Vi {
		if v != 0.0 {
			return true
		}
	}
	return false
}

// String 转为字符串（用于输出模型）
func (u *FTRLModelUnit) String() string {
	parts := []string{fmt.Sprintf("%.6g", u.Wi)}

	// vi
	for _, v := range u.Vi {
		parts = append(parts, fmt.Sprintf("%.6g", v))
	}

	// w_ni, w_zi
	parts = append(parts, fmt.Sprintf("%.6g", u.WNi))
	parts = append(parts, fmt.Sprintf("%.6g", u.WZi))

	// v_ni
	for _, vn := range u.VNi {
		parts = append(parts, fmt.Sprintf("%.6g", vn))
	}

	// v_zi
	for _, vz := range u.VZi {
		parts = append(parts, fmt.Sprintf("%.6g", vz))
	}

	return strings.Join(parts, " ")
}

// FTRLModel FTRL模型
type FTRLModel struct {
	MuBias    *FTRLModelUnit
	MuMap     map[string]*FTRLModelUnit
	FactorNum int
	InitMean  float64
	InitStdev float64
	mu        sync.RWMutex
}

// NewFTRLModel 创建FTRL模型
func NewFTRLModel(factorNum int, mean, stdev float64) *FTRLModel {
	return &FTRLModel{
		MuMap:     make(map[string]*FTRLModelUnit),
		FactorNum: factorNum,
		InitMean:  mean,
		InitStdev: stdev,
	}
}

// GetOrInitModelUnit 获取或初始化模型单元
func (m *FTRLModel) GetOrInitModelUnit(feature string) *FTRLModelUnit {
	m.mu.RLock()
	unit, exists := m.MuMap[feature]
	m.mu.RUnlock()

	if exists {
		return unit
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// 双重检查
	if unit, exists := m.MuMap[feature]; exists {
		return unit
	}

	unit = NewFTRLModelUnit(m.FactorNum, m.InitMean, m.InitStdev)
	m.MuMap[feature] = unit
	return unit
}

// GetOrInitModelUnitBias 获取或初始化bias单元
func (m *FTRLModel) GetOrInitModelUnitBias() *FTRLModelUnit {
	if m.MuBias == nil {
		m.mu.Lock()
		if m.MuBias == nil {
			m.MuBias = NewFTRLModelUnit(0, m.InitMean, m.InitStdev)
		}
		m.mu.Unlock()
	}
	return m.MuBias
}

// Predict FM预测
func (m *FTRLModel) Predict(x []struct{ Feature string; Value float64 }, bias float64, theta []*FTRLModelUnit) float64 {
	result := bias

	// 一阶项
	for i := 0; i < len(x); i++ {
		result += theta[i].Wi * x[i].Value
	}

	// 二阶交互项
	sum := make([]float64, m.FactorNum)
	for f := 0; f < m.FactorNum; f++ {
		sumF := 0.0
		sumSqr := 0.0
		for i := 0; i < len(x); i++ {
			d := theta[i].Vi[f] * x[i].Value
			sumF += d
			sumSqr += d * d
		}
		sum[f] = sumF
		result += 0.5 * (sumF*sumF - sumSqr)
	}

	return result
}

// LoadModel 加载模型
func (m *FTRLModel) LoadModel(modelPath, modelFormat string) error {
	if modelFormat == "txt" {
		return m.loadTxtModel(modelPath)
	} else if modelFormat == "bin" {
		return m.loadBinModel(modelPath)
	}
	return fmt.Errorf("unsupported model format: %s", modelFormat)
}

// loadTxtModel 加载文本模型
func (m *FTRLModel) loadTxtModel(modelPath string) error {
	file, err := os.Open(modelPath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	// 读取bias行
	if !scanner.Scan() {
		return fmt.Errorf("empty model file")
	}

	parts := strings.Fields(scanner.Text())
	if len(parts) != 4 {
		return fmt.Errorf("invalid bias line format")
	}

	m.MuBias = &FTRLModelUnit{}
	m.MuBias.Wi, err = strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return err
	}
	m.MuBias.WNi, err = strconv.ParseFloat(parts[2], 64)
	if err != nil {
		return err
	}
	m.MuBias.WZi, err = strconv.ParseFloat(parts[3], 64)
	if err != nil {
		return err
	}

	// 读取特征行
	for scanner.Scan() {
		parts := strings.Fields(scanner.Text())
		if len(parts) != 3*m.FactorNum+4 {
			return fmt.Errorf("invalid feature line format")
		}

		feature := parts[0]
		unit, err := NewFTRLModelUnitFromLine(m.FactorNum, parts)
		if err != nil {
			return err
		}

		m.MuMap[feature] = unit
	}

	return scanner.Err()
}

// loadBinModel 加载二进制模型
func (m *FTRLModel) loadBinModel(modelPath string) error {
	mbf := NewModelBinFile()
	if err := mbf.OpenForRead(modelPath); err != nil {
		return err
	}
	defer mbf.Close()

	info := mbf.GetInfo()
	
	// 验证factor_num
	if info.FactorNum != uint64(m.FactorNum) {
		return fmt.Errorf("factor_num mismatch: model=%d, expected=%d", info.FactorNum, m.FactorNum)
	}

	// 读取bias
	feaName, err := mbf.ReadOneFea()
	if err != nil {
		return fmt.Errorf("failed to read bias feature name: %v", err)
	}
	if feaName != BiasFeatureName {
		return fmt.Errorf("expected bias, got %s", feaName)
	}

	// 根据number_byte_len读取bias unit
	m.MuBias = &FTRLModelUnit{}
	if info.NumByteLen == 8 {
		// double
		if err := mbf.ReadOneUnitDouble(m.MuBias, 0); err != nil {
			return fmt.Errorf("failed to read bias unit: %v", err)
		}
	} else if info.NumByteLen == 4 {
		// float
		if err := mbf.ReadOneUnitFloat(m.MuBias, 0); err != nil {
			return fmt.Errorf("failed to read bias unit: %v", err)
		}
	} else {
		return fmt.Errorf("unsupported number_byte_len: %d", info.NumByteLen)
	}

	// 读取特征
	for {
		feaName, err := mbf.ReadOneFea()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return fmt.Errorf("failed to read feature name: %v", err)
		}

		unit := &FTRLModelUnit{
			Vi:  make([]float64, m.FactorNum),
			VNi: make([]float64, m.FactorNum),
			VZi: make([]float64, m.FactorNum),
		}

		if info.NumByteLen == 8 {
			if err := mbf.ReadOneUnitDouble(unit, m.FactorNum); err != nil {
				return fmt.Errorf("failed to read unit for %s: %v", feaName, err)
			}
		} else if info.NumByteLen == 4 {
			if err := mbf.ReadOneUnitFloat(unit, m.FactorNum); err != nil {
				return fmt.Errorf("failed to read unit for %s: %v", feaName, err)
			}
		}

		m.MuMap[feaName] = unit
	}

	return nil
}

// OutputModel 输出模型
func (m *FTRLModel) OutputModel(modelPath, modelFormat string) error {
	if modelFormat == "txt" {
		return m.outputTxtModel(modelPath)
	} else if modelFormat == "bin" {
		return m.outputBinModel(modelPath)
	}
	return fmt.Errorf("unsupported model format: %s", modelFormat)
}

// outputTxtModel 输出文本模型
func (m *FTRLModel) outputTxtModel(modelPath string) error {
	file, err := os.Create(modelPath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	// 输出bias
	fmt.Fprintf(writer, "%s %.6g %.6g %.6g\n", BiasFeatureName, m.MuBias.Wi, m.MuBias.WNi, m.MuBias.WZi)

	// 输出特征
	for feature, unit := range m.MuMap {
		fmt.Fprintf(writer, "%s %s\n", feature, unit.String())
	}

	return nil
}

// outputBinModel 输出二进制模型
func (m *FTRLModel) outputBinModel(modelPath string) error {
	// 计算unit长度: wi(8) + w_ni(8) + w_zi(8) + vi(8*k) + v_ni(8*k) + v_zi(8*k)
	unitLen := uint64(3*8 + 3*m.FactorNum*8)

	mbf := NewModelBinFile()
	if err := mbf.OpenForWrite(modelPath, 8, uint64(m.FactorNum), unitLen); err != nil {
		return err
	}
	defer mbf.Close()

	// 写入bias (factor_num = 0，所以没有v向量)
	if err := mbf.WriteOneFeaUnitDouble(BiasFeatureName, m.MuBias, 0, true); err != nil {
		return fmt.Errorf("failed to write bias: %v", err)
	}

	// 写入特征 (factor_num = m.FactorNum)
	for feature, unit := range m.MuMap {
		isNonZero := unit.IsNonZero()
		if err := mbf.WriteOneFeaUnitDouble(feature, unit, m.FactorNum, isNonZero); err != nil {
			return fmt.Errorf("failed to write feature %s: %v", feature, err)
		}
	}

	return nil
}

// PredictModel 预测模型（简化版，只包含wi和vi）
type PredictModel struct {
	MuBias    *PredictModelUnit
	MuMap     map[string]*PredictModelUnit
	FactorNum int
}

// PredictModelUnit 预测模型单元
type PredictModelUnit struct {
	Wi float64
	Vi []float64
}

// NewPredictModel 创建预测模型
func NewPredictModel(factorNum int) *PredictModel {
	return &PredictModel{
		MuMap:     make(map[string]*PredictModelUnit),
		FactorNum: factorNum,
	}
}

// GetScore 计算预测得分（包含sigmoid）
func (m *PredictModel) GetScore(x []struct{ Feature string; Value float64 }, bias float64) float64 {
	result := bias

	// 一阶项
	for i := 0; i < len(x); i++ {
		if unit, ok := m.MuMap[x[i].Feature]; ok {
			result += unit.Wi * x[i].Value
		}
	}

	// 二阶交互项
	for f := 0; f < m.FactorNum; f++ {
		sumF := 0.0
		sumSqr := 0.0
		for i := 0; i < len(x); i++ {
			if unit, ok := m.MuMap[x[i].Feature]; ok {
				d := unit.Vi[f] * x[i].Value
				sumF += d
				sumSqr += d * d
			}
		}
		result += 0.5 * (sumF*sumF - sumSqr)
	}

	// Sigmoid
	return 1.0 / (1.0 + math.Exp(-result))
}

// LoadModel 加载模型
func (m *PredictModel) LoadModel(modelPath, modelFormat string) error {
	if modelFormat == "txt" {
		return m.loadTxtModel(modelPath)
	} else if modelFormat == "bin" {
		return m.loadBinModel(modelPath)
	}
	return fmt.Errorf("unsupported model format: %s", modelFormat)
}

// loadTxtModel 加载文本模型
func (m *PredictModel) loadTxtModel(modelPath string) error {
	file, err := os.Open(modelPath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	// 读取bias
	if !scanner.Scan() {
		return fmt.Errorf("empty model file")
	}

	parts := strings.Fields(scanner.Text())
	if len(parts) != 4 {
		return fmt.Errorf("invalid bias line")
	}

	m.MuBias = &PredictModelUnit{Vi: make([]float64, 0)}
	m.MuBias.Wi, err = strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return err
	}

	// 读取特征
	for scanner.Scan() {
		parts := strings.Fields(scanner.Text())
		if len(parts) != 3*m.FactorNum+4 {
			return fmt.Errorf("invalid feature line")
		}

		feature := parts[0]
		unit := &PredictModelUnit{Vi: make([]float64, m.FactorNum)}

		unit.Wi, err = strconv.ParseFloat(parts[1], 64)
		if err != nil {
			return err
		}

		isNonZero := unit.Wi != 0.0
		for f := 0; f < m.FactorNum; f++ {
			unit.Vi[f], err = strconv.ParseFloat(parts[2+f], 64)
			if err != nil {
				return err
			}
			if unit.Vi[f] != 0.0 {
				isNonZero = true
			}
		}

		// 只加载非零特征
		if isNonZero {
			m.MuMap[feature] = unit
		}
	}

	return scanner.Err()
}

// loadBinModel 加载二进制模型
func (m *PredictModel) loadBinModel(modelPath string) error {
	mbf := NewModelBinFile()
	if err := mbf.OpenForRead(modelPath); err != nil {
		return err
	}
	defer mbf.Close()

	info := mbf.GetInfo()
	
	// 验证factor_num
	if info.FactorNum != uint64(m.FactorNum) {
		return fmt.Errorf("factor_num mismatch: model=%d, expected=%d", info.FactorNum, m.FactorNum)
	}

	// 读取bias
	feaName, err := mbf.ReadOneFea()
	if err != nil {
		return fmt.Errorf("failed to read bias feature name: %v", err)
	}
	if feaName != BiasFeatureName {
		return fmt.Errorf("expected bias, got %s", feaName)
	}

	// 读取bias unit（预测模型只需要wi，不需要n和z）
	biasUnit := &FTRLModelUnit{}
	if info.NumByteLen == 8 {
		if err := mbf.ReadOneUnitDouble(biasUnit, 0); err != nil {
			return fmt.Errorf("failed to read bias unit: %v", err)
		}
	} else if info.NumByteLen == 4 {
		if err := mbf.ReadOneUnitFloat(biasUnit, 0); err != nil {
			return fmt.Errorf("failed to read bias unit: %v", err)
		}
	} else {
		return fmt.Errorf("unsupported number_byte_len: %d", info.NumByteLen)
	}
	
	m.MuBias = &PredictModelUnit{
		Wi: biasUnit.Wi,
		Vi: make([]float64, 0),
	}

	// 读取特征
	for {
		feaName, err := mbf.ReadOneFea()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return fmt.Errorf("failed to read feature name: %v", err)
		}

		fullUnit := &FTRLModelUnit{
			Vi:  make([]float64, m.FactorNum),
			VNi: make([]float64, m.FactorNum),
			VZi: make([]float64, m.FactorNum),
		}

		if info.NumByteLen == 8 {
			if err := mbf.ReadOneUnitDouble(fullUnit, m.FactorNum); err != nil {
				return fmt.Errorf("failed to read unit for %s: %v", feaName, err)
			}
		} else if info.NumByteLen == 4 {
			if err := mbf.ReadOneUnitFloat(fullUnit, m.FactorNum); err != nil {
				return fmt.Errorf("failed to read unit for %s: %v", feaName, err)
			}
		}

		// 检查是否非零
		isNonZero := fullUnit.Wi != 0.0
		if !isNonZero {
			for _, v := range fullUnit.Vi {
				if v != 0.0 {
					isNonZero = true
					break
				}
			}
		}

		// 只加载非零特征
		if isNonZero {
			m.MuMap[feaName] = &PredictModelUnit{
				Wi: fullUnit.Wi,
				Vi: fullUnit.Vi,
			}
		}
	}

	return nil
}

