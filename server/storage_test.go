package server

import (
	"crypto/ecdsa"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p/enr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/storage"
)

func TestGetRandom(t *testing.T) {
	type testCase struct {
		desc   string
		total  int
		get    int
		should int
	}
	for _, tc := range []testCase{
		{"multirand", 100, 5, 5},
		{"single", 1, 1, 1},
		{"noentries", 0, 0, 0},
		{"samerand", 10, 10, 10},
		{"morethanpool", 5, 10, 5},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()
			memdb, _ := leveldb.Open(storage.NewMemStorage(), nil)
			s := NewStorage(memdb)
			for i := 0; i < tc.total; i++ {
				key, _ := crypto.GenerateKey()
				var r enr.Record
				require.NoError(t, enr.SignV4(&r, key))
				_, err := s.Add("some", r)
				require.NoError(t, err)
			}
			records, err := s.GetRandom("some", uint(tc.get))
			require.NoError(t, err)
			require.Len(t, records, tc.should)
			more, err := s.GetRandom("some", uint(tc.get))
			require.Equal(t, len(records), len(more))
		})
	}
}

func TestGetRandomMultipleTimes(t *testing.T) {
	memdb, _ := leveldb.Open(storage.NewMemStorage(), nil)
	s := NewStorage(memdb)
	for i := 0; i < 100; i++ {
		key, _ := crypto.GenerateKey()
		var r enr.Record
		require.NoError(t, enr.SignV4(&r, key))
		_, err := s.Add("some", r)
		require.NoError(t, err)
	}
	hits := map[common.Address]int{}
	for i := 0; i < 10; i++ {
		records, err := s.GetRandom("some", 1)
		require.NoError(t, err)
		require.Len(t, records, 1)
		var id enr.Secp256k1
		require.NoError(t, records[0].Load(&id))
		hits[crypto.PubkeyToAddress(ecdsa.PublicKey(id))]++
	}
	for _, hit := range hits {
		assert.True(t, hit < 3)
	}
}
