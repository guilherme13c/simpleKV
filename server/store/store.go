package store

import (
	"fmt"
	"regexp"
	"simpleKV/resp"
	"strings"
)

type IStore interface {
	Set(key string, value resp.Value)
	Get(key string) (resp.Value, bool)
	Del(key string) bool
	Scan(cursor string, matchPattern string, count int) resp.Value
}

type store struct {
	shards      []shard
	bloomFilter *CountingBloomFilter
}

func NewStore(numShards int, bloomSize uint32) IStore {
	shards := make([]shard, numShards)
	for i := range shards {
		shards[i] = shard{
			data: make(map[string]resp.Value),
		}
	}
	return &store{
		shards:      shards,
		bloomFilter: NewCountingBloomFilter(bloomSize, 3),
	}
}

func (s *store) Set(key string, value resp.Value) {
	shard := s.getShard(key)

	shard.mu.Lock()
	defer shard.mu.Unlock()

	shard.data[key] = value
}

func (s *store) Get(key string) (resp.Value, bool) {
	shard := s.getShard(key)

	shard.mu.RLock()
	defer shard.mu.RUnlock()

	val, ok := shard.data[key]

	return val, ok
}

func (s *store) Del(key string) bool {
	shard := s.getShard(key)

	shard.mu.Lock()
	defer shard.mu.Unlock()

	if _, ok := shard.data[key]; ok {
		delete(shard.data, key)
		return true
	}

	return false
}

func (s *store) Scan(cursor string, matchPattern string, count int) resp.Value {
	var allKeys []string

	var regex *regexp.Regexp
	var err error
	if matchPattern != "" {
		escapedPattern := strings.Replace(matchPattern, "*", ".*", -1)
		escapedPattern = strings.Replace(escapedPattern, "?", ".", -1)
		regex, err = regexp.Compile("^" + escapedPattern + "$")
		if err != nil {
			return resp.Value{
				Type:   resp.SIMPLE_ERROR,
				String: fmt.Sprintf("Invalid pattern: %v", err),
			}
		}
	}

	for i := range s.shards {
		shardKeys := s.shards[i].scanKeys(regex)
		allKeys = append(allKeys, shardKeys...)
	}

	startIdx := 0
	if cursor != "0" {
		var idx int
		_, err := fmt.Sscanf(cursor, "%d", &idx)
		if err != nil {
			startIdx = 0
		} else {
			startIdx = idx
		}
	}

	endIdx := startIdx + count
	if endIdx > len(allKeys) {
		endIdx = len(allKeys)
	}

	resultKeys := allKeys[startIdx:endIdx]

	newCursor := "0"
	if endIdx < len(allKeys) {
		newCursor = fmt.Sprintf("%d", endIdx)
	}

	return resp.Value{
		Type: resp.ARRAY,
		Array: []resp.Value{
			{
				Type:   resp.SIMPLE_STRING,
				String: newCursor,
			},
			{
				Type:  resp.ARRAY,
				Array: createBulkStringArray(resultKeys),
			},
		},
	}
}

func createBulkStringArray(keys []string) []resp.Value {
	var arr []resp.Value
	for _, key := range keys {
		arr = append(arr, resp.Value{
			Type:       resp.BULK_STRING,
			BulkString: key,
		})
	}
	return arr
}
