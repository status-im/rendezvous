package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"

	golog "github.com/ipfs/go-log"
	lcrypto "github.com/libp2p/go-libp2p-crypto"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/spf13/pflag"
	"github.com/status-im/rendezvous/server"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	gologging "github.com/whyrusleeping/go-logging"

	_ "github.com/status-im/go-multiaddr-ethv4"
)

var (
	port     = pflag.IntP("port", "p", 9090, "listener port")
	address  = pflag.StringP("address", "a", "0.0.0.0", "listener ip address")
	data     = pflag.StringP("data", "d", "/tmp/rendevouz", "Path where ENR infos will be stored.")
	generate = pflag.BoolP("generate", "g", false, "dump private key and exit.")
	keypath  = pflag.StringP("keypath", "k", "", "path to load private key")
	keyhex   = pflag.StringP("keyhex", "h", "", "private key hex")
)

func main() {
	pflag.Parse()
	golog.SetupLogging()
	golog.SetAllLoggers(gologging.INFO)
	priv, err := getKey()
	must(err)
	if *generate {
		bytes, err := priv.Bytes()
		must(err)
		fmt.Println(hex.EncodeToString(bytes))
		os.Exit(0)
	}
	laddr, err := ma.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%d", *address, *port))
	must(err)
	db, err := leveldb.OpenFile(*data, &opt.Options{OpenFilesCacheCapacity: 3})
	must(err)
	srv := server.NewServer(laddr, priv, server.NewStorage(db))
	must(srv.Start())
	defer srv.Stop()
	select {}
}

func getKey() (priv lcrypto.PrivKey, err error) {
	var data string
	if len(*keypath) != 0 {
		f, err := os.Open(*keypath)
		if err != nil {
			return priv, err
		}
		defer f.Close()
		hexBytes, err := ioutil.ReadAll(f)
		if err != nil {
			return priv, err
		}
		data = string(hexBytes)
	} else if len(*keyhex) != 0 {
		data = *keyhex
	}
	if len(data) != 0 {
		bytes, err := hex.DecodeString(data)
		if err != nil {
			return priv, err
		}
		return lcrypto.UnmarshalPrivateKey(bytes)
	}
	priv, _, err = lcrypto.GenerateKeyPairWithReader(lcrypto.Secp256k1, 2048, rand.Reader)
	return priv, err
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
