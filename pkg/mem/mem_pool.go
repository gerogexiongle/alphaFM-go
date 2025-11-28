package mem

import (
	"sync"
)

const blockSize = 64 * 1024 * 1024 // 64MB

// MemPool 内存池
type MemPool struct {
	mu       sync.Mutex
	current  []byte
	offset   int
}

var globalPool = &MemPool{}

// GetMem 从内存池获取内存
func GetMem(size int) []byte {
	globalPool.mu.Lock()
	defer globalPool.mu.Unlock()

	if size > blockSize {
		// 超大内存直接分配
		return make([]byte, size)
	}

	if globalPool.current == nil || globalPool.offset+size > len(globalPool.current) {
		// 分配新块
		globalPool.current = make([]byte, blockSize)
		globalPool.offset = 0
	}

	result := globalPool.current[globalPool.offset : globalPool.offset+size]
	globalPool.offset += size
	return result
}

// Reset 重置内存池
func Reset() {
	globalPool.mu.Lock()
	defer globalPool.mu.Unlock()
	globalPool.current = nil
	globalPool.offset = 0
}

