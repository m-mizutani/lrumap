package lrumap

import (
	"bytes"
	"errors"
)

// LruData is an interface for data object for LruMap.
// Key() must returns unique key that is byte slice in the table.
type LruData interface {
	Key() *[]byte
}

// LruMap is a main structure of the library. A developer accesses
// data object in the table via LruMap instance.
type LruMap struct {
	table   map[hashValue]*bucket
	frames  []frame
	current tick
	maxTick tick
	count   int
}

// New is a constructor of LruMap
func New(maxTick tick) *LruMap {
	lruMap := LruMap{
		table:   map[hashValue]*bucket{},
		frames:  make([]frame, maxTick+1),
		maxTick: maxTick,
	}
	return &lruMap
}

// Put inserts data object into LruMap table.
// LruMap does not allow to insert object with duplicated key.
func (x *LruMap) Put(obj LruData, ttl tick) error {
	if ttl > x.maxTick {
		return errors.New("TTL is over maxTick")
	}

	hv := fnvHash(obj.Key())
	bkt := x.table[hv]
	if bkt == nil {
		bkt = &bucket{}
		x.table[hv] = bkt
	}

	newNode := node{
		data: obj,
	}
	bkt.insert(&newNode)

	cur := x.getFrame(x.current + ttl)
	cur.add(&newNode)

	x.count++

	return nil
}

// Get returns data object if exists.
func (x *LruMap) Get(key *[]byte) LruData {
	hv := fnvHash(key)
	bkt := x.table[hv]
	if bkt == nil {
		return nil
	}

	searched := bkt.search(key)
	if searched == nil {
		return nil
	}
	return searched.data
}

// Prune is update current tick by adding `progress`.
// If there is data object(s), they will be pruned and returned as slice.
func (x *LruMap) Prune(progress tick) *[]LruData {
	var res []LruData
	for i := tick(0); i < progress; i++ {
		f := x.getFrame(x.current + i)
		res = append(res, (*f.prune())...)
	}

	x.count -= len(res)
	x.current += progress
	return &res
}

// Size returns number of data object in the LruMap table.
func (x *LruMap) Size() int {
	return x.count
}

func (x *LruMap) getFrame(t tick) *frame {
	p := t % tick(len(x.frames))
	return &x.frames[p]
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
	return x.matchKey(target.data.Key())
}

func (x *node) matchKey(key *[]byte) bool {
	return bytes.Equal(*(x.data.Key()), *key)
}

type frame struct {
	link *node
}

func (x *frame) add(target *node) {
	next := x.link
	x.link = target
	target.frameLink = next
}

func (x *frame) prune() *[]LruData {
	var prunedData []LruData
	for link := x.link; link != nil; link = link.frameLink {
		link.detach()
		prunedData = append(prunedData, link.data)
	}
	x.link = nil
	return &prunedData
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

func (x *bucket) search(key *[]byte) *node {
	for p := x.root.next; p != nil; p = p.next {
		if p.matchKey(key) {
			return p
		}
	}

	return nil
}

type hashValue uint64

// FNV hash based on gopacket.
// See http://isthe.com/chongo/tech/comp/fnv/.
func fnvHash(s *[]byte) (h hashValue) {
	h = fnvBasis
	for i := 0; i < len(*s); i++ {
		h ^= hashValue((*s)[i])
		h *= fnvPrime
	}
	return
}

const fnvBasis = 14695981039346656037
const fnvPrime = 1099511628211
