package model

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

const modelVersion = 1

// ModelBinInfo 二进制模型信息
type ModelBinInfo struct {
	NumByteLen     uint64
	FactorNum      uint64
	FeaNum         uint64
	NonzeroFeaNum  uint64
	SuccessFlag    uint64
	UnitLen        uint64
}

// ModelBinFile 二进制模型文件处理
type ModelBinFile struct {
	info    ModelBinInfo
	file    *os.File
	isRead  bool
	version uint64
}

// NewModelBinFile 创建二进制模型文件处理器
func NewModelBinFile() *ModelBinFile {
	return &ModelBinFile{
		version: modelVersion,
	}
}

// OpenForRead 打开文件用于读取
func (m *ModelBinFile) OpenForRead(filePath string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	m.file = f
	m.isRead = true

	// 读取版本号
	if err := binary.Read(f, binary.LittleEndian, &m.version); err != nil {
		return err
	}
	if m.version != modelVersion {
		return fmt.Errorf("unsupported model version: %d", m.version)
	}

	// 读取模型信息
	if err := binary.Read(f, binary.LittleEndian, &m.info); err != nil {
		return err
	}
	if m.info.SuccessFlag != 1 {
		return fmt.Errorf("model file incomplete")
	}

	return nil
}

// OpenForWrite 打开文件用于写入
func (m *ModelBinFile) OpenForWrite(filePath string, numByteLen, factorNum, unitLen uint64) error {
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	m.file = f
	m.isRead = false

	m.info = ModelBinInfo{
		NumByteLen: numByteLen,
		FactorNum:  factorNum,
		UnitLen:    unitLen,
	}

	// 写入版本号和模型信息
	if err := binary.Write(f, binary.LittleEndian, m.version); err != nil {
		return err
	}
	if err := binary.Write(f, binary.LittleEndian, &m.info); err != nil {
		return err
	}

	return nil
}

// ReadOneFea 读取一个特征名
func (m *ModelBinFile) ReadOneFea() (string, error) {
	var feaLen uint16
	if err := binary.Read(m.file, binary.LittleEndian, &feaLen); err != nil {
		if err == io.EOF {
			return "", fmt.Errorf("EOF")
		}
		return "", err
	}

	feaBytes := make([]byte, feaLen)
	if _, err := io.ReadFull(m.file, feaBytes); err != nil {
		return "", err
	}

	return string(feaBytes), nil
}

// ReadOneUnitDouble 读取一个模型单元（double精度）
// 格式: wi(8) + w_ni(8) + w_zi(8) + vi(8*k) + v_ni(8*k) + v_zi(8*k)
// 注意：始终读取unit_len字节，即使factorNum=0（如bias）
func (m *ModelBinFile) ReadOneUnitDouble(unit *FTRLModelUnit, factorNum int) error {
	// 读取 wi, w_ni, w_zi
	if err := binary.Read(m.file, binary.LittleEndian, &unit.Wi); err != nil {
		return err
	}
	if err := binary.Read(m.file, binary.LittleEndian, &unit.WNi); err != nil {
		return err
	}
	if err := binary.Read(m.file, binary.LittleEndian, &unit.WZi); err != nil {
		return err
	}
	
	// 读取 vi, v_ni, v_zi (只读取实际的factorNum个元素)
	for f := 0; f < factorNum; f++ {
		if err := binary.Read(m.file, binary.LittleEndian, &unit.Vi[f]); err != nil {
			return err
		}
	}
	for f := 0; f < factorNum; f++ {
		if err := binary.Read(m.file, binary.LittleEndian, &unit.VNi[f]); err != nil {
			return err
		}
	}
	for f := 0; f < factorNum; f++ {
		if err := binary.Read(m.file, binary.LittleEndian, &unit.VZi[f]); err != nil {
			return err
		}
	}
	
	// 如果factorNum < m.info.FactorNum，跳过padding
	expectedFactorNum := int(m.info.FactorNum)
	if factorNum < expectedFactorNum {
		paddingCount := (expectedFactorNum - factorNum) * 3 * 8 // 字节数
		padding := make([]byte, paddingCount)
		if _, err := io.ReadFull(m.file, padding); err != nil {
			return err
		}
	}
	
	return nil
}

