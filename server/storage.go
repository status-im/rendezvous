package server

import (
	"sync"

	"github.com/ethereum/go-ethereum/p2p/enr"
)

func NewStorage() *Storage {
	s := new(Storage)
	s.db = map[string]map[string]enr.Record{}
	return s
}

type Storage struct {
	mu sync.RWMutex
	db map[string]map[string]enr.Record
}

func (s *Storage) Add(topic string, record enr.Record) {
	s.mu.Lock()
	if _, ok := s.db[topic]; !ok {
		s.db[topic] = map[string]enr.Record{}
	}
	s.db[topic][string(record.NodeAddr())] = record
	s.mu.Unlock()
}

func (s *Storage) Remove(topic string, record enr.Record) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.db[topic]; !ok {
		return
	}
	delete(s.db[topic], string(record.NodeAddr()))
}

func (s *Storage) GetLimit(topic string, limit uint) (rst []enr.Record) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var i uint
	for _, record := range s.db[topic] {
		// copy?
		rst = append(rst, record)
		i++
		if i == limit {
			return
		}
	}
	return
}
