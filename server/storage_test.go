package server

import (
	"crypto/ecdsa"
	"strconv"
	"testing"
	"time"

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
			memdb, _ := leveldb.Open(storage.NewMemStorage(), nil)
			s := NewStorage(memdb)
			for i := 0; i < tc.total; i++ {
				key, _ := crypto.GenerateKey()
				var r enr.Record
				require.NoError(t, enr.SignV4(&r, key))
				_, err := s.Add("some", r, time.Time{})
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
		_, err := s.Add("some", r, time.Time{})
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

func TestGetRandomMultiTopics(t *testing.T) {
	first := "first"
	second := "second"
	firstSet := map[string]struct{}{}
	secondSet := map[string]struct{}{}
	memdb, _ := leveldb.Open(storage.NewMemStorage(), nil)
	s := NewStorage(memdb)
	for _, fixture := range []struct {
		topic string
		set   map[string]struct{}
	}{
		{first, firstSet},
		{second, secondSet},
	} {
		for i := 0; i < 5; i++ {
			key, _ := crypto.GenerateKey()
			var r enr.Record
			require.NoError(t, enr.SignV4(&r, key))
			_, err := s.Add(fixture.topic, r, time.Time{})
			require.NoError(t, err)
			fixture.set[crypto.PubkeyToAddress(key.PublicKey).Hex()] = struct{}{}
		}
	}
	firstRecods, err := s.GetRandom(first, 5)
	require.NoError(t, err)
	require.Len(t, firstRecods, 5)
	for _, r := range firstRecods {
		var id enr.Secp256k1
		require.NoError(t, r.Load(&id))
		addr := crypto.PubkeyToAddress(ecdsa.PublicKey(id))
		assert.Contains(t, firstSet, addr.Hex())
		assert.NotContains(t, secondSet, addr.Hex())
	}
	secondRecods, err := s.GetRandom(second, 5)
	require.NoError(t, err)
	require.Len(t, secondRecods, 5)
	for _, r := range secondRecods {
		var id enr.Secp256k1
		require.NoError(t, r.Load(&id))
		addr := crypto.PubkeyToAddress(ecdsa.PublicKey(id))
		assert.Contains(t, secondSet, addr.Hex())
		assert.NotContains(t, firstSet, addr.Hex())
	}
}

func TestIterateKeys(t *testing.T) {
	topic := "a"
	count := 5
	keys := map[string]struct{}{}
	memdb, _ := leveldb.Open(storage.NewMemStorage(), nil)
	s := NewStorage(memdb)
	for i := 0; i < count; i++ {
		pkey, _ := crypto.GenerateKey()
		var r enr.Record
		require.NoError(t, enr.SignV4(&r, pkey))
		key, err := s.Add(topic, r, time.Time{})
		require.NoError(t, err)
		keys[key] = struct{}{}
	}
	rstcount := 0
	require.NoError(t, s.IterateAllKeys(func(key RecordsKey, ttl time.Time) error {
		rstcount++
		assert.Contains(t, keys, key.String())
		return nil
	}))
	require.Equal(t, count, rstcount)
}

func BenchmarkRandomReads(b *testing.B) {
	for _, n := range []int{0, 3, 10, 20, 50, 100, 1000} {
		b.Run(strconv.Itoa(n), func(b *testing.B) {
			benchmarkRandomReads(b, n)
		})
	}
}

func benchmarkRandomReads(b *testing.B, records int) {
	topic := "a"
	memdb, _ := leveldb.Open(storage.NewMemStorage(), nil)
	s := NewStorage(memdb)
	for i := 0; i < records; i++ {
		key, _ := crypto.GenerateKey()
		var r enr.Record
		require.NoError(b, enr.SignV4(&r, key))
		_, err := s.Add(topic, r, time.Time{})
		require.NoError(b, err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.GetRandom(topic, 5)
	}
}
