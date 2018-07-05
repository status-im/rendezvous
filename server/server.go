package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/rlp"
	libp2p "github.com/libp2p/go-libp2p"
	crypto "github.com/libp2p/go-libp2p-crypto"
	host "github.com/libp2p/go-libp2p-host"
	net "github.com/libp2p/go-libp2p-net"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/status-im/rendezvous/protocol"
)

const (
	longestTTL         = 20 * time.Second
	cleanerPeriod      = 2 * time.Second
	maxLimit      uint = 10
)

func NewServer(laddr ma.Multiaddr, identity crypto.PrivKey, s Storage) *Server {
	srv := Server{
		laddr:        laddr,
		identity:     identity,
		storage:      s,
		cleaner:      NewCleaner(),
		writeTimeout: 10 * time.Second,
		readTimeout:  10 * time.Second,
	}
	return &srv
}

type Server struct {
	laddr    ma.Multiaddr
	identity crypto.PrivKey

	writeTimeout time.Duration
	readTimeout  time.Duration

	storage Storage
	cleaner *Cleaner

	h    host.Host
	addr ma.Multiaddr

	wg   sync.WaitGroup
	quit chan struct{}
}

func (srv *Server) Addr() ma.Multiaddr {
	return srv.addr
}

func (srv *Server) Start() error {
	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(srv.laddr.String()),
		libp2p.Identity(srv.identity),
	}
	h, err := libp2p.New(context.Background(), opts...)
	if err != nil {
		return err
	}
	srv.h = h
	srv.h.SetStreamHandler(protocol.VERSION, func(s net.Stream) {
		defer s.Close()
		rs := rlp.NewStream(s, 0)
		s.SetReadDeadline(time.Now().Add(srv.readTimeout))
		typ, err := rs.Uint()
		if err != nil {
			log.Printf("error reading message type: %v\n", err)
			return
		}
		s.SetReadDeadline(time.Now().Add(srv.readTimeout))
		resp, err := srv.msgParser(protocol.MessageType(typ), rs)
		if err != nil {
			log.Printf("error parsing message: %v\n", err)
			return
		}
		s.SetWriteDeadline(time.Now().Add(srv.writeTimeout))
		if err = rlp.Encode(s, resp); err != nil {
			log.Printf("error encoding response %v : %v\n", resp, err)
		}
	})
	addr, err := ma.NewMultiaddr(fmt.Sprintf("/ipfs/%s", h.ID().Pretty()))
	if err != nil {
		return err
	}
	srv.addr = srv.laddr.Encapsulate(addr)
	log.Println(srv.laddr.Encapsulate(addr))
	srv.quit = make(chan struct{})
	srv.wg.Add(1)
	go func() {
		for {
			select {
			case <-time.After(cleanerPeriod):
				srv.purgeOutdated()
			case <-srv.quit:
				srv.wg.Done()
				return
			}
		}
	}()
	return nil
}

func (srv *Server) Stop() {
	close(srv.quit)
	srv.wg.Wait()
	srv.h.Close()
}

func (srv *Server) purgeOutdated() {
	key := srv.cleaner.PopOneSince(time.Now())
	if len(key) == 0 {
		return
	}
	if err := srv.storage.RemoveByKey(key); err != nil {
		log.Printf("error removing key '%s' from storage: %v", key, err)
	}
}

type Decoder interface {
	Decode(val interface{}) error
}

func (srv *Server) msgParser(typ protocol.MessageType, d Decoder) (resp interface{}, err error) {
	switch typ {
	case protocol.REGISTER:
		var msg protocol.Register
		if err = d.Decode(&msg); err != nil {
			return protocol.RegisterResponse{Status: protocol.E_INVALID_CONTENT}, nil
		}
		if time.Duration(msg.TTL) > longestTTL {
			return protocol.RegisterResponse{Status: protocol.E_INVALID_TTL}, nil
		}
		key, err := srv.storage.Add(msg.Topic, msg.Record)
		if err != nil {
			return protocol.RegisterResponse{Status: protocol.E_INTERNAL_ERROR}, err
		}
		srv.cleaner.Add(time.Now().Add(time.Duration(msg.TTL)), key)
		return protocol.RegisterResponse{Status: protocol.OK}, nil
	case protocol.DISCOVER:
		var msg protocol.Discover
		if err = d.Decode(&msg); err != nil {
			return protocol.RegisterResponse{Status: protocol.E_INVALID_CONTENT}, nil
		}
		if msg.Limit > maxLimit {
			return protocol.RegisterResponse{Status: protocol.E_INVALID_LIMIT}, nil
		}
		records, err := srv.storage.GetRandom(msg.Topic, msg.Limit)
		if err != nil {
			return protocol.RegisterResponse{Status: protocol.E_INTERNAL_ERROR}, err
		}
		return protocol.DiscoverResponse{Status: protocol.OK, Records: records}, nil
	default:
		// don't send the response
		return nil, errors.New("unknown request type")
	}
}
