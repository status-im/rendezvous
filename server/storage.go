package server

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"

	"github.com/ethereum/go-ethereum/p2p/enr"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

const (
	maxRandomPool = 50
)

// NewStorage creates instance of the storage.
func NewStorage(db *leveldb.DB) Storage {
	return Storage{db: db}
}

// Storage manages records.
type Storage struct {
	db *leveldb.DB
}

// Add stores record using specified topic.
func (s Storage) Add(topic string, record enr.Record) (string, error) {
	key := make([]byte, 0, len([]byte(topic))+len(record.NodeAddr()))
	key = append(key, []byte(topic)...)
	key = append(key, record.NodeAddr()...)
	data, err := rlp.EncodeToBytes(record)
	if err != nil {
		return "", err
	}
	return string(key), s.db.Put(key, data, nil)
}

// RemoveBykey removes record from storage.
func (s *Storage) RemoveByKey(key string) error {
	return s.db.Delete([]byte(key), nil)
}

// GetRandom reads random records for specified topic up to specified limit.
func (s *Storage) GetRandom(topic string, limit uint) (rst []enr.Record, err error) {
	iter := s.db.NewIterator(util.BytesPrefix([]byte(topic)), nil)
	defer iter.Release()
	tlth := len([]byte(topic))
	key := make([]byte, tlth+32) // doesn't have to be precisely original length of the key
	hexes := map[string]struct{}{}
	// it might be too much cause we do crypto/rand.Read. requires profiling
	for i := uint(0); i < limit*limit && len(rst) < int(limit); i++ {
		if _, err := rand.Read(key); err != nil {
			return nil, err
		}
		copy(key, []byte(topic))
		iter.Seek(key)
		for _, f := range []func() bool{iter.Prev, iter.Next} {
			if f() && bytes.Equal([]byte(topic), iter.Key()[:tlth]) {
				var record enr.Record
				if err = rlp.DecodeBytes(iter.Value(), &record); err != nil {
					return nil, err
				}
				h := hex.EncodeToString(iter.Key())
				if _, exist := hexes[h]; exist {
					continue
				}
				hexes[h] = struct{}{}
				rst = append(rst, record)
				break
			}
		}
	}
	return rst, nil
}
