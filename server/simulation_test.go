package server

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p/enr"
	"github.com/gyuho/goraph"
	"github.com/stretchr/testify/require"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/storage"
	"github.com/ethereum/go-ethereum/p2p/enode"
)

type addr []byte

func (a addr) String() string {
	return hex.EncodeToString(a)
}

func (a addr) ID() goraph.ID {
	return goraph.StringID(a.String())
}

func TestGraphConnected(t *testing.T) {
	for iter := 0; iter < 10; iter++ {
		t.Run(fmt.Sprintf("Iteration/%d", iter), func(t *testing.T) {
			t.Parallel()
			topic := "a"
			n := 100
			k := 5
			memdb, _ := leveldb.Open(storage.NewMemStorage(), nil)
			s := NewStorage(memdb)
			enrs := make([]enr.Record, n)
			for i := 0; i < n; i++ {
				key, _ := crypto.GenerateKey()
				var r enr.Record
				require.NoError(t, enode.SignV4(&r, key))
				_, err := s.Add(topic, r, time.Time{})
				require.NoError(t, err)
				enrs[i] = r
			}
			graph := goraph.NewGraph()
			var last goraph.Node
			for i := range enrs {
				require.True(t, graph.AddNode(addr(enode.ValidSchemes.NodeAddr(&enrs[i]))))
			}
			require.Equal(t, n, graph.GetNodeCount())

			for i := range enrs {
				peers, err := s.GetRandom(topic, uint(k+1)) // in case if self will be returned
				require.NoError(t, err)
				added := 0
				for j := range peers {
					if bytes.Equal(enode.ValidSchemes.NodeAddr(&enrs[i]), enode.ValidSchemes.NodeAddr(&peers[j])) {
						continue
					}
					added++
					require.NoError(t, graph.AddEdge(addr(enode.ValidSchemes.NodeAddr(&enrs[i])).ID(), addr(enode.ValidSchemes.NodeAddr(&peers[j])).ID(), 0))
					if added == k {
						break
					}
				}
				last = addr(enode.ValidSchemes.NodeAddr(&enrs[i]))
			}
			require.Len(t, goraph.BFS(graph, last.ID()), n)

			max := 0
			for idI := range graph.GetNodes() {
				for idJ := range graph.GetNodes() {
					if idI.String() == idJ.String() {
						continue
					}
					hops := gossip(graph, idI, idJ)
					if hops > max {
						max = hops
					}
				}
			}
			require.True(t, max < 5)
		})
	}
}

func gossip(g goraph.Graph, id, target goraph.ID) (rst int) {
	type leveled struct {
		id  goraph.ID
		lvl int
	}
	q := []leveled{leveled{id, 0}}
	visited := make(map[goraph.ID]bool)
	visited[id] = true

	// while Q is not empty:
	for len(q) != 0 {

		u := q[0]
		q = q[1:len(q):len(q)]

		// for each vertex w adjacent to u:
		cmap, _ := g.GetTargets(u.id)
		for _, w := range cmap {
			// if w is not visited yet:
			if _, ok := visited[w.ID()]; !ok {
				q = append(q, leveled{w.ID(), u.lvl + 1}) // Q.push(w)
				visited[w.ID()] = true                    // label w as visited
				if w.ID().String() == target.String() {
					return u.lvl + 1
				}
			}
		}
		pmap, _ := g.GetSources(u.id)
		for _, w := range pmap {
			// if w is not visited yet:
			if _, ok := visited[w.ID()]; !ok {
				q = append(q, leveled{w.ID(), u.lvl + 1})
				visited[w.ID()] = true // label w as visited
				if w.ID().String() == target.String() {
					return u.lvl + 1
				}
			}
		}
	}
	return rst
}
