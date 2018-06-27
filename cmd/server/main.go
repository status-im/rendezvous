package main

import (
	"fmt"
	"math/rand"

	golog "github.com/ipfs/go-log"
	lcrypto "github.com/libp2p/go-libp2p-crypto"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/spf13/pflag"
	"github.com/status-im/rendezvous/server"
	gologging "github.com/whyrusleeping/go-logging"
)

var (
	port    = pflag.IntP("port", "p", 9090, "listener port")
	address = pflag.StringP("address", "a", "0.0.0.0", "listener ip address")
)

func main() {
	pflag.Parse()
	golog.SetupLogging()
	golog.SetAllLoggers(gologging.INFO)
	priv, _, err := lcrypto.GenerateKeyPairWithReader(lcrypto.RSA, 2048, rand.New(rand.NewSource(int64(*port))))
	must(err)
	laddr, err := ma.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%d", *address, *port))
	must(err)
	srv := server.NewServer(laddr, priv)
	must(srv.Start())
	select {}
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
