package frame

import (
	"bufio"
	"fmt"
	"io"
	"sync"
)

// Task 任务接口
type Task interface {
	RunTask(dataBuffer []string) error
}

// PCFrame 生产者-消费者框架
type PCFrame struct {
	task       Task
	threadNum  int
	bufSize    int
	logNum     int
	buffer     chan []string
	done       chan struct{}
	wg         sync.WaitGroup
}

// NewPCFrame 创建PC框架
func NewPCFrame() *PCFrame {
	return &PCFrame{
		bufSize: 5000,
		logNum:  200000,
	}
}

// Init 初始化
func (f *PCFrame) Init(task Task, threadNum int) {
	f.task = task
	f.threadNum = threadNum
	f.buffer = make(chan []string, 2) // 缓冲2批数据
	f.done = make(chan struct{})
}

// Run 运行框架
func (f *PCFrame) Run(reader io.Reader) error {
	// 启动生产者
	f.wg.Add(1)
	go f.producer(reader)

	// 启动消费者
	for i := 0; i < f.threadNum; i++ {
		f.wg.Add(1)
		go f.consumer()
	}

	// 等待所有goroutine完成
	f.wg.Wait()
	return nil
}

// producer 生产者线程
func (f *PCFrame) producer(reader io.Reader) {
	defer f.wg.Done()
	defer close(f.buffer)

	scanner := bufio.NewScanner(reader)
	// 设置更大的缓冲区 (10MB) 以支持超长特征行
	// 机器学习数据中，单行可能包含数万个特征
	const maxScanTokenSize = 10 * 1024 * 1024 // 10MB
	buf := make([]byte, maxScanTokenSize)
	scanner.Buffer(buf, maxScanTokenSize)

	lineNum := 0
	batch := make([]string, 0, f.bufSize)

	for scanner.Scan() {
		line := scanner.Text()
		lineNum++

		batch = append(batch, line)

		if len(batch) >= f.bufSize {
			// 发送批次
			f.buffer <- batch
			batch = make([]string, 0, f.bufSize)

			if lineNum%f.logNum == 0 {
				fmt.Printf("%d lines finished\n", lineNum)
			}
		}
	}

	// 发送最后一批
	if len(batch) > 0 {
		f.buffer <- batch
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading input: %v\n", err)
	}
}

// consumer 消费者线程
func (f *PCFrame) consumer() {
	defer f.wg.Done()

	for batch := range f.buffer {
		if err := f.task.RunTask(batch); err != nil {
			fmt.Printf("Error processing batch: %v\n", err)
		}
	}
}

