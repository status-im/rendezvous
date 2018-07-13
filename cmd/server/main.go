package main

import (
	"fmt"
	"math/rand"

	golog "github.com/ipfs/go-log"
	lcrypto "github.com/libp2p/go-libp2p-crypto"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/spf13/pflag"
	"github.com/status-im/rendezvous/server"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	gologging "github.com/whyrusleeping/go-logging"
)

var (
	port    = pflag.IntP("port", "p", 9090, "listener port")
	address = pflag.StringP("address", "a", "0.0.0.0", "listener ip address")
	data    = pflag.StringP("data", "d", "/tmp/rendevouz", "Path where ENR infos will be stored.")
)

func main() {
	pflag.Parse()
	golog.SetupLogging()
	golog.SetAllLoggers(gologging.INFO)
	priv, _, err := lcrypto.GenerateKeyPairWithReader(lcrypto.Secp256k1, 2048, rand.New(rand.NewSource(int64(*port))))
	must(err)
	laddr, err := ma.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%d", *address, *port))
	must(err)
	db, err := leveldb.OpenFile(*data, &opt.Options{OpenFilesCacheCapacity: 3})
	must(err)
	srv := server.NewServer(laddr, priv, server.NewStorage(db))
	must(srv.Start())
	defer srv.Stop()
	select {}
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
