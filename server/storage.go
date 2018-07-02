package server

import (
	"crypto/rand"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/p2p/enr"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/syndtr/goleveldb/leveldb"
)

func NewStorage(db *leveldb.DB) Storage {
	return Storage{db: db}
}

type Storage struct {
	db *leveldb.DB
}

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

func (s *Storage) RemoveByKey(key string) error {
	return s.db.Delete([]byte(key), nil)
}

func (s *Storage) GetRandom(topic string, limit uint) (rst []enr.Record, err error) {
	iter := s.db.NewIterator(nil, nil)
	defer iter.Release()
	tries := uint(0)
	found := map[string]struct{}{}
	for tries < limit*3 {
		tries++
		id := make([]byte, 32)
		key := []byte{}
		key = append(key, []byte(topic)...)
		if _, err = rand.Read(id[:]); err != nil {
			return
		}
		key = append(key, id...)
		iter.Seek(key)
		if iter.Next() {
			valkey := common.Bytes2Hex(iter.Key())
			if _, exist := found[valkey]; exist {
				continue
			}
			found[valkey] = struct{}{}
			var record enr.Record
			if err = rlp.DecodeBytes(iter.Value(), &record); err != nil {
				return
			}
			rst = append(rst, record)
			if uint(len(rst)) == limit {
				return
			}
		}
	}
	return
}
