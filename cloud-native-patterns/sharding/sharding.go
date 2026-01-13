package main

import (
	"hash/fnv"
	"reflect"
	"sync"
)

type Shard[K comparable, V any] struct {
	sync.RWMutex
	items map[K]V
}

type ShardMap[K comparable, V any] []*Shard[K, V]

func NewShardMap[K comparable, V any](nshards int) ShardMap[K, V] {
	shards := make([]*Shard[K, V], nshards)
	for i := 0; i < nshards; i++ {
		shard := make(map[K]V)
		shards[i] = &Shard[K, V]{items: shard}
	}
	return shards
}

func (m ShardMap[K, V]) getShardIndex(key K) int {
	str := reflect.ValueOf(key).String()
	hash := fnv.New32a()
	hash.Write([]byte(str))
	sum := int(hash.Sum32())
	return sum % len(m)
}

func (m ShardMap[K, V]) getShard(key K) *Shard[K, V] {
	idx := m.getShardIndex(key)
	return m[idx]
}

func (m ShardMap[K, V]) Get(key K) V {
	shard := m.getShard(key)
	shard.RLock()
	defer shard.RUnlock()

	return shard.items[key]
}

func (m ShardMap[K, V]) Set(key K, value V) {
	shard := m.getShard(key)
	shard.Lock()
	defer shard.Unlock()

	shard.items[key] = value
}

func (m ShardMap[K, V]) Keys() []K {
	var keys []K
	var mutex sync.Mutex

	var wg sync.WaitGroup
	wg.Add(len(m))

	for _, shard := range m {
		go func(s *Shard[K, V]) {
			s.RLock()
			defer s.RUnlock()
			defer wg.Done()

			for key, _ := range s {
				mutex.Lock()
				keys = append(keys, key)
				mutex.Unlock()
			}
		}(shard)
	}

	wg.Wait()

	return keys
}
