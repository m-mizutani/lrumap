package lrumap_test

import (
	"testing"

	"github.com/m-mizutani/lrumap"
	"github.com/stretchr/testify/assert"
)

type testData struct {
	data []byte
}

func (x *testData) Key() *[]byte {
	return &x.data
}

func TestBasicScenario(t *testing.T) {
	lru := lrumap.New(12)
	key1 := []byte("abc")
	key2 := []byte("xyz")
	data := testData{data: key1}

	// Lookup key from empty table
	assert.Equal(t, 0, lru.Size())
	res := lru.Get(&key1)
	assert.Nil(t, res)

	// Put data with key1
	err := lru.Put(&data, 3)
	assert.Nil(t, err)
	// Can lookup data with key1, but can not with key2
	assert.NotNil(t, lru.Get(&key1))
	assert.Nil(t, lru.Get(&key2))

	// Prune() updates current tick and remove expired data
	assert.Nil(t, lru.Prune(1))
	// But data's ttl is 3, current tick is 1
	assert.NotNil(t, lru.Get(&key1))
	// Prune() updates current tick to 2
	assert.Nil(t, lru.Prune(1))
	assert.NotNil(t, lru.Get(&key1))

	// Prune() updates current tick to 3, and run pruning process.
	// This method returns pruned objects if exists.
	pruned := lru.Prune(1)
	assert.NotNil(t, pruned)
	assert.Equal(t, 1, len(*pruned))
	data2, ok := (*pruned)[0].(*testData)
	assert.True(t, ok)
	assert.Equal(t, data, *data2)

	assert.Nil(t, lru.Get(&key1))

}
