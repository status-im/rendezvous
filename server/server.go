package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/ethereum/go-ethereum/rlp"
	libp2p "github.com/libp2p/go-libp2p"
	crypto "github.com/libp2p/go-libp2p-crypto"
	host "github.com/libp2p/go-libp2p-host"
	net "github.com/libp2p/go-libp2p-net"
	ma "github.com/multiformats/go-multiaddr"
)

func NewServer(laddr ma.Multiaddr, identity crypto.PrivKey) *Server {
	srv := Server{
		laddr:        laddr,
		identity:     identity,
		storage:      NewStorage(),
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

	storage *Storage
	cleaner *Cleaner

	h host.Host
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
	srv.h.SetStreamHandler("/rend/0.1.0", func(s net.Stream) {
		defer s.Close()
		rs := rlp.NewStream(s, 0)
		s.SetReadDeadline(time.Now().Add(srv.readTimeout))
		typ, err := rs.Uint()
		if err != nil {
			log.Printf("error reading message type: %v\n", err)
			return
		}
		s.SetReadDeadline(time.Now().Add(srv.readTimeout))
		resp, err := srv.msgParser(MessageType(typ), rs)
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
	log.Println(srv.laddr.Encapsulate(addr))
	return nil
}

type Decoder interface {
	Decode(val interface{}) error
}

func (srv *Server) msgParser(typ MessageType, d Decoder) (resp interface{}, err error) {
	switch typ {
	case REGISTER:
		var msg Register
		if err = d.Decode(&msg); err != nil {
			return RegisterResponse{Status: E_INVALID_CONTENT}, nil
		}
		srv.storage.Add(msg.Topic, msg.Record)
		return RegisterResponse{Status: OK}, nil
	case UNREGISTER:
		// do we need to allow unregister?
		// it can potentially be abused to remove good nodes from registry
		// alternative is to always remove node only by ttl, so that one who registered
		// a node in control
		var msg Unregister
		if err = d.Decode(&msg); err != nil {
			return RegisterResponse{Status: E_INVALID_CONTENT}, nil
		}
	case DISCOVER:
		var msg Discover
		if err = d.Decode(&msg); err != nil {
			return RegisterResponse{Status: E_INVALID_CONTENT}, nil
		}
		return DiscoverResponse{Records: srv.storage.GetLimit(msg.Topic, msg.Limit)}, nil
	default:
		// don't send the response
		return nil, errors.New("unknown request type")
	}
	return
}
