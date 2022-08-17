module github.com/status-im/rendezvous

replace github.com/ethereum/go-ethereum v1.10.21 => github.com/status-im/go-ethereum v1.10.4-status.4

go 1.15

require (
	github.com/btcsuite/btcd/btcec/v2 v2.2.0
	github.com/ethereum/go-ethereum v1.10.21
	github.com/gyuho/goraph v0.0.0-20171001060514-a7a4454fd3eb
	github.com/ipfs/go-log v1.0.5
	github.com/libp2p/go-libp2p v0.21.0
	github.com/libp2p/go-libp2p-core v0.19.1
	github.com/multiformats/go-multiaddr v0.6.0
	github.com/prometheus/client_golang v1.12.1
	github.com/spf13/pflag v1.0.3
	github.com/status-im/go-multiaddr-ethv4 v1.2.3
	github.com/stretchr/testify v1.8.0
	github.com/syndtr/goleveldb v1.0.1-0.20220614013038-64ee5596c38a
)
