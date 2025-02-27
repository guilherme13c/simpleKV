package store

import (
	"hash/fnv"
	"regexp"
	"simpleKV/resp"
	"sync"
)

type shard struct {
	mu   sync.RWMutex
	Data map[string]resp.Value
}

func (s *store) getShard(key string) *shard {
	hash := fnv.New32a()
	hash.Write([]byte(key))

	return &s.Shards[uint(hash.Sum32())%uint(len(s.Shards))]
}

func (sh *shard) scanKeys(regex *regexp.Regexp) []string {
	var keys []string
	for key := range sh.Data {
		if regex != nil && regex.MatchString(key) || regex == nil {
			keys = append(keys, key)
		}
	}
	return keys
}
