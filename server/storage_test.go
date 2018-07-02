package server

import (
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p/enr"
	"github.com/stretchr/testify/require"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/storage"
)

func TestGetRandom(t *testing.T) {
	memdb, _ := leveldb.Open(storage.NewMemStorage(), nil)
	s := NewStorage(memdb)
	for i := 0; i < 20; i++ {
		key, _ := crypto.GenerateKey()
		var r enr.Record
		require.NoError(t, enr.SignV4(&r, key))
		_, err := s.Add("some", r)
		require.NoError(t, err)
	}
	records, err := s.GetRandom("some", 2)
	require.NoError(t, err)
	require.Len(t, records, 2)
}
