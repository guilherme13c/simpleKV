package store

import (
	"hash/fnv"
	"sync"
)

type CountingBloomFilter struct {
	m      uint32   // Size of the filter (number of counters)
	k      uint32   // Number of hash functions
	filter []uint32 // The array of counters
	mu     sync.RWMutex
}

func NewCountingBloomFilter(m uint32, k uint32) *CountingBloomFilter {
	return &CountingBloomFilter{
		m:      m,
		k:      k,
		filter: make([]uint32, m),
	}
}

func (b *CountingBloomFilter) hash(key string) []uint32 {
	hashValues := make([]uint32, b.k)
	h := fnv.New32a()
	h.Write([]byte(key))
	baseHash := h.Sum32()

	for i := uint32(0); i < b.k; i++ {
		hashValues[i] = (baseHash + i*i) % b.m
	}

	return hashValues
}

func (b *CountingBloomFilter) Insert(key string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	hashes := b.hash(key)
	for _, hashVal := range hashes {
		b.filter[hashVal]++
	}
}

func (b *CountingBloomFilter) Remove(key string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	hashes := b.hash(key)
	for _, hashVal := range hashes {
		if b.filter[hashVal] > 0 {
			b.filter[hashVal]--
		}
	}
}

func (b *CountingBloomFilter) MightContain(key string) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()

	hashes := b.hash(key)
	for _, hashVal := range hashes {
		if b.filter[hashVal] == 0 {
			return false
		}
	}

	return true
}
