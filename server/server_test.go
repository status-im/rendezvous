package server

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p/enr"
	"github.com/status-im/rendezvous/protocol"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/storage"
)

func TestCleanOnRegister(t *testing.T) {
	topic := "any"
	type testCase struct {
		desc    string
		cleaned bool
		ttl     time.Duration
		sleep   time.Duration
		records int
		read    int
	}
	for _, tc := range []testCase{
		{"immediate", true, 0, 200 * time.Millisecond, 10, 10},
		{"notimmediate", true, 1 * time.Second, 2 * time.Second, 10, 10},
		{"notcleaned", false, 10 * time.Second, 200 * time.Millisecond, 10, 10},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			memdb, _ := leveldb.Open(storage.NewMemStorage(), nil)
			s := NewStorage(memdb)
			srv := NewServer(nil, nil, s)
			srv.cleanerPeriod = 10 * time.Millisecond
			require.Nil(t, srv.Addr())
			require.NoError(t, srv.startCleaner())
			defer srv.Stop()
			for i := 0; i < tc.records; i++ {
				key, _ := crypto.GenerateKey()
				var r enr.Record
				require.NoError(t, enr.SignV4(&r, key))
				resp, err := srv.register(protocol.Register{Record: r, Topic: topic, TTL: uint64(tc.ttl)})
				require.NoError(t, err)
				require.Equal(t, protocol.OK, resp.Status)
			}
			time.Sleep(tc.sleep)
			records, err := s.GetRandom(topic, uint(tc.read))
			require.NoError(t, err)
			if !tc.cleaned {
				require.Len(t, records, tc.read)
			} else {
				require.Empty(t, records)
			}
		})
	}
}

type regCase struct {
	Desc    string
	Request protocol.Register
	Status  protocol.ResponseStatus
	Err     error
}

func (t regCase) Decode(val interface{}) error {
	if t.Err != nil {
		return t.Err
	}
	reflect.ValueOf(val).Elem().Set(reflect.ValueOf(&t.Request).Elem())
	return nil
}

func TestRegisterRPC(t *testing.T) {
	topic := "any"
	valid := enr.Record{}
	key, _ := crypto.GenerateKey()
	enr.SignV4(&valid, key)
	for _, tc := range []regCase{
		{
			Desc:   "invalid",
			Err:    errors.New("error"),
			Status: protocol.E_INVALID_CONTENT,
		},
		{
			Desc:   "highttl",
			Status: protocol.E_INVALID_TTL,
			Request: protocol.Register{
				Topic: topic,
				TTL:   uint64(longestTTL + 1),
			},
		},
		{
			Desc:   "notopic",
			Status: protocol.E_INVALID_NAMESPACE,
			Request: protocol.Register{
				TTL: uint64(longestTTL - 1),
			},
		},
		{
			Desc:   "longtopic",
			Status: protocol.E_INVALID_NAMESPACE,
			Request: protocol.Register{
				TTL:   uint64(longestTTL - 1),
				Topic: string(make([]byte, maxTopicLength+1)),
			},
		},
		{
			Desc:   "invalidenr",
			Status: protocol.E_INVALID_ENR,
			Request: protocol.Register{
				Topic:  topic,
				TTL:    uint64(longestTTL - 1),
				Record: enr.Record{},
			},
		},
		{
			Desc:   "ok",
			Status: protocol.OK,
			Request: protocol.Register{
				Topic:  topic,
				TTL:    uint64(longestTTL - 1),
				Record: valid,
			},
		},
	} {
		t.Run(tc.Desc, func(t *testing.T) {
			memdb, _ := leveldb.Open(storage.NewMemStorage(), nil)
			s := NewStorage(memdb)
			srv := NewServer(nil, nil, s)
			resptype, resp, err := srv.msgParser(protocol.REGISTER, tc)
			require.NoError(t, err)
			assert.Equal(t, protocol.REGISTER_RESPONSE, resptype)
			assert.Equal(t, tc.Status, resp.(protocol.RegisterResponse).Status)
		})
	}
}

type discCase struct {
	Desc       string
	Request    protocol.Discover
	Status     protocol.ResponseStatus
	RecordsLen int
	Err        error
}

func (t discCase) Decode(val interface{}) error {
	if t.Err != nil {
		return t.Err
	}
	reflect.ValueOf(val).Elem().Set(reflect.ValueOf(&t.Request).Elem())
	return nil
}

func TestDiscoverRPC(t *testing.T) {
	topic := "any"
	for _, tc := range []discCase{
		{
			Desc:   "invalid",
			Err:    errors.New("test"),
			Status: protocol.E_INVALID_CONTENT,
		},
		{
			Desc:       "zerolimit",
			Request:    protocol.Discover{Topic: topic},
			Status:     protocol.OK,
			RecordsLen: 0,
		},
		{
			Desc:       "topicnotexist",
			Request:    protocol.Discover{Topic: "notexist", Limit: 10},
			Status:     protocol.OK,
			RecordsLen: 0,
		},
		{
			Desc:       "highlimit",
			Request:    protocol.Discover{Topic: topic, Limit: maxLimit + 1},
			Status:     protocol.OK,
			RecordsLen: int(maxLimit),
		},
		{
			Desc:       "underlimit",
			Request:    protocol.Discover{Topic: topic, Limit: maxLimit - 1},
			Status:     protocol.OK,
			RecordsLen: int(maxLimit) - 1,
		},
	} {
		t.Run(tc.Desc, func(t *testing.T) {
			memdb, _ := leveldb.Open(storage.NewMemStorage(), nil)
			s := NewStorage(memdb)
			srv := NewServer(nil, nil, s)
			for i := 0; i < 100; i++ {
				key, _ := crypto.GenerateKey()
				var r enr.Record
				require.NoError(t, enr.SignV4(&r, key))
				resp, err := srv.register(protocol.Register{Record: r, Topic: topic})
				require.NoError(t, err)
				require.Equal(t, protocol.OK, resp.Status)
			}

			resptype, resp, err := srv.msgParser(protocol.DISCOVER, tc)
			require.NoError(t, err)
			assert.Equal(t, protocol.DISCOVER_RESPONSE, resptype)
			assert.Equal(t, tc.Status, resp.(protocol.DiscoverResponse).Status)
			assert.Len(t, resp.(protocol.DiscoverResponse).Records, tc.RecordsLen)
		})
	}
}
