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
	longestTTL          = 20 * time.Second
	cleanerPeriod       = 2 * time.Second
	maxLimit       uint = 10
	maxTopicLength      = 50
)

// NewServer creates instance of the server.
func NewServer(laddr ma.Multiaddr, identity crypto.PrivKey, s Storage) *Server {
	srv := Server{
		laddr:         laddr,
		identity:      identity,
		storage:       s,
		cleaner:       NewCleaner(),
		writeTimeout:  10 * time.Second,
		readTimeout:   10 * time.Second,
		cleanerPeriod: cleanerPeriod,
	}
	return &srv
}

// Server provides rendezbous service over libp2p stream.
type Server struct {
	laddr    ma.Multiaddr
	identity crypto.PrivKey

	writeTimeout time.Duration
	readTimeout  time.Duration

	storage       Storage
	cleaner       *Cleaner
	cleanerPeriod time.Duration

	h    host.Host
	addr ma.Multiaddr

	wg   sync.WaitGroup
	quit chan struct{}
}

// Addr returns full server multiaddr (identity included).
func (srv *Server) Addr() ma.Multiaddr {
	return srv.addr
}

// Start creates listener.
func (srv *Server) Start() error {
	if err := srv.startListener(); err != nil {
		return err
	}
	return srv.startCleaner()
}

func (srv *Server) startCleaner() error {
	srv.quit = make(chan struct{})
	srv.wg.Add(1)
	go func() {
		for {
			select {
			case <-time.After(srv.cleanerPeriod):
				srv.purgeOutdated()
			case <-srv.quit:
				srv.wg.Done()
				return
			}
		}
	}()
	return nil
}

func (srv *Server) startListener() error {
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
	return nil
}

// Stop closes listener and waits till all helper goroutines are stopped.
func (srv *Server) Stop() {
	if srv.quit == nil {
		return
	}
	select {
	case <-srv.quit:
		return
	default:
	}
	close(srv.quit)
	srv.wg.Wait()
	if srv.h != nil {
		srv.h.Close()
	}
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

// Decoder is a decoder!
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
		return srv.register(msg)
	case protocol.DISCOVER:
		var msg protocol.Discover
		if err = d.Decode(&msg); err != nil {
			return protocol.DiscoverResponse{Status: protocol.E_INVALID_CONTENT}, nil
		}
		limit := msg.Limit
		if msg.Limit > maxLimit {
			limit = maxLimit
		}
		records, err := srv.storage.GetRandom(msg.Topic, limit)
		if err != nil {
			return protocol.DiscoverResponse{Status: protocol.E_INTERNAL_ERROR}, err
		}
		return protocol.DiscoverResponse{Status: protocol.OK, Records: records}, nil
	default:
		// don't send the response
		return nil, errors.New("unknown request type")
	}
}

func (srv *Server) register(msg protocol.Register) (protocol.RegisterResponse, error) {
	if len(msg.Topic) == 0 || len(msg.Topic) > maxTopicLength {
		return protocol.RegisterResponse{Status: protocol.E_INVALID_NAMESPACE}, nil
	}
	if time.Duration(msg.TTL) > longestTTL {
		return protocol.RegisterResponse{Status: protocol.E_INVALID_TTL}, nil
	}
	if !msg.Record.Signed() {
		return protocol.RegisterResponse{Status: protocol.E_INVALID_ENR}, nil
	}
	key, err := srv.storage.Add(msg.Topic, msg.Record)
	if err != nil {
		return protocol.RegisterResponse{Status: protocol.E_INTERNAL_ERROR}, err
	}
	srv.cleaner.Add(time.Now().Add(time.Duration(msg.TTL)), key)
	return protocol.RegisterResponse{Status: protocol.OK}, nil
}
