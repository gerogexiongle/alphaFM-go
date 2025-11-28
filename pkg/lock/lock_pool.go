package lock

import (
	"hash/fnv"
	"sync"
)

const lockPoolSize = 10009

// LockPool 特征锁池
type LockPool struct {
	locks     []sync.Mutex
	biasLock  sync.Mutex
}

// NewLockPool 创建锁池
func NewLockPool() *LockPool {
	return &LockPool{
		locks: make([]sync.Mutex, lockPoolSize),
	}
}

// GetFeatureLock 获取特征锁
func (lp *LockPool) GetFeatureLock(feature string) *sync.Mutex {
	h := fnv.New64a()
	h.Write([]byte(feature))
	index := h.Sum64() % lockPoolSize
	return &lp.locks[index]
}

// GetBiasLock 获取bias锁
func (lp *LockPool) GetBiasLock() *sync.Mutex {
	return &lp.biasLock
}

