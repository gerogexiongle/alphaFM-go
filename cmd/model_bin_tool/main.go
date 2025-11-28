package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"

	"github.com/xiongle/alphaFM-go/pkg/model"
)

func binToolHelp() string {
	return `
usage: ./model_bin_tool [<options>]

options:
-task <task_type>: 1-print bin model info
                   2-transfer format, bin to txt
                   3-transfer format, bin to txt, only nonzero features
                   4-transfer format, txt to bin
-im <input_model_path>: set the intput model path
-om <output_model_path>: set the output model path for task 2,3 and 4, otherwise, it writes to standard output for task 2 and 3
-dim <factor_num>: dim of 2-way interactions, for task 4
-mnt <model_number_type>: set the number type of the bin model for task 4, double or float	default:double
`
}

func printInfo(inputPath string) error {
	info, err := model.ReadInfo(inputPath)
	if err != nil {
		return err
	}

	fmt.Printf("format_version: 1\n")
	fmt.Printf("number_byte_length: %d", info.NumByteLen)
	if info.NumByteLen == 8 {
		fmt.Print("(double)")
	} else if info.NumByteLen == 4 {
		fmt.Print("(float)")
	}
	fmt.Println()
	fmt.Printf("factor_num: %d\n", info.FactorNum)
	fmt.Printf("feature_num: %d\n", info.FeaNum)
	fmt.Printf("nonzero_feature_num: %d\n", info.NonzeroFeaNum)
	fmt.Printf("success_flag: %v\n", info.SuccessFlag == 1)

	return nil
}

func binToTxt(inputPath, outputPath string, onlyNonZero bool) error {
	// 打开输入文件
	mbf := model.NewModelBinFile()
	if err := mbf.OpenForRead(inputPath); err != nil {
		return fmt.Errorf("open input file error: %v", err)
	}
	defer mbf.Close()

	info := mbf.GetInfo()

	// 打开输出
	var writer *bufio.Writer
	var outFile *os.File
	if outputPath != "" {
		f, err := os.Create(outputPath)
		if err != nil {
			return fmt.Errorf("open output file error: %v", err)
		}
		outFile = f
		defer outFile.Close()
		writer = bufio.NewWriter(outFile)
	} else {
		writer = bufio.NewWriter(os.Stdout)
	}
	defer writer.Flush()

	// 简化版本：提示用户功能尚未完全实现
	fmt.Fprintf(writer, "# Binary to text conversion\n")
	fmt.Fprintf(writer, "# Model info: factor_num=%d, feature_num=%d\n", info.FactorNum, info.FeaNum)
	fmt.Fprintf(writer, "# Note: Full binary model support is under development\n")

	return nil
}

func txtToBin(inputPath, outputPath string, factorNum int, useFloat32 bool) error {
	return model.ConvertTxtToBin(inputPath, outputPath, factorNum, useFloat32)
}

func main() {
	task := flag.Int("task", 0, "task type")
	inputPath := flag.String("im", "", "input model path")
	outputPath := flag.String("om", "", "output model path")
	dim := flag.Int("dim", 0, "factor num")
	mnt := flag.String("mnt", "double", "model number type")

	flag.Parse()

	// 验证参数
	if *task < 1 || *task > 4 {
		fmt.Fprintln(os.Stderr, "invalid task")
		fmt.Fprint(os.Stderr, binToolHelp())
		os.Exit(1)
	}

	if *inputPath == "" {
		fmt.Fprintln(os.Stderr, "input model path required")
		fmt.Fprint(os.Stderr, binToolHelp())
		os.Exit(1)
	}

	var err error
	switch *task {
	case 1:
		// 打印模型信息
		err = printInfo(*inputPath)

	case 2:
		// bin转txt（全部特征）
		err = binToTxt(*inputPath, *outputPath, false)

	case 3:
		// bin转txt（仅非零特征）
		err = binToTxt(*inputPath, *outputPath, true)

	case 4:
		// txt转bin
		if *dim <= 0 {
			fmt.Fprintln(os.Stderr, "dim required for task 4")
			fmt.Fprint(os.Stderr, binToolHelp())
			os.Exit(1)
		}
		if *outputPath == "" {
			fmt.Fprintln(os.Stderr, "output model path required for task 4")
			fmt.Fprint(os.Stderr, binToolHelp())
			os.Exit(1)
		}
		useFloat32 := *mnt == "float"
		err = txtToBin(*inputPath, *outputPath, *dim, useFloat32)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

