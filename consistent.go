package consistent

import (
	"hash/crc32"
	"sort"
)

// Not threadsafe
// Inspired by:
// https://github.com/golang/groupcache/blob/master/consistenthash/consistenthash.go
// https://github.com/stathat/consistent/blob/master/consistent.go

type Hash func(data []byte) uint32

type Consistent struct {
	hash    Hash
	keys    []int // Sorted
	hashMap map[int]string
}

func New(fn Hash) *Consistent {
	m := &Consistent{
		hash:    fn,
		hashMap: make(map[int]string),
	}

	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}

	return m
}

// Returns true if there are no items available.
func (m *Consistent) IsEmpty() bool {
	return len(m.keys) == 0
}

// Hash a key.
func (m *Consistent) Hash(key string) int {
	return int(m.hash([]byte(key)))
}

// Add a key to the hash.
func (m *Consistent) Add(key string) int {
	hash := m.Hash(key)

	if _, ok := m.hashMap[hash]; !ok {
		// Do not add another key to the sorted index if it already exists
		m.keys = append(m.keys, hash)
		sort.Ints(m.keys)
	}

	m.hashMap[hash] = key

	return hash
}

// Remove a key from the hash.
func (m *Consistent) Remove(key string) {
	hash := m.Hash(key)

	// Remove hash from m.keys
	i := sort.SearchInts(m.keys, hash)
	if i < len(m.keys) && m.keys[i] == hash {
		m.keys = append(m.keys[:i], m.keys[i+1:]...)
	}

	// Remove hash from hashMap
	delete(m.hashMap, hash)

	sort.Ints(m.keys)
}

// Get the item in the hash the provided key is in the range of.
func (m *Consistent) Get(key string) string {
	if m.IsEmpty() {
		return ""
	}

	hash := m.Hash(key)
	index := m.prev(hash)

	return m.hashMap[index]
}

// Get the next item in the hash to the provided key.
func (m *Consistent) Next(key string) string {
	if m.IsEmpty() {
		return ""
	}

	hash := m.Hash(key)
	candidate := m.next(hash)

	return m.hashMap[candidate]
}

// Get the range of hash keys to the provided item.
func (m *Consistent) Range(host string) (int, int) {
	if m.IsEmpty() {
		return 0, 0
	}

	from := m.Hash(host)
	to := m.next(from) - 1

	return from, to
}

func (m *Consistent) prev(hash int) int {
	rev := make([]int, len(m.keys))
	copy(rev, m.keys)
	sort.Sort(sort.Reverse(sort.IntSlice(rev)))

	i := sort.Search(len(rev), func(i int) bool { return rev[i] <= hash })

	if i == len(rev) {
		i -= 1
	}

	return rev[i]
}

func (m *Consistent) next(hash int) int {
	i := sort.Search(len(m.keys), func(i int) bool { return m.keys[i] > hash })

	if i == len(m.keys) {
		i = 0
	}

	return m.keys[i]
}
