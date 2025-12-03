package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/xiongle/alphaFM-go/pkg/frame"
	"github.com/xiongle/alphaFM-go/pkg/model"
	"github.com/xiongle/alphaFM-go/pkg/simd"
)

func trainHelp() string {
	return `
usage: cat sample | ./fm_train [<options>]

options:
-m <model_path>: set the output model path
-mf <model_format>: set the output model format, txt or bin	default:txt
-dim <k0,k1,k2>: k0=use bias, k1=use 1-way interactions, k2=dim of 2-way interactions	default:1,1,8
-init_stdev <stdev>: stdev for initialization of 2-way factors	default:0.1
-w_alpha <w_alpha>: w is updated via FTRL, alpha is one of the learning rate parameters	default:0.05
-w_beta <w_beta>: w is updated via FTRL, beta is one of the learning rate parameters	default:1.0
-w_l1 <w_L1_reg>: L1 regularization parameter of w	default:0.1
-w_l2 <w_L2_reg>: L2 regularization parameter of w	default:5.0
-v_alpha <v_alpha>: v is updated via FTRL, alpha is one of the learning rate parameters	default:0.05
-v_beta <v_beta>: v is updated via FTRL, beta is one of the learning rate parameters	default:1.0
-v_l1 <v_L1_reg>: L1 regularization parameter of v	default:0.1
-v_l2 <v_L2_reg>: L2 regularization parameter of v	default:5.0
-core <threads_num>: set the number of threads	default:1
-im <initial_model_path>: set the initial model path
-imf <initial_model_format>: set the initial model format, txt or bin	default:txt
-fvs <force_v_sparse>: if fvs is 1, set vi = 0 whenever wi = 0	default:0
-mnt <model_number_type>: double or float	default:double
-simd <simd_type>: SIMD optimization type (scalar, blas)	default:scalar
`
}

func parseDim(dimStr string) (bool, bool, int, error) {
	parts := strings.Split(dimStr, ",")
	if len(parts) != 3 {
		return false, false, 0, fmt.Errorf("invalid dim format")
	}

	k0, err := strconv.Atoi(parts[0])
	if err != nil {
		return false, false, 0, err
	}

	k1, err := strconv.Atoi(parts[1])
	if err != nil {
		return false, false, 0, err
	}

	k2, err := strconv.Atoi(parts[2])
	if err != nil {
		return false, false, 0, err
	}

	return k0 != 0, k1 != 0, k2, nil
}

func main() {
	rand.Seed(time.Now().UnixNano())

	// 定义命令行参数
	opt := model.NewTrainerOption()

	modelPath := flag.String("m", "", "model path")
	modelFormat := flag.String("mf", "txt", "model format")
	dimStr := flag.String("dim", "1,1,8", "k0,k1,k2")
	initStdev := flag.Float64("init_stdev", 0.1, "init stdev")
	wAlpha := flag.Float64("w_alpha", 0.05, "w alpha")
	wBeta := flag.Float64("w_beta", 1.0, "w beta")
	wL1 := flag.Float64("w_l1", 0.1, "w L1")
	wL2 := flag.Float64("w_l2", 5.0, "w L2")
	vAlpha := flag.Float64("v_alpha", 0.05, "v alpha")
	vBeta := flag.Float64("v_beta", 1.0, "v beta")
	vL1 := flag.Float64("v_l1", 0.1, "v L1")
	vL2 := flag.Float64("v_l2", 5.0, "v L2")
	core := flag.Int("core", 1, "threads num")
	initModelPath := flag.String("im", "", "initial model path")
	initModelFormat := flag.String("imf", "txt", "initial model format")
	fvs := flag.Int("fvs", 0, "force v sparse")
	mnt := flag.String("mnt", "double", "model number type")
	simdType := flag.String("simd", "scalar", "SIMD optimization type")

	flag.Parse()

	// 解析dim参数
	k0, k1, k2, err := parseDim(*dimStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid dim: %v\n", err)
		fmt.Fprint(os.Stderr, trainHelp())
		os.Exit(1)
	}

	// 设置选项
	opt.ModelPath = *modelPath
	opt.ModelFormat = *modelFormat
	opt.K0 = k0
	opt.K1 = k1
	opt.FactorNum = k2
	opt.InitStdev = *initStdev
	opt.WAlpha = *wAlpha
	opt.WBeta = *wBeta
	opt.WL1 = *wL1
	opt.WL2 = *wL2
	opt.VAlpha = *vAlpha
	opt.VBeta = *vBeta
	opt.VL1 = *vL1
	opt.VL2 = *vL2
	opt.ThreadsNum = *core
	opt.InitModelPath = *initModelPath
	opt.InitialModelFormat = *initModelFormat
	opt.ForceVSparse = *fvs == 1
	opt.ModelNumberType = *mnt
	
	// 解析SIMD类型
	parsedSIMD, err := simd.ParseVectorOpsType(*simdType)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid simd type: %v\n", err)
		fmt.Fprint(os.Stderr, trainHelp())
		os.Exit(1)
	}
	opt.SIMDType = parsedSIMD

	if *initModelPath != "" {
		opt.BInit = true
	}

	// 验证参数
	if opt.ModelPath == "" {
		fmt.Fprintln(os.Stderr, "model path required")
		fmt.Fprint(os.Stderr, trainHelp())
		os.Exit(1)
	}

	// 创建训练器
	trainer := model.NewFTRLTrainer(opt)

	// 如果需要加载初始模型
	if opt.BInit {
		fmt.Println("load model...")
		if err := trainer.LoadModel(opt.InitModelPath, opt.InitialModelFormat); err != nil {
			fmt.Fprintf(os.Stderr, "failed to load model: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("model loading finished")
	}

	// 运行训练框架
	pcFrame := frame.NewPCFrame()
	pcFrame.Init(trainer, opt.ThreadsNum)
	if err := pcFrame.Run(os.Stdin); err != nil {
		fmt.Fprintf(os.Stderr, "training error: %v\n", err)
		os.Exit(1)
	}

	// 输出模型
	fmt.Println("output model...")
	if err := trainer.OutputModel(opt.ModelPath, opt.ModelFormat); err != nil {
		fmt.Fprintf(os.Stderr, "failed to output model: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("model outputting finished")
}

