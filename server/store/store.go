package store

import (
	"fmt"
	"regexp"
	"simpleKV/resp"
	"strings"
	"sync"
	"time"
)

type IStore interface {
	Set(key string, value resp.Value)
	Get(key string) (resp.Value, bool)
	Del(key string) bool
	Scan(cursor string, matchPattern string, count int) resp.Value
	SaveToDisk() error
	LoadFromDisk() error
}

type store struct {
	Shards      []shard
	BloomFilter *CountingBloomFilter

	mu              sync.Mutex
	persistenceFile string
}

func NewStore(numShards int, bloomSize uint32) IStore {
	shards := make([]shard, numShards)
	for i := range shards {
		shards[i] = shard{
			Data: make(map[string]resp.Value),
		}
	}
	newStore := &store{
		Shards:          shards,
		BloomFilter:     NewCountingBloomFilter(bloomSize, 3),
		mu:              sync.Mutex{},
		persistenceFile: "dump.rdb",
	}

	newStore.LoadFromDisk()

	go func() {
		for {
			time.Sleep(60 * time.Second)

			err := newStore.SaveToDisk()
			if err != nil {
				fmt.Println("Error saving snapshot:", err)
			}
		}
	}()

	return newStore
}

func (s *store) Set(key string, value resp.Value) {
	shard := s.getShard(key)

	shard.mu.Lock()
	defer shard.mu.Unlock()

	shard.Data[key] = value

	s.BloomFilter.Insert(key)
}

func (s *store) Get(key string) (resp.Value, bool) {
	if !s.BloomFilter.MightContain(key) {
		return resp.Value{}, false
	}

	shard := s.getShard(key)

	shard.mu.RLock()
	defer shard.mu.RUnlock()

	val, ok := shard.Data[key]

	return val, ok
}

func (s *store) Del(key string) bool {
	if !s.BloomFilter.MightContain(key) {
		return false
	}

	shard := s.getShard(key)

	shard.mu.Lock()
	defer shard.mu.Unlock()

	if _, ok := shard.Data[key]; ok {
		delete(shard.Data, key)
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

	for i := range s.Shards {
		shardKeys := s.Shards[i].scanKeys(regex)
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

	endIdx := min(startIdx + count, len(allKeys))

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

func (s *store) SaveToDisk() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.saveSnapshot()
}

func (s *store) LoadFromDisk() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.loadSnapshot()
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
