package e2e

import (
	"context"
	"crypto/ecdsa"
	"math/rand"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/p2p/enr"
	golog "github.com/ipfs/go-log"
	lcrypto "github.com/libp2p/go-libp2p-core/crypto"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/status-im/rendezvous"
	"github.com/status-im/rendezvous/server"
	"github.com/stretchr/testify/require"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/storage"
)

func TestClientRegisterDiscover(t *testing.T) {

	golog.SetupLogging()
	lvl, _ := golog.LevelFromString("info")
	golog.SetAllLoggers(lvl)

	priv, _, err := lcrypto.GenerateKeyPairWithReader(lcrypto.Secp256k1, 2048, rand.New(rand.NewSource(1)))
	require.NoError(t, err)
	laddr, err := ma.NewMultiaddr("/ip4/127.0.0.1/tcp/7777")
	require.NoError(t, err)
	db, err := leveldb.Open(storage.NewMemStorage(), nil)
	require.NoError(t, err)
	srv := server.NewServer(laddr, priv, server.NewStorage(db))
	require.NoError(t, srv.Start())

	priv, _, err = lcrypto.GenerateKeyPairWithReader(lcrypto.Secp256k1, 2048, rand.New(rand.NewSource(2)))
	require.NoError(t, err)
	client, err := rendezvous.NewEphemeral()
	require.NoError(t, err)

	k, _ := crypto.GenerateKey()
	record := enr.Record{}
	record.Set(enr.IP{10, 0, 10, 24})
	record.Set(enr.TCP(8087))
	record.Set(enr.WithEntry("nonce", uint(1010)))
	require.NoError(t, enode.SignV4(&record, k))
	require.NoError(t, client.Register(context.TODO(), srv.Addr(), "any", record, 5*time.Second))
	records, err := client.Discover(context.TODO(), srv.Addr(), "any", 1)
	require.NoError(t, err)
	require.Len(t, records, 1)
	var id enode.Secp256k1
	require.NoError(t, records[0].Load(&id))
	require.Equal(t, crypto.PubkeyToAddress(k.PublicKey), crypto.PubkeyToAddress(ecdsa.PublicKey(id)))
}
