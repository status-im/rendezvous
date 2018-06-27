package server

import (
	"container/heap"
	"fmt"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p/enr"
)

func TestCleaner(t *testing.T) {
	c := NewCleaner()
	for _, ttl := range []time.Duration{3 * time.Minute, time.Minute, 2 * time.Minute} {
		k, _ := crypto.GenerateKey()
		r := enr.Record{}
		enr.SignV4(&r, k)
		c.Add(int64(ttl), r)
	}
	fmt.Println(c.heap)
	fmt.Println(heap.Pop(&c))
	fmt.Println(c.heap)
}
