rendezvous
==========

Implementation of [a rendezvous peers discovery protocol](https://github.com/libp2p/specs/pull/56).

## How to run?

Fetch dependencies

```
dep ensure
```

Start the server.

```
go run ./cmd/server/main.go
```

Copy multiaddr string from logs. In my case /ip4/127.0.0.1/tcp/9090/ipfs/QmYV4YUJ4capN4BmE3D8nKgPaaqzu6T1Vf2NxjTcVg8MVK.

Register ENR using client. Can be executed multiple times, key for ENR is randomly generated.

```
go run ./cmd/client/main.go -s /ip4/127.0.0.1/tcp/9090/ipfs/QmYV4YUJ4capN4BmE3D8nKgPaaqzu6T1Vf2NxjTcVg8MVK -r
```

Fetch discovered ENRs.

```
go run ./cmd/client/main.go -s /ip4/127.0.0.1/tcp/9090/ipfs/QmYV4YUJ4capN4BmE3D8nKgPaaqzu6T1Vf2NxjTcVg8MVK
```
