package lrumap

import (
	"bytes"
	"errors"
)

type LruData interface {
	Key() *[]byte
}

type LruMap struct {
	table   map[hashValue]*bucket
	frames  []frame
	current tick
	maxTick tick
}

func New(maxTick tick) *LruMap {
	lruMap := LruMap{
		frames:  make([]frame, maxTick+1),
		maxTick: maxTick,
	}
	return &lruMap
}

func (x *LruMap) Put(data LruData, ttl tick) error {
	if ttl > x.maxTick {
		return errors.New("TTL is over maxTick")
	}

	hv := fnvHash(*data.Key())
	bkt := x.table[hv]
	if bkt == nil {
		bkt = &bucket{}
		x.table[hv] = bkt
	}

	return nil
}

func (x *LruMap) Get(key *[]byte) LruData {
	return nil
}

func (x *LruMap) Prune(progress tick) *[]LruData {
	return nil
}

func (x *LruMap) Size() uint {
	return 0
}

type tick uint64

type node struct {
	next, prev *node
	frameLink  *node
	data       LruData
	latest     tick
	ttl        tick
}

func (x *node) attach(target *node) {
	next := x.next
	x.next = target
	target.prev = x
	if next != nil {
		next.prev = target
		target.next = next
	}
}

func (x *node) detach() {
	if x.next != nil {
		x.next.prev = x.prev
	}
	if x.prev != nil {
		x.prev.next = x.next
	}
	x.next = nil
	x.prev = nil
	return
}

func (x *node) equals(target *node) bool {
	if x.data == nil || target.data == nil {
		return false
	}
	return bytes.Equal(*(x.data.Key()), *(target.data.Key()))
}

type frame struct {
	link *node
}

type bucket struct {
	root node
}

func (x *bucket) insert(newNode *node) error {
	var p *node
	for p = &x.root; p.next != nil; p = p.next {
		if p.equals(newNode) {
			return errors.New("Duplicated key")
		}
	}

	p.attach(newNode)
	return nil
}

type hashValue uint64

// FNV hash based on gopacket.
// See http://isthe.com/chongo/tech/comp/fnv/.
func fnvHash(s []byte) (h hashValue) {
	h = fnvBasis
	for i := 0; i < len(s); i++ {
		h ^= hashValue(s[i])
		h *= fnvPrime
	}
	return
}

const fnvBasis = 14695981039346656037
const fnvPrime = 1099511628211
