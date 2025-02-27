package store

import (
	"hash/fnv"
	"regexp"
	"simpleKV/resp"
	"sync"
)

type shard struct {
	mu   sync.RWMutex
	data map[string]resp.Value
}

func (s *store) getShard(key string) *shard {
	hash := fnv.New32a()
	hash.Write([]byte(key))

	return &s.shards[uint(hash.Sum32())%uint(len(s.shards))]
}

func (sh *shard) scanKeys(regex *regexp.Regexp) []string {
	var keys []string
	for key := range sh.data {
		if regex != nil && regex.MatchString(key) || regex == nil {
			keys = append(keys, key)
		}
	}
	return keys
}