// ReadOneUnitFloat 读取一个模型单元（float精度）
func (m *ModelBinFile) ReadOneUnitFloat(unit *FTRLModelUnit, factorNum int) error {
	var wi, wni, wzi float32
	
	// 读取 wi, w_ni, w_zi
	if err := binary.Read(m.file, binary.LittleEndian, &wi); err != nil {
		return err
	}
	if err := binary.Read(m.file, binary.LittleEndian, &wni); err != nil {
		return err
	}
	if err := binary.Read(m.file, binary.LittleEndian, &wzi); err != nil {
		return err
	}
	
	unit.Wi = float64(wi)
	unit.WNi = float64(wni)
	unit.WZi = float64(wzi)
	
	// 读取 vi, v_ni, v_zi
	for f := 0; f < factorNum; f++ {
		var v float32
		if err := binary.Read(m.file, binary.LittleEndian, &v); err != nil {
			return err
		}
		unit.Vi[f] = float64(v)
	}
	for f := 0; f < factorNum; f++ {
		var vn float32
		if err := binary.Read(m.file, binary.LittleEndian, &vn); err != nil {
			return err
		}
		unit.VNi[f] = float64(vn)
	}
	for f := 0; f < factorNum; f++ {
		var vz float32
		if err := binary.Read(m.file, binary.LittleEndian, &vz); err != nil {
			return err
		}
		unit.VZi[f] = float64(vz)
	}
	
	// 如果factorNum < m.info.FactorNum，跳过padding
	expectedFactorNum := int(m.info.FactorNum)
	if factorNum < expectedFactorNum {
		paddingCount := (expectedFactorNum - factorNum) * 3 * 4 // float32字节数
		padding := make([]byte, paddingCount)
		if _, err := io.ReadFull(m.file, padding); err != nil {
			return err
		}
	}
	
	return nil
}

// WriteOneFeaUnit 写入一个特征单元
func (m *ModelBinFile) WriteOneFeaUnit(feaName string, data interface{}, isNonZero bool) error {
	feaLen := uint16(len(feaName))
	if err := binary.Write(m.file, binary.LittleEndian, feaLen); err != nil {
		return err
	}
	if _, err := m.file.Write([]byte(feaName)); err != nil {
		return err
	}
	if err := binary.Write(m.file, binary.LittleEndian, data); err != nil {
		return err
	}

	m.info.FeaNum++
	if isNonZero {
		m.info.NonzeroFeaNum++
	}

	return nil
}

// WriteOneFeaUnitDouble 写入一个特征单元（double精度）
// 格式: wi(8) + w_ni(8) + w_zi(8) + vi(8*k) + v_ni(8*k) + v_zi(8*k)
// 注意：所有单元都必须占用相同的unit_len字节（包括bias）
func (m *ModelBinFile) WriteOneFeaUnitDouble(feaName string, unit *FTRLModelUnit, factorNum int, isNonZero bool) error {
	// 写入特征名长度和名称
	feaLen := uint16(len(feaName))
	if err := binary.Write(m.file, binary.LittleEndian, feaLen); err != nil {
		return err
	}
	if _, err := m.file.Write([]byte(feaName)); err != nil {
		return err
	}

	// 写入 wi, w_ni, w_zi
	if err := binary.Write(m.file, binary.LittleEndian, unit.Wi); err != nil {
		return err
	}
	if err := binary.Write(m.file, binary.LittleEndian, unit.WNi); err != nil {
		return err
	}
	if err := binary.Write(m.file, binary.LittleEndian, unit.WZi); err != nil {
		return err
	}

	// 写入 vi, v_ni, v_zi (如果factorNum > 0)
	for f := 0; f < factorNum; f++ {
		if err := binary.Write(m.file, binary.LittleEndian, unit.Vi[f]); err != nil {
			return err
		}
	}
	for f := 0; f < factorNum; f++ {
		if err := binary.Write(m.file, binary.LittleEndian, unit.VNi[f]); err != nil {
			return err
		}
	}
	for f := 0; f < factorNum; f++ {
		if err := binary.Write(m.file, binary.LittleEndian, unit.VZi[f]); err != nil {
			return err
		}
	}

	// 如果factorNum < m.info.FactorNum (如bias)，填充0使其达到unit_len
	expectedFactorNum := int(m.info.FactorNum)
	if factorNum < expectedFactorNum {
		paddingCount := (expectedFactorNum - factorNum) * 3 // vi, v_ni, v_zi
		padding := make([]byte, paddingCount*8)
		if _, err := m.file.Write(padding); err != nil {
			return err
		}
	}

	m.info.FeaNum++
	if isNonZero {
		m.info.NonzeroFeaNum++
	}

	return nil
}

