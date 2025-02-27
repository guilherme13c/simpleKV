package store

import (
	"hash/fnv"
	"sync"
)

type CountingBloomFilter struct {
	M      uint32   // Size of the filter (number of counters)
	K      uint32   // Number of hash functions
	Filter []uint32 // The array of counters
	mu     sync.RWMutex
}

func NewCountingBloomFilter(m uint32, k uint32) *CountingBloomFilter {
	return &CountingBloomFilter{
		M:      m,
		K:      k,
		Filter: make([]uint32, m),
	}
}

func (b *CountingBloomFilter) hash(key string) []uint32 {
	hashValues := make([]uint32, b.K)
	h := fnv.New32a()
	h.Write([]byte(key))
	baseHash := h.Sum32()

	for i := uint32(0); i < b.K; i++ {
		hashValues[i] = (baseHash + i*i) % b.M
	}

	return hashValues
}

func (b *CountingBloomFilter) Insert(key string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	hashes := b.hash(key)
	for _, hashVal := range hashes {
		b.Filter[hashVal]++
	}
}

func (b *CountingBloomFilter) Remove(key string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	hashes := b.hash(key)
	for _, hashVal := range hashes {
		if b.Filter[hashVal] > 0 {
			b.Filter[hashVal]--
		}
	}
}

func (b *CountingBloomFilter) MightContain(key string) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()

	hashes := b.hash(key)
	for _, hashVal := range hashes {
		if b.Filter[hashVal] == 0 {
			return false
		}
	}

	return true
}
