package sample

import (
	"fmt"
	"strconv"
	"strings"
)

// FMSample 样本数据结构
type FMSample struct {
	Y int                         // 标签: 1 或 -1
	X []FeatureValue              // 特征列表
}

// FeatureValue 特征和值
type FeatureValue struct {
	Feature string
	Value   float64
}

// ParseSample 解析样本字符串
func ParseSample(line string) (*FMSample, error) {
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return nil, fmt.Errorf("empty line")
	}

	sample := &FMSample{
		X: make([]FeatureValue, 0),
	}

	// 解析标签
	label, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid label: %v", err)
	}
	if label > 0 {
		sample.Y = 1
	} else {
		sample.Y = -1
	}

	// 解析特征
	for i := 1; i < len(parts); i++ {
		kv := strings.Split(parts[i], ":")
		if len(kv) != 2 {
			return nil, fmt.Errorf("invalid feature format: %s", parts[i])
		}

		value, err := strconv.ParseFloat(kv[1], 64)
		if err != nil {
			return nil, fmt.Errorf("invalid feature value: %v", err)
		}

		// 跳过值为0的特征
		if value != 0 {
			sample.X = append(sample.X, FeatureValue{
				Feature: kv[0],
				Value:   value,
			})
		}
	}

	return sample, nil
}

