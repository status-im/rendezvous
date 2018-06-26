package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p/enr"
	golog "github.com/ipfs/go-log"
	lcrypto "github.com/libp2p/go-libp2p-crypto"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/spf13/pflag"
	"github.com/status-im/rendezvous/client"
	gologging "github.com/whyrusleeping/go-logging"
)

var (
	port   = pflag.IntP("port", "p", 9091, "")
	server = pflag.StringP("server", "s", "", "rendevouz server")
	reg    = pflag.BoolP("reg", "r", false, "")
)

func main() {
	pflag.Parse()
	golog.SetupLogging()
	golog.SetAllLoggers(gologging.INFO)
	priv, _, err := lcrypto.GenerateKeyPairWithReader(lcrypto.RSA, 2048, rand.New(rand.NewSource(int64(*port))))
	must(err)
	laddr, err := ma.NewMultiaddr(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", *port))
	must(err)
	srv, err := ma.NewMultiaddr(*server)
	must(err)
	client, err := client.New(laddr, priv)
	must(err)
	if *reg {
		k, _ := crypto.GenerateKey()
		record := enr.Record{}
		record.Set(enr.IP{10, 0, 10, 24})
		record.Set(enr.TCP(8087))
		must(enr.SignV4(&record, k))
		must(client.Register(context.TODO(), srv, "topic", record))
	} else {
		records, err := client.Discover(context.TODO(), srv, "topic", 5)
		must(err)
		for _, r := range records {
			var (
				ip   enr.IP
				port enr.TCP
			)
			must(r.Load(&ip))
			must(r.Load(&port))
			log.Printf("loaded enr with address %v:%v", ip, port)
		}
	}
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
