package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

var (
	GlobalReplicas  = defaultReplicas
	defaultHash     = crc32.ChecksumIEEE // default use crc32.ChecksumIEEE as hash func
	defaultReplicas = 150
)

type Hash func(data []byte) uint32

type Map struct {
	hash     Hash           // hash func, a kind of datas need to be sure that with the same hash func
	replicas int            // 虚拟节点倍数
	keys     []int          // 哈希环，维护有序
	hashMap  map[int]string // 虚拟节点与真实节点的映射表（key是虚拟节点hash, value is the name of reality node）
}

type ConsOptions func(*Map)

// New with the replicas number and hash function
func New(opts ...ConsOptions) *Map {
	m := Map{
		hash:     defaultHash,
		replicas: defaultReplicas,
		hashMap:  make(map[int]string),
	}
	for _, opt := range opts {
		opt(&m)
	}
	return &m
}

func Replicas(replicas int) ConsOptions {
	return func(m *Map) {
		m.replicas = replicas
	}
}

func HashFunc(hash Hash) ConsOptions {
	return func(m *Map) {
		m.hash = hash
	}
}

// Add adds some keys to the hash.
// keys is the name of reality node
func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		// 一个真实节点对应多个虚拟节点
		for i := 0; i < m.replicas; i++ {
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			m.keys = append(m.keys, hash)
			// 维护虚拟节点与真实节点的映射关系
			m.hashMap[hash] = key
		}
	}
	sort.Ints(m.keys)
}

// Get gets the closest item in the hash to the provided key.
func (m *Map) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}
	hash := int(m.hash([]byte(key)))
	// 顺时针找到第一个匹配的虚拟节点
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})
	// 如果没有找到匹配的虚拟节点，返回哈希环上的第一个节点
	return m.hashMap[m.keys[idx%len(m.keys)]]
}

// Remove removes some node from the hash.
func (m *Map) Remove(key string) {
	for i := 0; i < m.replicas; i++ {
		hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
		idx := sort.SearchInts(m.keys, hash)
		m.keys = append(m.keys[:idx], m.keys[idx+1:]...)
		delete(m.hashMap, hash)
	}
}
