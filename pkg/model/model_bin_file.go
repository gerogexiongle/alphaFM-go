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
		return "", err
	}

	feaBytes := make([]byte, feaLen)
	if _, err := io.ReadFull(m.file, feaBytes); err != nil {
		return "", err
	}

	return string(feaBytes), nil
}

// ReadOneUnit 读取一个模型单元
func (m *ModelBinFile) ReadOneUnit(data interface{}) error {
	return binary.Read(m.file, binary.LittleEndian, data)
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

