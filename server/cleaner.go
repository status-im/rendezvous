package server

import (
	"container/heap"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/p2p/enr"
)

// definitely rename
func NewCleaner() Cleaner {
	return Cleaner{
		heap:      []string{},
		deadlines: map[string]int64{},
	}
}

type Cleaner struct {
	heap      []string
	deadlines map[string]int64
}

func (c Cleaner) Id(index int) string {
	return c.heap[index]
}

func (c Cleaner) Len() int {
	return len(c.heap)
}

func (c Cleaner) Less(i, j int) bool {
	return c.deadlines[c.Id(i)] < c.deadlines[c.Id(j)]
}
func (c Cleaner) Swap(i, j int) {
	c.heap[i], c.heap[j] = c.heap[j], c.heap[i]
}

func (c *Cleaner) Push(record interface{}) {
	c.heap = append(c.heap, record.(string))
}

func (c *Cleaner) Pop() interface{} {
	old := c.heap
	n := len(old)
	x := old[n-1]
	c.heap = old[0 : n-1]
	delete(c.deadlines, x)
	return x
}

func (c *Cleaner) Add(ttl int64, record enr.Record) {
	c.deadlines[common.Bytes2Hex(record.NodeAddr())] = time.Now().Add(time.Duration(ttl)).UnixNano()
	heap.Push(c, common.Bytes2Hex(record.NodeAddr()))
}