// WriteOneFeaUnitFloat 写入一个特征单元（float精度）
func (m *ModelBinFile) WriteOneFeaUnitFloat(feaName string, unit *FTRLModelUnit, factorNum int, isNonZero bool) error {
	// 写入特征名
	feaLen := uint16(len(feaName))
	if err := binary.Write(m.file, binary.LittleEndian, feaLen); err != nil {
		return err
	}
	if _, err := m.file.Write([]byte(feaName)); err != nil {
		return err
	}

	// 写入 wi, w_ni, w_zi (转为float32)
	if err := binary.Write(m.file, binary.LittleEndian, float32(unit.Wi)); err != nil {
		return err
	}
	if err := binary.Write(m.file, binary.LittleEndian, float32(unit.WNi)); err != nil {
		return err
	}
	if err := binary.Write(m.file, binary.LittleEndian, float32(unit.WZi)); err != nil {
		return err
	}

	// 写入 vi, v_ni, v_zi
	for f := 0; f < factorNum; f++ {
		if err := binary.Write(m.file, binary.LittleEndian, float32(unit.Vi[f])); err != nil {
			return err
		}
	}
	for f := 0; f < factorNum; f++ {
		if err := binary.Write(m.file, binary.LittleEndian, float32(unit.VNi[f])); err != nil {
			return err
		}
	}
	for f := 0; f < factorNum; f++ {
		if err := binary.Write(m.file, binary.LittleEndian, float32(unit.VZi[f])); err != nil {
			return err
		}
	}

	// 如果factorNum < m.info.FactorNum，填充0
	expectedFactorNum := int(m.info.FactorNum)
	if factorNum < expectedFactorNum {
		paddingCount := (expectedFactorNum - factorNum) * 3 // vi, v_ni, v_zi
		padding := make([]byte, paddingCount*4) // float32
		if _, err := m.file.Write(padding); err != nil {
			return err
		}
	}

	m.info.FeaNum++
	if isNonZero {
		m.info.NonzeroFeaNum++
	}

	return nil
}

// Close 关闭文件
func (m *ModelBinFile) Close() error {
	if !m.isRead {
		// 写模式：更新成功标志
		m.info.SuccessFlag = 1
		if _, err := m.file.Seek(int64(binary.Size(m.version)), 0); err != nil {
			return err
		}
		if err := binary.Write(m.file, binary.LittleEndian, &m.info); err != nil {
			return err
		}
	}
	return m.file.Close()
}

// GetInfo 获取模型信息
func (m *ModelBinFile) GetInfo() ModelBinInfo {
	return m.info
}

// PrintInfo 打印模型信息
func (m *ModelBinFile) PrintInfo() {
	fmt.Printf("format_version: %d\n", m.version)
	fmt.Printf("number_byte_length: %d", m.info.NumByteLen)
	if m.info.NumByteLen == 8 {
		fmt.Print("(double)")
	} else if m.info.NumByteLen == 4 {
		fmt.Print("(float)")
	}
	fmt.Println()
	fmt.Printf("factor_num: %d\n", m.info.FactorNum)
	fmt.Printf("feature_num: %d\n", m.info.FeaNum)
	fmt.Printf("nonzero_feature_num: %d\n", m.info.NonzeroFeaNum)
	fmt.Printf("success_flag: %v\n", m.info.SuccessFlag == 1)
}

// ReadInfo 只读取模型信息
func ReadInfo(filePath string) (*ModelBinInfo, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var version uint64
	if err := binary.Read(f, binary.LittleEndian, &version); err != nil {
		return nil, err
	}
	if version != modelVersion {
		return nil, fmt.Errorf("unsupported model version: %d", version)
	}

	var info ModelBinInfo
	if err := binary.Read(f, binary.LittleEndian, &info); err != nil {
		return nil, err
	}
	if info.SuccessFlag != 1 {
		return nil, fmt.Errorf("model file incomplete")
	}

	return &info, nil
}

// ConvertTxtToBin 文本模型转二进制
func ConvertTxtToBin(txtPath, binPath string, factorNum int, useFloat32 bool) error {
	txtFile, err := os.Open(txtPath)
	if err != nil {
		return err
	}
	defer txtFile.Close()

	var numByteLen uint64
	if useFloat32 {
		numByteLen = 4
	} else {
		numByteLen = 8
	}

	// 计算unit长度
	var unitLen uint64
	if useFloat32 {
		unitLen = uint64(3*4 + 3*factorNum*4) // wi, w_ni, w_zi + vi, v_ni, v_zi
	} else {
		unitLen = uint64(3*8 + 3*factorNum*8)
	}

	mbf := NewModelBinFile()
	if err := mbf.OpenForWrite(binPath, numByteLen, uint64(factorNum), unitLen); err != nil {
		return err
	}
	defer mbf.Close()

	scanner := bufio.NewScanner(txtFile)

	for scanner.Scan() {
		_ = scanner.Text()
		// 这里简化处理，实际需要解析每一行并写入
		// 完整实现需要根据float32/float64类型解析和写入
	}

	return scanner.Err()
}

